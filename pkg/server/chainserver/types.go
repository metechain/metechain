package chainserver

import (
	"fmt"
	"math/big"

	"errors"

	"metachain/pkg/blockchain"

	"github.com/buaazp/fasthttprouter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ChainBlock struct {
	Height           uint64              `json:"height"`       //当前块号
	PrevHash         string              `json:"prevHash"`     //上一块的hash json:"prevBlockHash --> json:"prevHash
	Hash             string              `json:"hash"`         //当前块hash
	Transactions     []*ChainTransaction `json:"transactions"` //交易数据
	Root             string              `json:"root"`         //Txhash 默克根
	SnapRoot         string              `json:"snaproot"`     //默克根
	Version          uint64              `json:"version"`      //版本号
	Timestamp        uint64              `json:"timestamp"`    //时间戳
	UsedTime         uint64              `json:"usedtime"`
	Miner            string              `json:"miner"`            //矿工地址
	Difficulty       *big.Int            `json:"difficulty"`       //难度
	GlobalDifficulty *big.Int            `json:"globaldifficulty"` // 全网难度
	Nonce            uint64              `json:"nonce"`            //区块nonce
	GasLimit         uint64              `json:"gasLimit"`
	GasUsed          uint64              `json:"gasUsed"`
}

// Transaction
type ChainTransaction struct {
	Version uint64 `json:"Version"`
	Type    uint8  `json:"Type"`
	From    string `json:"From"`
	To      string `json:"To"`
	Amount  uint64 `json:"Amount"`
	Nonce   uint64 `json:"Nonce"`

	GasLimit  uint64 `json:"GasLimit"`
	GasFeeCap uint64 `json:"GasFeeCap"`
	GasPrice  uint64 `json:"GasPrice"`

	Input     string `json:"Input"`
	Signature string `json:"Signature"`

	GasUsed  uint64 `json:"GasUsed"`
	BlockNum uint64 `json:"BlockNum"`
}

type TransactionReceipt struct {
	BlockHash         common.Hash    `json:"blockHash"`
	BlockNumber       string         `json:"blockNumber"`
	ContractAddress   common.Address `json:"contractAddress"`
	CumulativeGasUsed string         `json:"cumulativeGasUsed"`
	From              common.Address `json:"from"`
	GasUsed           string         `json:"gasUsed"`
	Logs              []*types.Log   `json:"logs"`
	LogsBloom         types.Bloom    `json:"logsBloom"`
	Status            string         `json:"status"`
	To                common.Address `json:"to"`

	TransactionHash  common.Hash `json:"transactionHash"`
	TransactionIndex string      `json:"transactionIndex"`

	Root common.Hash `json:"root"`
}

type Server struct {
	address string
	bc      blockchain.Blockchains
	r       *fasthttprouter.Router
	grpcIp  string
}

type resultInfo struct {
	ErrorCode int    `json:"code"`
	ErrorMsg  string `json:"message"`
	Result    uint64 `json:"result"`
}

type resultString struct {
	ErrorCode int    `json:"code"`
	ErrorMsg  string `json:"message"`
	Result    string `json:"hash"`
}

type resultTransaction struct {
	ErrorCode   int               `json:"code"`
	ErrorMsg    string            `json:"message"`
	Transaction *ChainTransaction `json:"transaction"`
}

type resTransactionReceipt struct {
	ErrorCode   int                 `json:"code"`
	ErrorMsg    string              `json:"message"`
	Transaction *TransactionReceipt `json:"transaction"`
}

type resultBlock struct {
	ErrorCode int         `json:"code"`
	ErrorMsg  string      `json:"message"`
	Block     *ChainBlock `json:"block"`
}

type resultContract struct {
	ErrorCode int           `json:"code"`
	ErrorMsg  string        `json:"message"`
	Contract  *contractData `json:"contractdata"`
}

type contractData struct {
	Name        string   `json:"name"`
	Symbol      string   `json:"symbol"`
	Decimal     *big.Int `json:"decimal"`
	TotalSupply *big.Int `json:"totalSupply"`
}

var (
	NAME        string = "0x06fdde03"
	SYMBOL      string = "0x95d89b41"
	DECIMALS    string = "0x313ce567"
	TOTALSUPPLY string = "0x18160ddd"

	BalanceOfPerfix = "0x70a08231000000000000000000000000"
)

type resultPledgeInfo struct {
	ErrorCode      int    `json:"code"`
	ErrorMsg       string `json:"message"`
	TotalPledge    uint64 `json:"totalPledge"`
	TotalMined     uint64 `json:"totalMined"`
	WholeNetPledge uint64 `json:"wholeNetPledge"`
}

type resultBalance struct {
	ErrorCode int      `json:"code"`
	ErrorMsg  string   `json:"message"`
	Balance   *big.Int `json:"balance"`
}

var (
	Success          = 0
	ErrJSON          = -41201
	ErrNoTransaction = -41203
	ErrNoTxByHash    = -41204
	ErrData          = -41205
	ErrNoBlock       = -41206
	ErrtTx           = -41207
	ErrNoBlockHeight = -41208
)

func getString(mp map[string]interface{}, k string) (string, error) {
	v, ok := mp[k]
	if !ok {
		return "", errors.New(fmt.Sprintf("'%s' not exist", k))
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", errors.New(fmt.Sprintf("'%s' not string", k))
}

func uint64ToHexString(val uint64) string {
	return stringToHex(fmt.Sprintf("%X", val))
}

func stringToHex(s string) string {
	return "0x" + s
}
