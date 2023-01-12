package grpcserver

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"

	"metachain/pkg/blockchain"
	"metachain/pkg/config"
	_ "metachain/pkg/crypto/sigs/ed25519"
	_ "metachain/pkg/crypto/sigs/secp"
	"metachain/pkg/logger"
	"metachain/pkg/miner"
	"metachain/pkg/p2p"
	"metachain/pkg/server"
	"metachain/pkg/server/grpcserver/message"

	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/ethereum/go-ethereum/core/types"

	"metachain/pkg/transaction"
	"metachain/pkg/txpool"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const DECIWEI = 1000000000

var Deciwei = new(big.Int).Mul(big.NewInt(DECIWEI), big.NewInt(10000000))

type Greeter struct {
	Bc       blockchain.Blockchains
	Tp       *txpool.Pool
	Cfg      *config.CfgInfo
	Node     *p2p.Node
	Miner    *miner.Miner
	NodeName string
}

var _ message.GreeterServer = (*Greeter)(nil)

func NewGreeter(bc blockchain.Blockchains, tp *txpool.Pool, cfg *config.CfgInfo) *Greeter {
	return &Greeter{Bc: bc, Tp: tp, Cfg: cfg}
}

func (g *Greeter) RunGrpc() {
	lis, err := net.Listen("tcp", g.Cfg.SververCfg.GRpcAddress)
	if err != nil {
		logger.Error("net.Listen", zap.Error(err))
		os.Exit(-1)
	}

	server := grpc.NewServer(grpc.UnaryInterceptor(server.IpInterceptor))

	message.RegisterGreeterServer(server, g)

	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}

func (g *Greeter) GetBalance(ctx context.Context, in *message.ReqBalance) (*message.ResBalance, error) {

	addr := common.HexToAddress(in.Address)
	balance, err := g.Bc.GetBalance(&addr)
	if err != nil {
		logger.Error("g.Bc.GetBalance", zap.Error(err), zap.String("address", in.Address))
		return nil, err
	}
	return &message.ResBalance{Balance: balance.String()}, nil
}

func (g *Greeter) SendTransaction(ctx context.Context, in *message.ReqTransaction) (*message.ResTransaction, error) {

	// from, err := address.NewAddrFromString(in.From)
	// if err != nil {
	// 	return nil, err
	// }

	from := common.HexToAddress(in.From)
	to := common.HexToAddress(in.To)

	// to, err := address.NewAddrFromString(in.To)
	// if err != nil {
	// 	return nil, err
	// }

	balance, err := g.Bc.GetAvailableBalance(&from)
	if err != nil {
		return nil, err
	}
	// if in.Amount+in.GasLimit*in.GasPrice > balance {
	// 	return nil, fmt.Errorf("from(%v) balance(%v) is not enough or out of gas.", from, balance)
	// }
	amount, ok := new(big.Int).SetString(in.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("amount:%s", in.Amount))
	}

	gasPrice, ok := new(big.Int).SetString(in.GasPrice, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("gas price:%s", in.GasPrice))
	}

	gaslimit, ok := new(big.Int).SetString(in.GasLimit, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("gas limit:%s", in.GasLimit))
	}

	gasFeeCap, ok := new(big.Int).SetString(in.GasFeeCap, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf(" gas fee cap:%s", in.GasFeeCap))
	}

	// if in.GasLimit*in.GasPrice == 0 {
	// 	return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	// }
	gas := new(big.Int).Mul(gasPrice, gaslimit)
	if gas.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	if balance.Cmp(new(big.Int).Add(amount, gas)) == -1 {
		return nil, fmt.Errorf("from(%v) balance(%v) is not enough or out of gas.", from, balance)
	}

	// sign, err := crypto.DeserializeSignature(in.Sign)
	// if err != nil {
	// 	return nil, err
	// }

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      &from,
			To:        &to,
			Amount:    amount,
			Nonce:     in.Nonce,
			GasLimit:  gaslimit,
			GasPrice:  gasPrice,
			GasFeeCap: gasFeeCap,
			Input:     in.Input,
			Type:      transaction.TransferTransaction,
		},
		Signature: in.Sign,
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}

	data, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	if g.Node != nil {
		g.Node.SendMessage(p2p.PayloadMessageType, append([]byte{0}, data...))
	}

	hash := hex.EncodeToString(tx.Hash())

	return &message.ResTransaction{Hash: hash}, nil
}

/* func (g *Greeter) SendTransaction(ctx context.Context, in *message.ReqTransaction) (*message.ResTransaction, error) {
	return nil, fmt.Errorf("stop translation")
} */

/* func (g *Greeter) SendLockTransaction(ctx context.Context, in *message.ReqTransaction) (*message.ResTransaction, error) {

	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}
	balance, err := g.Bc.GetAvailableBalance(from)
	if err != nil {
		return nil, err
	}
	if in.Amount+in.GasLimit*in.GasPrice > balance {
		return nil, errors.New("freeze balance is not enough or out of gas.")
	}

	sign, err := crypto.DeserializeSignature(in.Sign)
	if err != nil {
		return nil, err
	}

	if in.GasLimit*in.GasPrice == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      from,
			To:        to,
			Amount:    in.Amount,
			Nonce:     in.Nonce,
			GasLimit:  in.GasLimit,
			GasPrice:  in.GasPrice,
			GasFeeCap: in.GasFeeCap,
			Input:     in.Input,
			Type:      transaction.LockTransaction,
		},
		Signature: *sign,
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}
	hash := hex.EncodeToString(tx.Hash())

	return &message.ResTransaction{Hash: hash}, nil
} */

/* func (g *Greeter) SendUnlockTransaction(ctx context.Context, in *message.ReqTransaction) (*message.ResTransaction, error) {
	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}

	balance, err := g.Bc.GetAllFreezeBalance(from)
	if err != nil {
		return nil, err
	}

	available, err := g.Bc.GetAvailableBalance(from)
	if err != nil {
		return nil, err
	}

	FreezeBalance, err := g.Bc.GetSingleFreezeBalance(from, to)
	if err != nil {
		return nil, err
	}

	if in.GasLimit*in.GasPrice > available || in.Amount > balance || in.Amount > FreezeBalance {
		return nil, errors.New("freeze balance is not enough or out of gas.")
	}

	sign, err := crypto.DeserializeSignature(in.Sign)
	if err != nil {
		return nil, err
	}

	if in.GasLimit*in.GasPrice == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      from,
			To:        to,
			Amount:    in.Amount,
			Nonce:     in.Nonce,
			GasLimit:  in.GasLimit,
			GasPrice:  in.GasPrice,
			GasFeeCap: in.GasFeeCap,
			Input:     in.Input,
			Type:      transaction.UnlockTransaction,
		},
		Signature: *sign,
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}
	hash := hex.EncodeToString(tx.Hash())

	return &message.ResTransaction{Hash: hash}, nil
} */

func (g *Greeter) GetBlockByNum(ctx context.Context, in *message.ReqBlockByNumber) (*message.RespBlock, error) {
	b, err := g.Bc.GetBlockByHeight(in.Height)
	if err != nil {
		logger.Error("GetBlockByHeight", zap.Error(fmt.Errorf("error:%v,height %v", err.Error(), in.Height)))
		return &message.RespBlock{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}
	blockbyte, err := b.Serialize()
	if err != nil {
		logger.Error("Serialize", zap.String("error", err.Error()))
		return &message.RespBlock{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}
	return &message.RespBlock{Data: blockbyte, Code: 0}, nil

}

func (g *Greeter) GetBlockByHash(ctx context.Context, in *message.ReqBlockByHash) (*message.RespBlockDate, error) {
	hash, err := transaction.StringToHash(blockchain.Check0x(in.Hash))
	if err != nil {
		logger.Error("StringToHash", zap.String("error", err.Error()), zap.String("hash", in.Hash))
		return &message.RespBlockDate{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}
	block, err := g.Bc.GetBlockByHash(hash)
	if err != nil {
		//logger.Error("GetBlockByHash", zap.Error(fmt.Errorf("error:%v,hash: %v", err.Error(), in.Hash)))
		return &message.RespBlockDate{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}
	blockbyte, err := block.Serialize()
	if err != nil {
		logger.Error("Serialize", zap.String("error", err.Error()))
		return &message.RespBlockDate{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}
	return &message.RespBlockDate{Data: blockbyte, Code: 0}, nil
}

func (g *Greeter) GetTxByHash(ctx context.Context, in *message.ReqTxByHash) (*message.RespTxByHash, error) {
	hash, _ := transaction.StringToHash(blockchain.Check0x(in.Hash))
	tx, err := g.Bc.GetTransactionByHash(hash)
	if err != nil {
		//logger.Error("GetTransactionByHash", zap.Error(fmt.Errorf("error:%v,hash:%v", err.Error(), in.Hash)))
		st, err := g.Tp.GetTxByHash(in.Hash)
		if err != nil {
			return &message.RespTxByHash{Data: []byte{}, Code: -1, Message: err.Error()}, nil
		}
		tx = transaction.NewFinishedTransaction(st, nil, 0)
	}

	txbytes, err := tx.Serialize()
	if err != nil {
		logger.Error("tx Serialize", zap.String("error", err.Error()))
		return &message.RespTxByHash{Data: []byte{}, Code: -1, Message: err.Error()}, nil
	}

	return &message.RespTxByHash{Data: txbytes, Code: 0}, nil
}

func (g *Greeter) GetAddressNonceAt(ctx context.Context, in *message.ReqNonce) (*message.ResposeNonce, error) {

	addr := common.HexToAddress(in.Address)
	resp, err := g.Bc.GetNonce(&addr)
	if err != nil {
		return nil, err
	}

	return &message.ResposeNonce{Nonce: resp}, nil
}

//send eth signed transaction
func (g *Greeter) SendEthSignedRawTransaction(ctx context.Context, in *message.ReqEthSignTransaction) (*message.ResEthSignTransaction, error) {
	ethFrom := common.HexToAddress(in.EthFrom)
	ethData := in.EthData

	l := len(ethData)
	if l <= 0 {
		logger.Error("Wrong eth signed data", zap.Error(fmt.Errorf("ethData length[%v] <= 0", len(ethData))))
		return nil, fmt.Errorf("Wrong eth signed data:%s", fmt.Errorf("ethData length[%v] <= 0", len(ethData)).Error())
	}
	logger.Info("rpc Into SendEthSignedRawTransaction:", zap.String("from", in.EthFrom), zap.Int32("data lenght", int32(l)))
	var from, to common.Address
	var evm transaction.EvmContract
	var amount, gasPrice, gasLimit = big.NewInt(0), big.NewInt(0), big.NewInt(0)
	var tag transaction.TransactionType

	if len(in.KtoFrom) > 0 && len(in.MsgHash) > 0 {
		from = common.HexToAddress(in.KtoFrom)
		evm.MsgHash = in.MsgHash
	} else {
		from = common.HexToAddress(in.EthFrom)
	}

	ethTx, err := transaction.DecodeEthData(ethData)
	if err != nil {
		logger.Info("SendEthSignedTransaction decodeData error", zap.Error(err))
		return nil, fmt.Errorf("decodeData error:%s", err.Error())
	}
	if ethTx.ChainId().Int64() != g.Cfg.ChainCfg.ChainId {
		return nil, fmt.Errorf("error chain ID: %v", ethTx.ChainId().Int64())
	}

	evm.EthData = ethData

	if len(ethTx.Data()) > 0 {
		to = common.Address{}
		amount.SetInt64(0)
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
		// ethTo := *ethTx.To()
		// addr, err := address.StringToAddress(ethTo.Hex())
		// if err != nil {
		// 	logger.Info("StringToAddress error", zap.Error(err))
		// 	return nil, fmt.Errorf("eth addr to kto addr error:%s", err.Error())
		// }
		to = *ethTx.To()

		deci := new(big.Int).SetUint64(blockchain.ETHDECIMAL)
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
		// if gasPrice >= DECIWEI {
		// 	gasPrice = gasPrice / DECIWEI
		// }

		if gasPrice.Cmp(Deciwei) >= 0 {
			gasPrice.Div(gasPrice, Deciwei)
		}
	}

	if ethTx.Gas() == 0 {
		// gasLimit = g.Cfg.ChainCfg.GasLimit
		gasLimit = new(big.Int).SetUint64(g.Cfg.ChainCfg.GasLimit)
	} else {
		// gasLimit = ethTx.Gas()
		gasLimit = new(big.Int).SetUint64(ethTx.Gas())
	}

	input, err := transaction.EncodeEvmData(&evm)
	if err != nil {
		logger.Info("EncodeEvmData error", zap.Error(err))
		return nil, fmt.Errorf("EncodeEvmData error:%s", err.Error())
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
	}
	err = g.Tp.Add(tx)
	if err != nil {
		logger.Info("EthSignedRawTransaction add tx pool error:", zap.Error(err))
		return nil, fmt.Errorf(err.Error())
	}
	//g.n.Broadcast(tx)

	logger.Info("rpc End SendEthSignedRawTransaction:", zap.String("hash", tx.HashToString()))
	return &message.ResEthSignTransaction{Hash: tx.HashToString()}, nil
}

/* func (g *Greeter) SendEthSignedRawTransaction(ctx context.Context, in *message.ReqEthSignTransaction) (*message.ResEthSignTransaction, error) {
	return nil, fmt.Errorf("stop translation")
} */

// //send pledge transaction
// func (g *Greeter) SendSignedPledgeTransaction(ctx context.Context, in *message.ReqPledgeTransaction) (*message.ResPledgeTransaction, error) {
// 	from, err := address.StringToAddress(in.From)
// 	if err != nil {
// 		return nil, err
// 	}

// 	avib, err := g.Bc.GetAvailableBalance(from)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if avib < in.Amount+in.GasLimit*in.GasPrice {
// 		return nil, fmt.Errorf("address[%v] no enough balance to pledge or out of gas,available[%v],amount[%v],gas[%v]", from, avib, in.Amount, in.GasLimit*in.GasPrice)
// 	}

// 	sign, err := crypto.DeserializeSignature(in.Signature)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var txtype uint8 = 0
// 	if in.Type == "pledge" {
// 		if in.Amount <= 0 {
// 			return nil, fmt.Errorf("pledge amount couldn't <= 0,amount[%v]", in.Amount)
// 		}
// 		txtype = transaction.PledgeTrasnaction
// 	} else if in.Type == "break" {
// 		if in.Amount != 0 {
// 			return nil, fmt.Errorf("pledge amount should be 0,amount[%v]", in.Amount)
// 		}
// 		totP, err := g.Bc.GetTotalPledge(from)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if totP == 0 {
// 			return nil, fmt.Errorf("total pledge[%v] is 0", totP)
// 		}
// 		txtype = transaction.PledgeBreakTransaction
// 	} else {
// 		return nil, fmt.Errorf("unsupport pledge type[%v]", in.Type)
// 	}
// 	tx := &transaction.SignedTransaction{
// 		Transaction: transaction.Transaction{
// 			From:      from,
// 			To:        address.ZeroAddress,
// 			Amount:    in.Amount,
// 			Nonce:     in.Nonce,
// 			GasLimit:  in.GasLimit,
// 			GasPrice:  in.GasPrice,
// 			GasFeeCap: in.GasFeeCap,
// 			Type:      txtype,
// 		},
// 		Signature: *sign,
// 	}

// 	if err := g.Tp.Add(tx); err != nil {
// 		return nil, err
// 	}

// 	//g.n.Broadcast(tx)
// 	logger.Info("rpc End SendSignedPledgeTransaction:", zap.String("hash", tx.HashToString()))
// 	return &message.ResPledgeTransaction{Hash: tx.HashToString()}, nil
// }

/* //get eth address by kto address
func (g *Greeter) GetETHAddress(ctx context.Context, in *message.ReqKtoAddress) (*message.ResEthAddress, error) {
	// kaddr, err := address.NewAddrFromString(in.Ktoaddress)
	// if err != nil {
	// 	logger.Error("rpc NewAddrFromString", zap.String("error", err.Error()))
	// 	return &message.ResEthAddress{Message: err.Error(), Code: -1}, nil
	// }

	kaddr := common.HexToAddress(in.Ktoaddress)
	res, err := g.Bc.GetBindingEthAddress(&kaddr)
	if err != nil {
		//logger.Error("rpc GetBindingEthAddress", zap.String("error", err.Error()))
		return &message.ResEthAddress{Message: err.Error(), Code: -1}, nil
	}
	return &message.ResEthAddress{Ethaddress: res, Code: 0}, nil
}

//get address  by eth address
func (g *Greeter) GetMetaAddress(ctx context.Context, in *message.ReqEthAddress) (*message.ResKtoAddress, error) {
	res, err := g.Bc.GetBindingMetaAddress(in.Ethaddress)
	if err != nil {
		//logger.Error("rpc GetBindingKtoAddress", zap.String("error", err.Error()))
		return &message.ResKtoAddress{Message: err.Error(), Code: -1}, nil
	}
	return &message.ResKtoAddress{Ktoaddress: res.String(), Code: 0}, nil
} */

//call contract
func (g *Greeter) CallSmartContract(ctx context.Context, in *message.ReqCallContract) (*message.ResCallContract, error) {
	res, gas, err := g.Bc.CallSmartContract(in.Contractaddress, in.Origin, in.Inputcode, in.Value)
	if err != nil {
		logger.Error("rpc CallSmartContract", zap.String("Result", res), zap.String("error", err.Error()))
		return &message.ResCallContract{Result: res, Msg: err.Error(), Code: -1}, nil
	}
	return &message.ResCallContract{Result: res, Gas: gas, Code: 0}, nil
}

//get code by contract address
func (g *Greeter) GetCode(ctx context.Context, in *message.ReqEvmGetcode) (*message.ResEvmGetcode, error) {
	code := g.Bc.GetCode(in.Contract)
	if len(code) <= 0 {
		return nil, fmt.Errorf("code not exist")
	}
	return &message.ResEvmGetcode{Code: common.Bytes2Hex(code)}, nil
}

//get storage by hash
func (g *Greeter) GetStorageAt(ctx context.Context, in *message.ReqGetstorage) (*message.ResGetstorage, error) {
	ret := g.Bc.GetStorageAt(in.Addr, blockchain.Check0x(in.Hash))
	return &message.ResGetstorage{Result: ret.String()}, nil
}

//get max block height
func (g *Greeter) GetMaxBlockHeight(ctx context.Context, in *message.ReqMaxBlockHeight) (*message.ResMaxBlockHeight, error) {
	maxH, err := g.Bc.GetMaxBlockHeight()
	if err != nil {
		logger.Error("rpc GetMaxBlockHeight", zap.String("error", err.Error()))
		return nil, err
	}
	return &message.ResMaxBlockHeight{MaxHeight: maxH}, nil
}

//get evm logs
/* func (g *Greeter) Getlogs(ctx context.Context, in *message.ReqLogs) (*message.ResLogs, error) {
	if in.FromBlock > in.ToBlock {
		return nil, fmt.Errorf("Wrong fromBlock[%v] and to toBlock[%v]", in.FromBlock, in.ToBlock)
	}
	logs := g.Bc.GetLogs()
	var resLogs []*evmtypes.Log
	for _, log := range logs {
		if common.HexToHash(in.BlockHash) == log.BlockHash {
			resLogs = append(resLogs, log)
		} else if log.BlockNumber >= in.FromBlock && log.BlockNumber <= in.ToBlock {
			if common.HexToAddress(in.Address) == log.Address {
				resLogs = append(resLogs, log)
			}
		}
	}
	bslog, err := json.Marshal(&resLogs)
	if err != nil {
		return nil, err
	}
	return &message.ResLogs{Evmlogs: bslog}, nil
} */

//get evm logs
func (g *Greeter) GetLogs(ctx context.Context, in *message.ReqLogs) (*message.ResLogs, error) {

	var filter int

	if len(in.Address) > 0 {
		filter = 1

	}
	if in.FromBlock > in.ToBlock {
		return nil, fmt.Errorf("Wrong fromBlock[%v] and to toBlock[%v]", in.FromBlock, in.ToBlock)
	}

	tipblock, err := g.Bc.Tip()
	if err != nil {
		return nil, fmt.Errorf("Wrong get block info err")
	}
	if tipblock.Height < in.ToBlock {
		return nil, fmt.Errorf("toblock[%v]>maxBlockHeight[%v]", in.ToBlock, tipblock.Height)
	}

	var logs, filterlog []*evmtypes.Log

	if len(in.BlockHash) != 0 {
		hash, err := transaction.StringToHash(blockchain.Check0x(in.BlockHash))
		if err != nil {
			return nil, fmt.Errorf("invalid hash[%s]", in.BlockHash)
		}

		block, err := g.Bc.GetBlockByHash(hash)
		if err != nil {
			return nil, fmt.Errorf("invalid hash[%s]", in.BlockHash)
		}
		filterlog = g.Bc.GetLogByHeight(block.Height)
	} else {
		for i := in.FromBlock; i <= in.ToBlock; i++ {
			log := g.Bc.GetLogByHeight(i)
			filterlog = append(filterlog, log...)
		}
	}
	if filter == 1 {
		for _, l := range filterlog {
			if common.HexToAddress(in.Address) == l.Address {
				logs = append(logs, l)
			}
		}
	}

	bslog, err := json.Marshal(&logs)
	if err != nil {
		return nil, err
	}

	return &message.ResLogs{Evmlogs: bslog}, nil

}

// //get total transaction pledge and mined pledge by address
// func (g *Greeter) GetTotalPledge(ctx context.Context, in *message.ReqKtoAddress) (*message.ResPledge, error) {
// 	addr, err := address.StringToAddress(in.Ktoaddress)
// 	if err != nil {
// 		logger.Info("GetTotalPledge string address to Address error", zap.Error(err))
// 		return &message.ResPledge{Code: -1, Message: err.Error()}, nil
// 	}

// 	totalP, err := g.Bc.GetTotalPledge(addr)
// 	if err != nil {
// 		logger.Info("GetTotalPledge error", zap.Error(err))
// 		return &message.ResPledge{Code: -1, Message: err.Error()}, nil
// 	}

// 	totalM, err := g.Bc.GetTotalMined(addr)
// 	if err != nil {
// 		logger.Info("GetTotalMined error", zap.Error(err))
// 		return &message.ResPledge{Code: -1, Message: err.Error()}, nil
// 	}
// 	return &message.ResPledge{TotalPledge: totalP, TotalMined: totalM, Code: 0}, nil
// }
/*
func (g *Greeter) CreateToken(ctx context.Context, in *message.ReqTokenCreate) (*message.HashMsg, error) {

	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}

	balance, err := g.Bc.GetAvailableBalance(from)
	if err != nil {
		return nil, err
	}
	if in.GasLimit*in.GasPrice > balance {
		return nil, errors.New("freeze balance is not enough or out of gas.")
	}
	sign, err := crypto.DeserializeSignature(in.Signature)
	if err != nil {
		return nil, err
	}

	if in.GasLimit*in.GasPrice == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      from,
			To:        to,
			Nonce:     in.Nonce,
			GasLimit:  in.GasLimit,
			GasPrice:  in.GasPrice,
			GasFeeCap: in.GasFeeCap,
			Type:      transaction.TransferTransaction,
			Input:     in.Input,
		},
		Signature: *sign,
	}

	script := string(tx.Input)
	_, err = g.Bc.GetTokenRoot(from, script)
	if err != nil {
		return nil, errors.New("get token root err")
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}

	data, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	g.Node.SendMessage(p2p.PayloadMessageType, append([]byte{0}, data...))
	hash := hex.EncodeToString(tx.Hash())

	return &message.HashMsg{Hash: hash}, nil

} */
/* func (g *Greeter) MintToken(ctx context.Context, in *message.ReqTokenCreate) (*message.HashMsg, error) {

	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}

	if len(in.Symbol) == 0 {
		logger.Info("symbol", zap.String("symbol", in.Symbol))
		return &message.HashMsg{Code: -1, Message: "invalid symbol"}, nil
	}

	sign, err := crypto.DeserializeSignature(in.Signature)
	if err != nil {
		return nil, err
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      from,
			To:        to,
			Nonce:     in.Nonce,
			GasLimit:  in.GasLimit,
			GasPrice:  in.GasPrice,
			GasFeeCap: in.GasFeeCap,
			Type:      transaction.TransferTransaction,
			Input:     in.Input,
		},
		Signature: *sign,
	}

	script := string(tx.Input)
	_, err = g.Bc.GetTokenRoot(from, script)
	if err != nil {
		return nil, errors.New("get token root err")
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}

	data, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	g.Node.SendMessage(p2p.PayloadMessageType, append([]byte{0}, data...))
	hash := hex.EncodeToString(tx.Hash())

	return &message.HashMsg{Hash: hash}, nil

} */
/* func (g *Greeter) SendToken(ctx context.Context, in *message.ReqTokenTransaction) (*message.RespTokenTransaction, error) {
	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}

	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}
	if len(in.Input) == 0 {
		logger.Info("symbol", zap.String("symbol", string(in.Input)))
		return nil, fmt.Errorf("symbol:%s", in.Input)
	}

	sign, err := crypto.DeserializeSignature(in.Signature)
	if err != nil {
		return nil, err
	}

	if in.GasLimit*in.GasPrice == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      from,
			To:        to,
			Nonce:     in.Nonce,
			GasLimit:  in.GasLimit,
			GasPrice:  in.GasPrice,
			GasFeeCap: in.GasFeeCap,
			Type:      transaction.TransferTransaction,
			Input:     in.Input,
		},
		Signature: *sign,
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}

	script := string(in.Input)
	_, err = g.Bc.GetTokenRoot(from, script)
	if err != nil {
		logger.Error("Failed to get token root", zap.String("from", from.String()),
			zap.String("script", string(in.Input)))
		return nil, grpc.Errorf(codes.InvalidArgument, "get root failed")
	}

	data, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	g.Node.SendMessage(p2p.PayloadMessageType, append([]byte{0}, data...))
	hash := hex.EncodeToString(tx.Hash())

	return &message.RespTokenTransaction{Hash: hash}, nil

} */
/* func (g *Greeter) GetBalanceToken(ctx context.Context, in *message.ReqTokenBalance) (*message.RespTokenBalance, error) {
	addr, err := address.NewAddrFromString(in.Address)
	if err != nil {
		return nil, err
	}

	balance, err := g.Bc.GetTokenBalance(addr, []byte(in.Symbol))
	if err != nil {
		logger.Error("g.Bc.GetTokenBalance", zap.Error(err), zap.String("address", in.Address), zap.String("symbol", in.Symbol))
		return nil, grpc.Errorf(codes.InvalidArgument, "symbol:\"%s\",address:%s", in.Symbol, in.Address)
	}

	demic, err := g.Bc.GetTokenDemic([]byte(in.Symbol))
	if err != nil {
		logger.Error("get demic:", zap.String("symbol", in.Symbol), zap.Error(err))
		return nil, grpc.Errorf(codes.InvalidArgument, "symbol:\"%s\",address:%s", in.Symbol, in.Address)
	}

	return &message.RespTokenBalance{Demic: demic, Balnce: balance}, nil
} */

/* func (g *Greeter) GetSingleFreezeBalance(ctx context.Context, in *message.ReqSignBalance) (*message.ResBalance, error) {

	from, err := address.NewAddrFromString(in.From)
	if err != nil {
		return nil, err
	}
	to, err := address.NewAddrFromString(in.To)
	if err != nil {
		return nil, err
	}

	bal, err := g.Bc.GetSingleFreezeBalance(from, to)
	if err != nil {
		return nil, err
	}
	return &message.ResBalance{Balance: bal}, nil
} */

/* func (g *Greeter) GetAllFreezeBalance(ctx context.Context, in *message.ReqBalance) (*message.ResBalance, error) {
	from, err := address.NewAddrFromString(in.Address)
	if err != nil {
		return nil, err
	}

	bal, err := g.Bc.GetAllFreezeBalance(from)
	if err != nil {
		return nil, err
	}
	return &message.ResBalance{Balance: bal}, nil
} */

// func (g *Greeter) GetWholeNetworkPledge(ctx context.Context, in *message.ReqWholeNetworkPledge) (*message.ResWholeNetworkPledge, error) {
// 	total, err := g.Bc.GetWholeNetWorkTotalPledge()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &message.ResWholeNetworkPledge{WholeNetworkPledge: total}, nil
// }

// func (g *Greeter) GetAvailableBalance(ctx context.Context, in *message.ReqGetAvailableBalance) (*message.ResGetAvailableBalance, error) {
// 	addr, err := address.StringToAddress(in.Address)
// 	if err != nil {
// 		logger.Info("StringToAddress error", zap.Error(err))
// 		return nil, err
// 	}
// 	avi, err := g.Bc.GetAvailableBalance(addr)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &message.ResGetAvailableBalance{AvailableBalance: avi}, nil
// }

func (g *Greeter) GetHasherPerSecond(ctx context.Context, req *message.ReqHasherPerSecond) (*message.ResHasherPerSecond, error) {
	resp := &message.ResHasherPerSecond{}
	if g.Miner == nil {
		resp.HasherPerSecond = 0
		return resp, nil
	}
	h := g.Miner.HashesPerSecond()
	resp.Address = h.MinerAddr.String()
	resp.HasherPerSecond = float32(h.HasherPerSecond)
	resp.Uuid = g.NodeName
	return resp, nil
}

func (g *Greeter) WithdrawToEthaddr(ctx context.Context, in *message.Req_WithdrawToEthaddr) (*message.Res_WithdrawToEthaddr, error) {
	// from, err := address.StringToAddress(in.From)
	// if err != nil {
	// 	logger.Info("StringToAddress error", zap.Error(err))
	// 	return nil, fmt.Errorf("eth addr to kto addr error:%s", err.Error())
	// }

	// to, err := address.StringToAddress(in.To)
	// if err != nil {
	// 	logger.Info("StringToAddress error", zap.Error(err))
	// 	return nil, fmt.Errorf("eth addr to kto addr error:%s", err.Error())
	// }

	from, to := common.HexToAddress(in.From), common.HexToAddress(in.To)
	amount, ok := new(big.Int).SetString(in.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("amount:%s", in.Amount))
	}

	gasPrice, ok := new(big.Int).SetString(in.GasPrice, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("gas price:%s", in.GasPrice))
	}

	gaslimit, ok := new(big.Int).SetString(in.GasLimit, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf("gas limit:%s", in.GasLimit))
	}

	gasFeeCap, ok := new(big.Int).SetString(in.GasFeeCap, 10)
	if !ok {
		return nil, fmt.Errorf("invalid parameter:%s", fmt.Sprintf(" gas fee cap:%s", in.GasFeeCap))
	}

	/* 	streth, err := from.NewCommonAddr()
	   	if err != nil {
	   		logger.Info("NewCommonAddr error", zap.Error(err))
	   		return nil, fmt.Errorf("NewCommonAddr error:%s", err.Error())
	   	} */

	streth := from

	kaddr, err := g.Bc.GetBindingMetaAddress(streth.Hex())
	if err != nil {
		return nil, fmt.Errorf("GetBindingKtoAddress ethaddr[%v], error:%s", streth.Hex(), err.Error())
	}

	avb, err := g.Bc.GetAvailableBalance(kaddr)
	if err != nil {
		return nil, fmt.Errorf("GetAvailableBalance ethaddr[%v], error:%s", kaddr, err.Error())
	}

	// if in.GasLimit*in.GasPrice == 0 {
	// 	return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	// }

	// if in.Amount+in.GasLimit*in.GasPrice > avb {
	// 	return nil, fmt.Errorf("from(%v) balance(%v) is not enough or out of gas.", kaddr, avb)
	// }

	gas := new(big.Int).Mul(gasPrice, gaslimit)
	if gas.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("error: one of gasprice[%v] and gaslimit[%v] is 0", in.GasLimit, in.GasPrice)
	}

	if avb.Cmp(new(big.Int).Add(amount, gas)) == -1 {
		return nil, fmt.Errorf("from(%v) balance(%v) is not enough or out of gas.", from, avb)
	}

	nonce, err := g.Bc.GetNonce(&from)
	if err != nil {
		logger.Info("GetNonce error", zap.Error(err), zap.String("address", from.String()))
		return nil, fmt.Errorf("GetNonce error:%s,from:%v", err.Error(), from)
	}
	if nonce != in.Nonce {
		return nil, fmt.Errorf("error: from [%v] nonce[%v] not equal ethTx.Nonce[%v]", from, nonce, in.Nonce)
	}

	/* 	sign, err := crypto.DeserializeSignature(in.Sign)
	   	if err != nil {
	   		return nil, err
	   	} */
	tx := &transaction.SignedTransaction{
		Transaction: transaction.Transaction{
			From:      &from,
			To:        &to,
			Amount:    amount,
			Nonce:     in.Nonce,
			GasLimit:  gaslimit,
			GasPrice:  gasPrice,
			GasFeeCap: gasFeeCap,
			Input:     in.Input,
			Type:      transaction.WithdrawToEthTransaction,
		},
		Signature: in.Sign,
	}

	err = g.Tp.Add(tx)
	if err != nil {
		return nil, err
	}

	data, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	if g.Node != nil {
		g.Node.SendMessage(p2p.PayloadMessageType, append([]byte{0}, data...))
	}

	return &message.Res_WithdrawToEthaddr{Hash: hex.EncodeToString(tx.Hash())}, nil
}

/* func (g *Greeter) GetBasePledge(ctx context.Context, in *message.ReqBasePledge) (*message.RespBasePledge, error) {
	total, err := g.Bc.GetBasePledge()
	if err != nil {
		return nil, err
	}
	return &message.RespBasePledge{TotalNumber: total}, nil
} */
