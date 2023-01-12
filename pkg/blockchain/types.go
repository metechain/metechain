package blockchain

import (
	"crypto/sha1"
	"math/big"
	"sync"

	"metachain/pkg/contract/evm"
	"metachain/pkg/storage/store"
	"metachain/pkg/transaction"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
)

const (
	MAXUINT64 = ^uint64(0)
)

var (
	// SnapRootKey key to store snaproot in database
	SnapRootKey = []byte("snapRoot")
	// HeightKey key to store height in database
	HeightKey = []byte("height")
	// FreezeKey store the map name of freeze balance
	FreezeKey = []byte("freeze")
)

var (
	// SnapRootPrefix prefix of block snapRoot
	SnapRootPrefix = []byte("blockSnap")
	// HeightPrefix prefix of block height key
	HeightPrefix = []byte("blockheight")

	BindingKey     = []byte("binding")
	CREATECONTRACT = "create"
	CALLCONTRACT   = "call"
)

// Blockchain blockchain data structure
type Blockchain struct {
	mu sync.RWMutex
	db store.DB
	//	cdb      *bgdb.Database
	sdb      *state.StateDB
	evm      *evm.Evm
	ChainCfg *ChainConfig
}

var ETHDECIMAL uint64 = 10000000

const MAXGASLIMIT uint64 = 10000000000000
const MINGASLIMIT uint64 = 2100000
const CONTRACTMINGASLIMIT uint64 = 21000

type limit struct {
	maxGas         *big.Int
	minGas         *big.Int
	contractMinGas *big.Int
}

var Limit *limit = &limit{
	maxGas:         new(big.Int).Mul(big.NewInt(100), transaction.UnitPrecision),
	minGas:         new(big.Int).Mul(big.NewInt(2100000), big.NewInt(10000000)),
	contractMinGas: new(big.Int).Mul(big.NewInt(21000), big.NewInt(10000000)),
}

func (l *limit) MaxGasLimit() *big.Int {
	return new(big.Int).Set(l.maxGas)
}
func (l *limit) MinGasLimit() *big.Int {
	return new(big.Int).Set(l.minGas)
}
func (l *limit) ContractMinGasLimit() *big.Int {
	return new(big.Int).Set(l.contractMinGas)
}

// TxIndex transaction data index structure
type TxIndex struct {
	Height uint64
	Index  uint64
}

const (
	PledgeCycle    = 120                          //pledge cydle (days)
	DayMinedBlocks = 7500                         //total blocks per day online
	CycleMax       = PledgeCycle * DayMinedBlocks //total blocks per pledge cycle
)

const (
	DestroyRatio = 20  // pledge default ratio
	MINEDDIVID   = 70  //dividend
	DIVIDEND     = 100 //divid
)

func Check0x(input string) string {
	if len(input) > 2 {
		if input[:2] == "0x" {
			input = input[2:]
		}
	}
	return input
}

//get pledge keys info
func getPledgeKVsInfo(DBTransaction store.Transaction, prefix []byte) ([][]byte, [][]byte, error) {
	return DBTransaction.Mkvs(prefix)
}

func commonAddrToStoreAddr(caddr common.Address, prefix []byte) common.Address {
	hash := sha1.Sum(append(prefix, caddr.Bytes()...))
	caddr.SetBytes(hash[:])
	return caddr
}

func CommonAddrToStoreAddr(caddr common.Address, prefix []byte) common.Address {
	hash := sha1.Sum(append(prefix, caddr.Bytes()...))
	caddr.SetBytes(hash[:])
	return caddr
}

func Uint64ToBigInt(v uint64) *big.Int {
	return new(big.Int).SetUint64(v)
}

func GetMinerAmount(height uint64) *big.Int {

	return new(big.Int).Mul(big.NewInt(50000), transaction.UnitPrecision)
}
