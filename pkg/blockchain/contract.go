package blockchain

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"metachain/pkg/block"
	"metachain/pkg/contract/evm"
	"metachain/pkg/logger"
	"metachain/pkg/storage/store"

	"metachain/pkg/transaction"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

//binding address and eth address
func (bc *Blockchain) bindingAddress(DBTransaction store.Transaction, ethaddr, metaaddr []byte) error {
	return DBTransaction.Mset(BindingKey, ethaddr, metaaddr)
}

//get bingding meta address by eth address
func (bc *Blockchain) GetBindingMetaAddress(ethAddr string) (*common.Address, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()
	return getBindingMetaAddress(DBTransaction, ethAddr)
}

func getBindingMetaAddress(DBTransaction store.Transaction, ethAddr string) (*common.Address, error) {
	data, err := DBTransaction.Mget(BindingKey, common.HexToAddress(ethAddr).Bytes())
	if err != nil {
		if err.Error() != "NotExist" {
			logger.Error("GetBindingMetaAddress error", zap.String("ethaddr", ethAddr), zap.Error(err))
		}
		return nil, err
	}
	addr := common.BytesToAddress(data)
	return &addr, nil
}

//get bingding eth address by Meta address
func (bc *Blockchain) GetBindingEthAddress(MetaAddr *common.Address) (string, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()

	ethAddr, err := DBTransaction.Mget(BindingKey, MetaAddr.Bytes())
	if err != nil {
		return "", err
	}

	return common.BytesToAddress(ethAddr).String(), nil
}

//call contract with out transaction
func (bc *Blockchain) CallSmartContract(contractAddr, origin, callInput, value string) (string, string, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))

	snapshotId := bc.evm.GetSnapshot()
	defer func() {
		bc.evm.RevertToSnapshot(snapshotId)
	}()

	{
		//set block into into evm
		maxH, err := bc.getMaxBlockHeight()
		if err != nil {
			logger.Error("Failed to getMaxBlockHeight", zap.Error(err))
			return "", "", err
		}
		if maxH > 0 {
			currB, err := bc.getBlockByHeight(maxH)
			if err != nil {
				logger.Error("Failed to getBlockByHeight", zap.Error(err))
				return "", "", err
			}
			miner := *currB.Miner
			bc.evm.SetBlockInfo(currB.Height, currB.Timestamp, miner, currB.GlobalDifficulty)
		}
	}

	vl := big.NewInt(0)
	if len(value) > 0 {
		if res, ok := big.NewInt(0).SetString(value, 16); ok {
			vl = res.Div(res, Uint64ToBigInt(ETHDECIMAL))
		}
	}

	bc.evm.SetConfig(vl, new(big.Int).SetUint64(bc.ChainCfg.GasPrice), MAXGASLIMIT, common.HexToAddress(origin))

	if len(Check0x(callInput)) > 0 && len(contractAddr) > 0 {
		ret, gasleft, err := bc.evm.Call(common.HexToAddress(contractAddr), common.HexToAddress(origin), common.Hex2Bytes(Check0x(callInput)))
		if err != nil {
			logger.Info("call contract", zap.String("ret", common.Bytes2Hex(ret)), zap.Error(err))
			return common.Bytes2Hex(ret), "", err
		}
		gasUsed := MAXGASLIMIT - gasleft
		if gasUsed < CONTRACTMINGASLIMIT {
			gasUsed = CONTRACTMINGASLIMIT
		}
		return common.Bytes2Hex(ret), fmt.Sprintf("%X", gasUsed), nil
	}

	logger.Error("failed to Call", zap.Error(fmt.Errorf("wrong input[%v] or contractaddr[%v]", callInput, contractAddr)))
	return "", "", fmt.Errorf("wrong input[%v] or contractaddr[%v]", callInput, contractAddr)
}

//get contract code by address
func (bc *Blockchain) GetCode(contractAddr string) []byte {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.sdb.GetCode(common.HexToAddress(contractAddr))
}

//get contract code by address
func (bc *Blockchain) SetCode(contractAddr common.Address, code []byte) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.sdb.SetCode(contractAddr, code)
	return
}

//get evm logs
func (bc *Blockchain) GetLogs() []*evmtypes.Log {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.evm.Logs()
}

//get evm logs
func (bc *Blockchain) GetLogByHeight(height uint64) []*evmtypes.Log {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var evmlog []*evmtypes.Log
	block, err := bc.getBlockByHeight(height)
	if err != nil {
		logger.Error("getBlockByHeight err", zap.Error(err))
		return nil
	}

	for _, tx := range block.Transactions {
		if len(tx.Input) == 0 {
			continue
		}
		evmtx, err := transaction.DecodeEvmData(tx.Input)
		if err != nil {
			fmt.Println("DecodeEvmData err", err)
			return nil
		}

		for _, log := range evmtx.Logs {
			evmlog = append(evmlog, log)
		}

	}

	return evmlog
}

//handle contract transaction
func (bc *Blockchain) handleContractTransaction(block *block.Block, DBTransaction store.Transaction, tx *transaction.FinishedTransaction, index int) (*big.Int, error) {
	var gasLeft, gasLimit = new(big.Int), new(big.Int)
	// gasLimit = tx.Transaction.GasLimit * tx.Transaction.GasPrice

	gasLimit.Mul(tx.GasLimit, tx.GasPrice)
	// if gasLimit == 0 {
	if gasLimit.Cmp(big.NewInt(0)) == 0 {
		// gasLimit = MAXGASLIMIT
		gasLimit = Limit.MaxGasLimit()
	}
	evmC, err := transaction.DecodeEvmData(tx.Input)
	if err != nil {
		logger.Error("DecodeEvmData input error:", zap.Error(err))
		return gasLeft, err
	}

	eth_tx, err := transaction.DecodeEthData(evmC.EthData)
	if err != nil {
		logger.Error("DecodeEvmData input error:", zap.Error(err))
		return gasLeft, err
	}

	logger.Info("handleContractTransaction info", zap.Uint64("eth_tx gasprice", eth_tx.GasPrice().Uint64()), zap.String("gaslimit", gasLimit.String()), zap.String("tx.gasprice", tx.GasPrice.String()),
		zap.String("origin", evmC.Origin.Hex()), zap.Uint64("nonce", tx.Transaction.Nonce), zap.Uint64("eth_tx.Value()", eth_tx.Value().Uint64()))

	TxValue := eth_tx.Value().Div(eth_tx.Value(), Uint64ToBigInt(ETHDECIMAL))
	bc.evm.SetConfig(TxValue, tx.GasPrice, gasLimit.Uint64(), evmC.Origin)
	bc.evm.Prepare(common.BytesToHash(tx.Hash()), common.BytesToHash(block.Hash), index)

	defer func() {
		input, err := transaction.EncodeEvmData(evmC)
		if err != nil {
			logger.Error("EncodeEvmData error", zap.Error(err))
			return
		}
		tx.Input = input
	}()

	switch evmC.Operation {
	case "create", "Create":
		bc.evm.SetNonce(evmC.Origin, tx.Transaction.Nonce)
		ret, contractAddr, left, err := bc.evm.Create(evmC.CreateCode, evmC.Origin)

		// gasLeft = left
		gasLeft = Uint64ToBigInt(left)
		// if gasLeft < MINGASLIMIT {
		// 	gasLeft = MINGASLIMIT
		// }
		if gasLeft.Cmp(Limit.MinGasLimit()) == -1 {
			gasLeft = Limit.MinGasLimit()
		}

		if err != nil {
			logger.Error("faile to create contract", zap.Error(fmt.Errorf("hash:%v,gasLeft:%v,error:%v", hex.EncodeToString(tx.Hash()), gasLeft, err))) //zap.String("name", evmC.ContractName), zap.Uint64("gasLeft", gasLeft), zap.Error(callErr))
			evmC.Ret = common.Bytes2Hex(ret)
			evmC.Status = false
			//callErr = err
			return gasLeft, nil
		}
		evmC.ContractAddr = contractAddr
		logger.Info("Create contract successfully", zap.String("contract address:", evmC.ContractAddr.Hex()), zap.String("gasLeft", gasLeft.String()), zap.String("ret:", evmC.Ret), zap.String("origin", evmC.Origin.Hex()))
	case "call", "Call":
		bc.evm.SetNonce(evmC.Origin, tx.Transaction.Nonce+1)
		ret, left, err := bc.evm.Call(evmC.ContractAddr, evmC.Origin, evmC.CallInput)

		// gasLeft = left
		gasLeft = Uint64ToBigInt(left)
		// if gasLeft < MINGASLIMIT {
		// 	gasLeft = MINGASLIMIT
		// }
		if gasLeft.Cmp(Limit.MinGasLimit()) == -1 {
			gasLeft = Limit.MinGasLimit()
		}

		if err != nil {
			logger.Error("faile to Call", zap.String("hash", hex.EncodeToString(tx.Hash())), zap.String("gasLeft", gasLeft.String()), zap.Error(err))
			evmC.Ret = common.Bytes2Hex(ret)
			evmC.Status = false
			//callErr = err
			return gasLeft, nil
		}
		evmC.Ret = common.Bytes2Hex(ret)
		logger.Info("Call contract successfully", zap.String("contract address:", evmC.ContractAddr.Hex()), zap.String("gasLeft", gasLeft.String()), zap.String("ret:", evmC.Ret), zap.String("origin", evmC.Origin.Hex()))
	}
	evmC.Status = true
	evmC.Logs = bc.evm.GetLogs(common.BytesToHash(tx.Hash()), common.BytesToHash(block.Hash))

	return gasLeft, nil
}

//get storage at address
func (bc *Blockchain) GetStorageAt(addr, hash string) common.Hash {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.evm.GetStorageAt(common.HexToAddress(addr), common.HexToHash(hash))
}
