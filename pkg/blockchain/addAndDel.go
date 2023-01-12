package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"

	"metachain/pkg/block"
	"metachain/pkg/contract/evm"
	"metachain/pkg/logger"
	"metachain/pkg/storage/miscellaneous"
	"metachain/pkg/storage/store"
	"metachain/pkg/storage/store/bg/bgdb"
	"metachain/pkg/transaction"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"go.uber.org/zap"
)

func (bc *Blockchain) NewTransaction() store.Transaction {
	return bc.db.NewTransaction()
}

// AddBlock add blocks to blockchain
func (bc *Blockchain) AddTempBlock(block *block.Block, DBTransaction store.Transaction) error {
	logger.Debug("AddTempBlock", zap.Uint64("blockHeight", block.Height), zap.String("hash", hex.EncodeToString(block.Hash)))

	var err error
	var height, prevHeight uint64
	if err != nil {
		logger.Error("AddBlock Failed, getPledgeReleaseInfo", zap.Error(err))
		return err
	}

	if block.Height != 1 {
		SnapRoothash, err := DBTransaction.Get(SnapRootKey)
		if err != nil {

			return err
		}
		startHash, err := factCommit(bc.sdb, true)
		if err != nil {
			logger.Error("Failed to set factCommit", zap.Error(err))
			return err
		}

		if !bytes.Equal(SnapRoothash, startHash.Bytes()) {
			logger.Error("snaproot  is changed", zap.String("SnapRootkey-hash=[", hex.EncodeToString(SnapRoothash)),
				zap.String("],nowsnaproothash=[", hex.EncodeToString(startHash.Bytes())))
			//REVERT = fmt.Errorf("snaproot not equal,old root hash:%v,current root hash:%v", hex.EncodeToString(SnapRoothash), startHash)
			return err
		}
	}
	prevHeight, err = getMaxBlockHeight(DBTransaction)
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		return err
	}

	/* 	{
	   		fileName := fmt.Sprintf("AddTempBlock_%v", prevHeight)
	   		readStatedbDate(sdb, fileName)
	   	}
	*/
	height = prevHeight + 1
	if block.Height != height {
		return fmt.Errorf("height error:current height=%d,commit height=%d", prevHeight, block.Height)
	}

	// height -> hash
	hash := block.Hash
	if err = DBTransaction.Set(append(HeightPrefix, miscellaneous.E64func(height)...), hash); err != nil {
		logger.Error("Failed to set height and hash", zap.Error(err))
		return err
	}

	// reset block height
	DBTransaction.Del(HeightKey)
	DBTransaction.Set(HeightKey, miscellaneous.E64func(height))

	//set block into into evm
	{
		bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))
		miner := *block.Miner
		bc.evm.SetBlockInfo(block.Height, block.Timestamp, miner, block.GlobalDifficulty)
	}

	var blockGasU = new(big.Int)
	for index, tx := range block.Transactions {
		if tx.Transaction.IsCoinBaseTransaction() {
			txHash := tx.Hash()
			if err = setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("amount", tx.Amount.String()))
				return err
			}

			//	gas := tx.Transaction.GasLimit * tx.Transaction.GasPrice
			// tx.GasUsed = tx.Transaction.GasLimit * tx.Transaction.GasPrice
			tx.GasUsed = new(big.Int).Mul(tx.GasLimit, tx.GasPrice)
			// blockGasU += tx.GasUsed
			blockGasU.Add(blockGasU, tx.GasUsed)
			if err = setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("from address", block.Miner.String()), zap.String("fee", tx.GasUsed.String()))
				return err
			}

			if err := bc.setToAccount(block, &tx.Transaction); err != nil {
				logger.Error("Failed to set account", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("amount", tx.Amount.String()))
				return err
			}
			// } else if tx.Transaction.IsLockTransaction() || tx.Transaction.IsUnlockTransaction() {
			// 	txHash := tx.Hash()
			// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
			// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.Uint64("amount", tx.Transaction.Amount))
			// 		return err
			// 	}

			// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, height, uint64(index)); err != nil {
			// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.Uint64("amount", tx.Transaction.Amount))
			// 		return err
			// 	}

			// 	nonce := tx.Transaction.Nonce + 1
			// 	if err := setNonce(bc.sdb, tx.Transaction.From, nonce); err != nil {
			// 		logger.Error("Failed to set nonce", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.Uint64("amount", tx.Transaction.Amount))
			// 		return err
			// 	}

			// 	tx.GasUsed = tx.Transaction.GasLimit * tx.Transaction.GasPrice
			// 	blockGasU += tx.GasUsed
			// 	if err = setMinerFee(bc, block.Miner, tx.GasUsed); err != nil {
			// 		logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("from address", block.Miner.String()), zap.Uint64("fee", tx.GasUsed))
			// 		return err
			// 	}

			// 	if tx.Transaction.IsLockTransaction() {
			// 		if err := setFreezeAccount(bc, tx.Transaction.From, tx.Transaction.To, tx.Transaction.Amount, tx.GasUsed, 1); err != nil {
			// 			logger.Error("Faile to setFreezeAccount", zap.String("address", tx.Transaction.From.String()),
			// 				zap.Uint64("amount", tx.Transaction.Amount))
			// 			return err
			// 		}
			// 	} else {
			// 		if err := setFreezeAccount(bc, tx.Transaction.From, tx.Transaction.To, tx.Transaction.Amount, tx.GasUsed, 0); err != nil {
			// 			logger.Error("Faile to setFreezeAccount", zap.String("address", tx.Transaction.From.String()),
			// 				zap.Uint64("amount", tx.Transaction.Amount))
			// 			return err
			// 		}
			// 	}
			// } else if tx.Transaction.IsPledgeTrasnaction() || tx.IsPledgeBreakTransaction() {
			// 	logger.SugarLogger.Debug(">>>>>>>>>>>>>>>>>>>>>>Start IsPledgeTrasnaction")
			// 	txHash := tx.Hash()
			// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
			// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
			// 		return err
			// 	}
			// 	nonce := tx.Transaction.Nonce + 1
			// 	if err := setNonce(bc.sdb, tx.Transaction.From, nonce); err != nil {
			// 		logger.Error("Failed to set nonce", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
			// 		return err
			// 	}
			// 	{
			// 		bfp, _ := getTotalPledge(bc.sdb, tx.From)
			// 		logger.SugarLogger.Debugf("Before pledge from[%v] total pledge[%v]\n", tx.From, bfp)
			// 	}
			// 	var destroy uint64
			// 	if tx.Transaction.IsPledgeTrasnaction() {
			// 		err := bc.handlePledgeTransaction(block.Height, &tx.Transaction)
			// 		if err != nil {
			// 			logger.Error("Failed to Pledge", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
			// 			return err
			// 		}
			// 	} else if tx.IsPledgeBreakTransaction() {
			// 		des, err := bc.handlePledgeBreakTransaction(block, &tx.Transaction)
			// 		if err != nil {
			// 			logger.Error("handlePledgeBreakTransaction Failed", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
			// 			return err
			// 		}
			// 		tx.Input = miscellaneous.EB64func(destroy)
			// 		destroy = des
			// 	}
			// 	{
			// 		afp, _ := getTotalPledge(bc.sdb, tx.From)
			// 		logger.SugarLogger.Debugf("After pledge from[%v] total pledge[%v]\n", tx.From, afp)
			// 	}

			// 	tx.GasUsed = tx.Transaction.GasLimit * tx.Transaction.GasPrice
			// 	blockGasU += tx.GasUsed
			// 	//logger.SugarLogger.Error(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>tx.GasUsed", tx.GasUsed)
			// 	if err = setMinerFee(bc, block.Miner, tx.GasUsed); err != nil {
			// 		logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)), zap.Uint64("gasUsed", tx.GasUsed))
			// 		return err
			// 	}
			// 	if err := setAccount(bc, tx, destroy); err != nil {
			// 		logger.Error("Failed to set balance", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.Uint64("amount", tx.Transaction.Amount))
			// 		return err
			// 	}
		} else if tx.Transaction.IsEvmContractTransaction() {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
				return err
			}

			gasLeft, err := bc.handleContractTransaction(block, DBTransaction, tx, index)
			if err != nil {
				logger.Error("Failed to HandleContractTransaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
				return err
			}

			evmcfg := bc.evm.GetConfig()
			// if evmcfg.GasLimit < gasLeft {
			if gasLeft.Cmp(new(big.Int).SetUint64(evmcfg.GasLimit)) > 0 {
				logger.Error("Failed to HandleContractTransaction", zap.Error(fmt.Errorf("hash[%v],evm gaslimit[%v] < gasLeft[%v]", transaction.HashToString(txHash), evmcfg.GasLimit, gasLeft)))
				return fmt.Errorf("error: hash[%v],evm gaslimit[%v] < gasLeft[%v]", transaction.HashToString(txHash), evmcfg.GasLimit, gasLeft)
			}
			// tx.GasUsed = evmcfg.GasLimit - gasLeft
			// blockGasU += tx.GasUsed
			tx.GasUsed = new(big.Int).Sub(new(big.Int).SetUint64(evmcfg.GasLimit), gasLeft)
			blockGasU.Add(blockGasU, tx.GasUsed)

			if err = setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)), zap.String("gasUsed", tx.GasUsed.String()))
				return err
			}

			// update balance
			if err := setAccount(bc, tx); err != nil {
				logger.Error("Failed to set balance", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				return err
			}
		} else {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				return err
			}

			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				return err
			}
			// update nonce,txs in block must be ordered
			nonce := tx.Transaction.Nonce + 1
			if err := setNonce(bc.sdb, *tx.Transaction.From, nonce); err != nil {
				logger.Error("Failed to set nonce", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				return err
			}

			//err

			// tx.GasUsed = tx.Transaction.GasLimit * tx.Transaction.GasPrice
			// blockGasU += tx.GasUsed
			tx.GasUsed = new(big.Int).Mul(tx.GasLimit, tx.GasPrice)
			blockGasU.Add(blockGasU, tx.GasUsed)

			if err = setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("from address", block.Miner.String()), zap.String("fee", tx.GasUsed.String()))
				return err
			}

			// update balance
			if err := setAccount(bc, tx); err != nil {
				logger.Error("Failed to set balance", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				return err
			}
		}
	}

	// {
	// 	//release 70% Mined
	// 	err := bc.releaseMined(block)
	// 	if err != nil {
	// 		logger.Error("Failed to Release Mined", zap.Error(err))
	// 		return err
	// 	}
	// }
	{
		//release pledge
		// err := bc.releasePledge(block.Height)
		// if err != nil {
		// 	logger.Error("Failed to Release Plege", zap.Error(err))
		// 	return err
		// }
	}

	comHash, err := factCommit(bc.sdb, true)
	if err != nil {
		logger.Error("Failed to set factCommit", zap.Error(err))
		return err
	}

	// func() {
	// 	host, _ := getClientIp()

	// 	f, err := os.Create(fmt.Sprintf("%d-%s-%s", block.Height, hex.EncodeToString(block.Hash[:2]), host[len(host)-3:]))
	// 	if err != nil {
	// 		return
	// 	}
	// 	defer f.Close()

	// 	s := []string{
	// 		"otK06f341e65ffdeb15c3a17cedd2d5c787631b9c9b8",
	// 		"otK4VKLJvdXexvLgQrM42BcPFQ2MCn7k97KVA8abwBNhsF4",
	// 		"otKDSbdJ5WYdeSFb4L73VWZzTD3uQGEAqarFnKoDtFiddyU",
	// 		"otK9sFhbjDdjEHvcdH6n9dtQws1m4ptsAWAy7DhqGdrUFai",
	// 		"otKE3xiLZ2V3EekoqcA5iUbewkpXdEU5uUjECrQbu9ScHUS",
	// 		"otK024969ff38e1e0aad64a5e59d92d018621b5af33a",
	// 		"otK1R6xwQUdzSvF3B9tkRdhfJtEiPZRQUpdy57m3nbksJHM",
	// 		"otK06e007bef2a2379a075552f8d9c4cfa0d13b72daa",
	// 		"otK9NaPZxx2YiVKhxfXgEu5XASU7cDhctMCw1EhLUcmn45R",
	// 		"otK01853fb6f2a7322fa1e8175287ca4aadd2c5eb7a1",
	// 		"otKBDGQRGMbQxrGHaj9PXADkWaRB14rjWvnG8Rco9WtDryy",
	// 		"otK0b6adddf480c7e055fa73bfbe291d4949c02b8de2",
	// 		"otKHH7bxS99D6224HEe28fpqFkBqutjs1mjGVCD25rQpiKT",
	// 		"otK00ec366a1750672669827c2eb442ca70bf6841441",
	// 	}

	// 	for _, v := range s {
	// 		from, _ := address.NewAddrFromString(v)
	// 		caller, _ := (&from).NewCommonAddr()
	// 		callerBalance := sdb.GetBalance(caller)
	// 		callerNonce := sdb.GetNonce(caller)
	// 		pledge, _ := getTotalMined(sdb, from)
	// 		f.WriteString(fmt.Sprintf("addr:%s nonce:%d bal:%s pledge:%d\n", from.String(), callerNonce, callerBalance.String(), pledge))
	// 	}

	// 	// m := make(map[string]string)
	// 	// for _, tx := range block.Transactions {
	// 	// 	from := tx.Caller()
	// 	// 	if _, ok := m[from.String()]; !ok {
	// 	// 		caller, _ := (&from).NewCommonAddr()
	// 	// 		callerBalance := sdb.GetBalance(caller)
	// 	// 		f.WriteString(fmt.Sprintf("%s:%s\n", from.String(), callerBalance.String()))
	// 	// 	}

	// 	// 	to := tx.Receiver()
	// 	// 	if _, ok := m[to.String()]; !ok {
	// 	// 		Receiver, _ := (&to).NewCommonAddr()
	// 	// 		ReceiverBalance := sdb.GetBalance(Receiver)
	// 	// 		f.WriteString(fmt.Sprintf("%s:%s\n", to.String(), ReceiverBalance.String()))
	// 	// 	}
	// 	// }
	// }()

	if block.Height != 1 {
		oldSnapRootkey, err := DBTransaction.Get(SnapRootKey)
		if err != nil {
			return err
		}

		logger.Info("AddTempBlock", zap.String("oldSnapRootkey", hex.EncodeToString(oldSnapRootkey)))
		logger.Info("AddTempBlock", zap.String("b.SnapRootkey", hex.EncodeToString(block.SnapRoot)))
		logger.Info("AddTempBlock", zap.String("newSnapRootkey", hex.EncodeToString(comHash.Bytes())))

		if bytes.Equal(oldSnapRootkey, block.SnapRoot) {
			block.SnapRoot = comHash.Bytes()
		} else if !bytes.Equal(comHash.Bytes(), block.SnapRoot) {
			return fmt.Errorf("SnapRoot not equal")
		}
	} else {
		block.SnapRoot = comHash.Bytes()
	}

	// hash -> block
	block.GasUsed = blockGasU
	data, err := block.Serialize()
	if err != nil {
		logger.Error("failed serialize block", zap.Error(err))
		return err
	}

	if err = DBTransaction.Set(hash, data); err != nil {
		logger.Error("Failed to set block", zap.Error(err))
		return err
	}

	{
		//	bc.updateNewStateDB(bc.sdb)
		comHashWrite, err := factCommit(bc.sdb, true)
		if err != nil {
			logger.Error("Failed to set factCommit", zap.Error(err))
			return err
		}
		if err = DBTransaction.Set(append(SnapRootPrefix, miscellaneous.E64func(height)...), comHashWrite.Bytes()); err != nil {
			logger.Error("Failed to set height and hash", zap.Error(err))
			return err
		}
		DBTransaction.Del(SnapRootKey)
		DBTransaction.Set(SnapRootKey, comHashWrite.Bytes())
	}

	logger.Info("End addBlock", zap.Uint64("blockHeight", block.Height))
	return nil
}

func getClientIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}

		}
	}

	return "", errors.New("Can not find the client ip address!")

}

func updateNewStateByRoot(bc *Blockchain, root common.Hash) (*state.StateDB, error) {
	cdb := bgdb.NewBadgerDatabase(bc.db)
	sdb := state.NewDatabase(cdb)
	stdb, err := state.New(root, sdb, nil)
	if err != nil {
		logger.Error("failed to new state")
		return nil, err
	}
	//bc.sdb = stdb
	return stdb, nil
}

//   回滚到指定快照的余额
func updateNewState(bc *Blockchain, root common.Hash) error {
	var height, prevHeight uint64
	prevHeight, err := bc.getMaxBlockHeight()
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		return err
	}
	height = prevHeight + 1
	if err := bc.db.Set(append(SnapRootPrefix, miscellaneous.E64func(height)...), root.Bytes()); err != nil {
		logger.Error("Failed to set height and hash", zap.Error(err))
		return err
	}
	if err := bc.db.Del(SnapRootKey); err != nil {
		logger.Error("Failed to bc.db.Del", zap.Error(err))
		return err
	}
	if err := bc.db.Set(SnapRootKey, root.Bytes()); err != nil {
		logger.Error("Failed to bc.db.Set", zap.Error(err))
		return err
	}
	cdb := bgdb.NewBadgerDatabase(bc.db)
	sdb := state.NewDatabase(cdb)
	stdb, err := state.New(root, sdb, nil)
	if err != nil {
		logger.Error("failed to new state")
		return err
	}
	//bc = &Blockchain{sdb: stdb}
	bc.sdb = stdb
	return nil
}

// DeleteBlock delete some blocks from the blockchain
// DeleteBlock(10):delete block data larger than 10, including 10
func (bc *Blockchain) DeleteTempBlockTest(height uint64, DBTransaction store.Transaction) error {

	dbHeight, err := bc.getMaxBlockHeight()
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		return err
	}

	/* {
		fileName := fmt.Sprintf("DeleteTempBlockTest_%v", dbHeight)
		readStatedbDate(bc.sdb, fileName)
	} */

	if height > dbHeight {
		return fmt.Errorf("Wrong height to delete,[%v] should <= current height[%v]", height, dbHeight)
	}

	for dH := dbHeight; dH >= height; dH-- {

		logger.Info("Start to delete block", zap.Uint64("height", dH))
		block, err := bc.getBlockByHeight(dH)
		if err != nil {
			logger.Error("failed to get block", zap.Error(err))
			return err
		}

		for i, tx := range block.Transactions {
			if tx.IsCoinBaseTransaction() {
				if err := deleteTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), *tx, uint64(i)); err != nil {
					logger.Error("Failed to del transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
						zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
					return err
				}
			} else if tx.IsEvmContractTransaction() {
				//新链无绑定地址接口，暂忽略setBindingKey
				if err := deleteTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), *tx, uint64(i)); err != nil {
					logger.Error("Failed to del transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
						zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
					return err
				}

			} else {
				if tx.Transaction.Input != nil || !bytes.Equal(tx.Transaction.Input, []byte("")) || len(tx.Transaction.Input) != 0 {
					spilt := strings.Split(string(tx.Transaction.Input), "\"")
					if spilt[0] == "new " {
						if err := delTokenKey(DBTransaction, spilt[1]); err != nil {
							logger.Error("failed to delTokenKey", zap.Error(err))
							return err
						}
					}
				}

				if err := deleteTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), *tx, uint64(i)); err != nil {
					logger.Error("Failed to del transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
						zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
					return err
				}

				if err := deleteTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), *tx, uint64(i)); err != nil {
					logger.Error("Failed to del transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
						zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
					return err
				}
			}

		}

		// process snapshot
		sn, err := DBTransaction.Get(append(SnapRootPrefix, miscellaneous.E64func(block.Height-1)...))
		if err != nil {
			logger.Error("Failed to DBTransaction.Get", zap.Error(err))
			return err
		}

		if err = DBTransaction.Del(append(SnapRootPrefix, miscellaneous.E64func(block.Height)...)); err != nil {
			logger.Error("Failed to DBTransaction.Del", zap.Error(err))
			return err
		}

		// Do not delete hash
		// height -> hash
		if err = DBTransaction.Del(append(HeightPrefix, miscellaneous.E64func(block.Height)...)); err != nil {
			logger.Error("Failed to Del height and hash", zap.Error(err))
			return err
		}
		// hash -> block
		//	hash := block.Hash
		//	if err = DBTransaction.Del(hash); err != nil {
		//		logger.Error("Failed to Del block", zap.Error(err))
		//		return err
		//	}

		DBTransaction.Set(SnapRootKey, sn)
		DBTransaction.Set(HeightKey, miscellaneous.E64func(dH-1))

		{
			//previous set block into into evm
			previousbBlock, err := bc.getBlockByHeight(dH - 1)
			if err != nil {
				logger.Error("failed to get block", zap.Error(err))
				return err
			}

			previousMiner := *previousbBlock.Miner
			// if err != nil {
			// 	logger.Error("failed to previousbBlock.Miner.NewCommonAddr", zap.Error(err))
			// 	return err
			// }
			bc.evm.SetBlockInfo(previousbBlock.Height, previousbBlock.Timestamp, previousMiner, previousbBlock.GlobalDifficulty)
		}

	}

	root, err := DBTransaction.Get(SnapRootKey)
	if err != nil {
		logger.Error("failed to get SnapRootKey", zap.Error(err))
		return err
	}

	sdb, err := updateNewStateByRoot(bc, common.BytesToHash(root))
	if err != nil {
		logger.Error("failed to updateNewStateByRoot", zap.Error(err))
		return err
	}
	bc.sdb = sdb

	bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))

	logger.Info("End delete")
	return nil
}

func delTokenKey(DBTransaction store.Transaction, tokenId string) error {

	if err := DBTransaction.Del([]byte(tokenId)); err != nil {
		logger.Error("failed to DBTransaction.Del to byte", zap.Error(err))
		return err
	}

	if err := DBTransaction.Del(pKey(tokenId)); err != nil {
		logger.Error("failed to DBTransaction.Del to pkey", zap.Error(err))
		return err
	}

	return nil
}

func pKey(id string) []byte {
	var buf bytes.Buffer
	buf.WriteString("p")
	buf.WriteString("/")
	buf.WriteString(id)
	return buf.Bytes()
}
