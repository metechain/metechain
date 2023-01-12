package client

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"metachain/pkg/block"
	"metachain/pkg/blockchain"
	"metachain/pkg/config"
	"metachain/pkg/logger"
	"metachain/pkg/transaction"
	"metachain/pkg/txpool"

	"go.uber.org/zap"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type Client struct {
	Bc  blockchain.Blockchains
	Tp  *txpool.Pool
	Cfg *config.CfgInfo
}

//new Client
func New(bc blockchain.Blockchains, tp *txpool.Pool, cfg *config.CfgInfo) *Client {
	return &Client{Bc: bc, Tp: tp, Cfg: cfg}
}

//cal contract
func (c *Client) ContractCall(origin string, contractAddr string, callInput, value string) (string, string, error) {
	return c.Bc.CallSmartContract(contractAddr, origin, callInput, value)
}

//Get from Balance
func (c *Client) GetBalance(from string) (*big.Int, error) {
	// addr, err := address.StringToAddress(from)
	// if err != nil {
	// 	return nil, err
	// }
	addr := common.HexToAddress(from)
	return c.Bc.GetBalance(&addr)
}

//get block by hash
func (c *Client) GetBlockByHash(hash string) (*block.Block, error) {
	h, err := transaction.StringToHash(blockchain.Check0x(hash))
	if err != nil {
		return nil, err
	}
	return c.Bc.GetBlockByHash(h)
}

//get block by Number
func (c *Client) GetBlockByNumber(num uint64) (*block.Block, error) {
	return c.Bc.GetBlockByHeight(num)
}

//Get Transaction By Hash
func (c *Client) GetTransactionByHash(hash string) (*transaction.FinishedTransaction, error) {
	h, err := transaction.StringToHash(blockchain.Check0x(hash))
	if err != nil {
		return nil, err
	}
	ftx, err := c.Bc.GetTransactionByHash(h)
	if err != nil {
		stx, err := c.Tp.GetTxByHash(blockchain.Check0x(hash))
		if err != nil {
			return nil, err
		}
		return &transaction.FinishedTransaction{SignedTransaction: *stx}, nil
	}
	return ftx, nil
}

//Get Code by contract Address
func (c *Client) GetCode(contractAddr string) string {
	return common.Bytes2Hex(c.Bc.GetCode(contractAddr))
}

//Get address Nonce
func (c *Client) GetNonce(from string) (uint64, error) {
	// addr, err := address.StringToAddress(from)
	// if err != nil {
	// 	return 0, err
	// }
	addr := common.HexToAddress(from)
	return c.Bc.GetNonce(&addr)
}

//Send signed Transaction
func (c *Client) SendRawTransaction(rawTx string) (string, error) {
	arr := strings.Split(rawTx, "0x0x0x")
	var msgHash []byte
	var metaFrom string
	if len(arr) > 1 {
		arrmk := strings.Split(arr[1], "0x0x")
		if len(arrmk) > 1 {
			msgHash, _ = hex.DecodeString(arrmk[0])

			kaddr, _ := hex.DecodeString(arrmk[1])
			metaFrom = base58.Encode(kaddr)
		}
		rawTx = arr[0]
	}

	decTX, err := hexutil.Decode(rawTx)
	if err != nil {
		return "", err
	}
	var tx types.Transaction

	err = rlp.DecodeBytes(decTX, &tx)
	if err != nil {
		return "", err
	}

	signer := types.NewEIP2930Signer(tx.ChainId())
	mas, err := tx.AsMessage(signer, nil)
	if err != nil {
		return "", err
	}

	logger.InfoLogger.Printf("chainserver sendRawTransaction:{mas.From:%v,to:%v,amount:%v,nounce:%v,hash:%v,gas:%v,gasPrice:%v,txType:%v,chainId:%v,tx lenght:%v}\n", mas.From(), tx.To(), tx.Value(), tx.Nonce(), tx.Hash(), tx.Gas(), tx.GasPrice(), tx.Type(), tx.ChainId(), len(tx.Data()))

	return c.sendRawTransaction(mas.From().Hex(), rawTx, metaFrom, msgHash)
}

//send eth signed transaction
func (g *Client) sendRawTransaction(EthFrom, EthData, metaFrom string, MsgHash []byte) (string, error) {
	ethFrom := common.HexToAddress(EthFrom)
	ethData := EthData

	if len(ethData) <= 0 {
		return "", fmt.Errorf("Wrong eth signed data:%s", fmt.Errorf("ethData length[%v] <= 0", len(ethData)).Error())
	}

	var from, to common.Address
	var evm transaction.EvmContract
	var amount, gasPrice, gasLimit = new(big.Int), new(big.Int), new(big.Int)
	var tag transaction.TransactionType
	//	var signType crypto.SigType

	if len(metaFrom) > 0 && len(MsgHash) > 0 {
		// f, err := address.NewAddrFromString(metaFrom)
		// if err != nil {
		// 	return "", err
		// }
		from = common.HexToAddress(metaFrom)
		//signType = crypto.TypeED25519
		evm.MsgHash = MsgHash
	} else {
		from = common.HexToAddress(EthFrom)
		//signType = crypto.TypeSecp256k1
	}

	ethTx, err := transaction.DecodeEthData(ethData)
	if err != nil {
		return "", fmt.Errorf("decodeData error:%s", err.Error())
	}
	if ethTx.ChainId().Int64() != g.Cfg.ChainCfg.ChainId {
		return "", fmt.Errorf("error chain ID: %v", ethTx.ChainId().Int64())
	}
	evm.EthData = EthData

	if len(ethTx.Data()) > 0 {
		to = common.Address{}
		amount = new(big.Int)
		if ethTx.To() == nil {
			evm.Origin = ethFrom
			evm.CreateCode = ethTx.Data()
			evm.Operation = blockchain.CREATECONTRACT
		} else {
			evm.ContractAddr = *ethTx.To()
			evm.CallInput = ethTx.Data()
			evm.Operation = blockchain.CALLCONTRACT
			evm.Origin = ethFrom
		}
		tag = transaction.EvmContractTransaction
	} else {
		to = *ethTx.To()

		deci := blockchain.Uint64ToBigInt(blockchain.ETHDECIMAL)
		value := ethTx.Value().Div(ethTx.Value(), deci)
		// amount = value.Uint64()
		amount.Set(value)
		tag = transaction.EvmMetaTransaction
		evm.Status = true
	}

	if ethTx.GasPrice().Uint64() == 0 {
		gasPrice = new(big.Int).SetUint64(g.Cfg.ChainCfg.GasPrice)
	} else {
		// gasPrice = ethTx.GasPrice().Uint64()
		gasPrice.Set(ethTx.GasPrice())
	}

	if ethTx.Gas() == 0 {
		// gasLimit = g.Cfg.ChainCfg.GasLimit
		gasLimit = new(big.Int).SetUint64(g.Cfg.ChainCfg.GasLimit)
	} else {
		gasLimit = new(big.Int).SetUint64(ethTx.Gas())
	}

	input, err := transaction.EncodeEvmData(&evm)
	if err != nil {
		return "", fmt.Errorf("EncodeEvmData error:%s", err.Error())
	}

	nonce, err := g.Bc.GetNonce(&from)
	if err != nil {
		return "", fmt.Errorf("GetNonce error:%s,from:%v", err.Error(), from)
	}
	if nonce != ethTx.Nonce() {
		return "", fmt.Errorf("error: from [%v] nonce[%v] not equal ethTx.Nonce[%v]", from, nonce, ethTx.Nonce())
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      &from,
			To:        &to,
			Amount:    amount,
			Nonce:     ethTx.Nonce(),
			GasLimit:  gasLimit,
			GasPrice:  gasPrice,
			GasFeeCap: gasLimit,
			Type:      tag,
			Input:     input,
		},
		Signature: transaction.ParseEthSignature(&ethTx),
		/* 		Signature: crypto.Signature{
			SigType: signType,
			Data:    transaction.ParseEthSignature(&ethTx),
		}, */
	}
	err = g.Tp.Add(tx)
	if err != nil {
		return "", fmt.Errorf("add tx pool error:%s", err.Error())
	}

	logger.Info("rpc End SendEthSignedRawTransaction:", zap.String("hash", tx.HashToString()))
	return tx.HashToString(), nil
}

//Get Transaction Receipt by hash
func (c *Client) GetTransactionReceipt(hash string) (*transaction.FinishedTransaction, error) {
	h, err := transaction.StringToHash(blockchain.Check0x(hash))
	if err != nil {
		return nil, err
	}
	return c.Bc.GetTransactionByHash(h)
}

//GetStorageAt
func (c *Client) GetStorageAt(addr, hash string) string {
	return c.Bc.GetStorageAt(addr, hash).Hex()
}

//get Logs
/* func (c *Client) GetLogs(address string, fromB, toB uint64, topics []string, blockH string) []*types.Log {
	logs := c.Bc.GetLogs()
	var resLogs []*types.Log
	for _, log := range logs {
		if common.HexToHash(blockH) == log.BlockHash {
			resLogs = append(resLogs, log)
		} else if log.BlockNumber >= fromB && log.BlockNumber <= toB {
			if common.HexToAddress(address) == log.Address {
				resLogs = append(resLogs, log)
			}
		}
	}
	return resLogs
} */
func (c *Client) GetLogs(address string, fromB, toB uint64, topics []string, blockH string) []*types.Log {

	if fromB > toB {
		return nil
	}

	tipblock, err := c.Bc.Tip()
	if err != nil {
		return nil
	}
	if tipblock.Height < toB {
		fmt.Printf("toblock[%v]>maxBlockHeight[%v]", toB, tipblock.Height)
		return nil
	}

	var logs []*types.Log

	/* if len(blockH) != 0 {

		hash, err := transaction.StringToHash(blockchain.Check0x(blockH))
		if err != nil {
			fmt.Printf("invalid hash[%s]", blockH)
			return nil
		}

		tx, err := c.Bc.GetTransactionByHash(hash)
		if err != nil {
			fmt.Printf("invalid hash[%s]", hash)
			return nil
		}
		log := c.Bc.GetLogByHeight(tx.BlockNum)

		return log
	} */
	if len(blockH) != 0 {

		hash, err := transaction.StringToHash(blockchain.Check0x(blockH))
		if err != nil {
			fmt.Printf("invalid hash[%s]", blockH)
			return nil
		}

		b, err := c.Bc.GetBlockByHash(hash)
		if err != nil {
			fmt.Printf("invalid hash[%s]", hash)
			return nil
		}
		log := c.Bc.GetLogByHeight(b.Height)

		return log
	}
	for i := fromB; i <= toB; i++ {
		log := c.Bc.GetLogByHeight(i)
		logs = append(logs, log...)
	}
	return logs

}

//Get Max BlockNumber
func (c *Client) GetMaxBlockNumber() (uint64, error) {
	return c.Bc.GetMaxBlockHeight()
}
