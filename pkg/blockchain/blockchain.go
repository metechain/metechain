package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"metachain/pkg/contract/evm"
	"metachain/pkg/txpool"

	"github.com/cockroachdb/pebble"

	"metachain/pkg/block"
	"metachain/pkg/logger"
	"metachain/pkg/storage/merkle"
	"metachain/pkg/storage/miscellaneous"
	"metachain/pkg/storage/store"
	"metachain/pkg/storage/store/bg/bgdb"
	"metachain/pkg/transaction"
	"metachain/pkg/util/difficulty"
	diffhash "metachain/pkg/util/difficulty/hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"go.uber.org/zap"
)

var (
	BlockHeight    uint64
	NotExistHeight uint64
	BlockNotExist  = "BLOCKNOTEXIST"
)

// 从1开始挖矿
const InitHeight = 0

type ChainConfig struct {
	ChainId   int64  `yaml:"chainid"`
	NetworkId int64  `yaml:"networkId"`
	GasLimit  uint64 `yaml:"gaslimit"`
	GasPrice  uint64 `yaml:"gasprice"`
	Miner     *common.Address
}

var (
	_ Blockchains        = (*Blockchain)(nil)
	_ txpool.IBlockchain = (*Blockchain)(nil)
)

// New create blockchain object
func New(bgs store.DB, cfg *ChainConfig) (*Blockchain, error) {
	//bgs := bg.New(db)
	cdb := bgdb.NewBadgerDatabase(bgs)
	sdb := state.NewDatabase(cdb)

	bc := &Blockchain{db: bgs, ChainCfg: cfg}

	if err := bc.fallbackOneBleck(); err != nil {
		return nil, fmt.Errorf("fallbackOneBleck:%w", err)
	}

	root, err := getSnapRootLock(bgs)
	if err != nil {
		logger.Error("failed to getSnapRoot")
		return nil, err
	}

	stdb, err := state.New(root, sdb, nil)
	if err != nil {
		logger.Error("failed to new state", zap.Error(err))
		return nil, err
	}

	bc.sdb = stdb

	bc.evm = evm.NewEvm(bc.sdb, cfg.ChainId, cfg.GasLimit, new(big.Int).SetUint64(cfg.GasPrice))

	// 初始化创世块并添加
	a, err := bc.GetMaxBlockHeight()
	if err != nil {
		panic(err)
	}
	if a == 0 {
		b := block.NewGenesisBlock(*cfg.Miner)
		return bc, bc.AddBlock(b)
	}

	return bc, nil
}

func (bc *Blockchain) fallbackOneBleck() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	h, err := bc.getMaxBlockHeight()
	if errors.Is(err, pebble.ErrNotFound) {
		return nil
	}

	if err != nil {
		return err
	}

	if h > 0 {
		return bc.fallbackBlock(h - 1)
	}
	return nil
}

func (bc *Blockchain) fallbackBlock(h uint64) error {
	//获取快照
	b, err := bc.getBlockByHeight(h)
	if err != nil {
		return err
	}
	var evmsnaphot common.Hash

	copy(evmsnaphot[:], b.SnapRoot[:common.HashLength])
	sdb, err := updateNewStateByRoot(bc, evmsnaphot)
	if err != nil {
		return err
	}

	txn := bc.db.NewTransaction()
	defer txn.Commit()

	err = txn.Set(HeightKey, miscellaneous.E64func(h))
	if err != nil {
		return err
	}

	err = txn.Set(SnapRootKey, evmsnaphot[:])
	if err != nil {
		bc.sdb = sdb
	}
	return err

}

func (bc *Blockchain) Close() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.db.Close()
}

func (bc *Blockchain) SyncLock() {
	bc.mu.Lock()
}

func (bc *Blockchain) SyncUnlock() {
	bc.mu.Unlock()
}

//Get  blockchain top
func (bc *Blockchain) Tip() (*block.Block, error) {

	bc.mu.Lock()
	defer bc.mu.Unlock()
	h, err := bc.getMaxBlockHeight()
	if err != nil {
		return nil, err
	}

	hash, err := bc.getHash(h)
	if err != nil {
		return nil, err
	}

	blockData, err := bc.db.Get(hash)
	if err != nil {
		return nil, err
	}

	return block.Deserialize(blockData)
}

func (bc *Blockchain) FindChainBranch(Hash []byte) (*block.Block, error) {

	hash := Hash

	for {
		ismainchain, err := bc.IsMainChainBlock(hash)
		if err != nil {
			return nil, err
		}

		if ismainchain {
			return bc.getBlockByHash(hash)
		}
		block, err := bc.getBlockByHash(hash)
		if err != nil {
			return nil, err
		}
		hash = block.PrevHash
	}

}

//
func (bc *Blockchain) IsMainChainBlock(hash []byte) (bool, error) {

	block, err := bc.getBlockByHash(hash)
	if err != nil {
		return false, err
	}
	mainChainHash, err := bc.GetHash(block.Height)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, mainChainHash), nil

}

// GetAvailableBalance get available balance of address
func (bc *Blockchain) GetAvailableBalance(address *common.Address) (*big.Int, error) {
	//	bc.mu.RLock()
	//	defer bc.mu.RUnlock()
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.getAvailableBalance(address)
}

// getAvailableBalance get available balance
func (bc *Blockchain) getAvailableBalance(address *common.Address) (*big.Int, error) {
	balance, err := bc.getBalance(*address)
	if err != nil {
		logger.Error("get balance", zap.Error(err))
		return nil, err
	}

	// feeBalance, err := bc.getAllFreezeBalance(address)
	// if err != nil {
	// 	logger.Error("get lock balance", zap.Error(err))
	// 	return nil, err
	// }
	// totalMined, err := bc.getTotalMined(address)
	// if err != nil {
	// 	logger.Error("get total mined", zap.Error(err))
	// 	return nil, err
	// }

	// TODO:需要删掉
	// totalPledge, err := getTotalPledge(bc.sdb, address)
	// if err != nil {
	// 	logger.Error("get total pledge", zap.Error(err))
	// 	return nil, err
	// }

	// aviBalance := balance - feeBalance - totalMined- totalPledge

	// if MAXUINT64-aviBalance < feeBalance+totalMined+totalPledge {
	// 	return 0, fmt.Errorf("available balance is error, MAXUINT64(%d) - aviBalance(%d) < feeBalance(%d)-totalMined(%d) < totalPledge(%d)", MAXUINT64, aviBalance, feeBalance, totalMined, totalPledge)
	// }

	return new(big.Int).Set(balance), nil
}

// GetHash get the hash corresponding to the block height
func (bc *Blockchain) GetHash(height uint64) (hash []byte, err error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return bc.db.Get(append(HeightPrefix, miscellaneous.E64func(height)...))
}

func (bc *Blockchain) getHash(height uint64) (hash []byte, err error) {

	return bc.db.Get(append(HeightPrefix, miscellaneous.E64func(height)...))
}

// GetMaxBlockHeight get maximum block height
func (bc *Blockchain) GetMaxBlockHeight() (height uint64, err error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.getMaxBlockHeight()
}

// getMaxBlockHeight get maximum block height
func (bc *Blockchain) getMaxBlockHeight() (uint64, error) {
	heightBytes, err := bc.db.Get(HeightKey)
	if err == pebble.ErrNotFound {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return miscellaneous.D64func(heightBytes)
}

func getMaxBlockHeight(DBTransaction store.Transaction) (uint64, error) {
	heightBytes, err := DBTransaction.Get(HeightKey)
	if err == store.NotExist {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return miscellaneous.D64func(heightBytes)
}

// GetBlockByHeight get the block corresponding to the block height
func (bc *Blockchain) GetBlockByHeight(height uint64) (*block.Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.getBlockByHeight(height)
}

// getBlockByHeight get the block corresponding to the block height
func (bc *Blockchain) getBlockByHeight(height uint64) (*block.Block, error) {
	if height < 1 {
		return nil, errors.New("parameter error")
	}
	//Get the hash first
	hash, err := bc.db.Get(append(HeightPrefix, miscellaneous.E64func(height)...))
	if err != nil {
		return nil, err
	}
	//Then get the block through hash
	blockData, err := bc.db.Get(hash)
	if err != nil {
		return nil, err
	}

	return block.Deserialize(blockData)
}

func getBlockByHeight(height uint64, tx store.Transaction) (*block.Block, error) {
	if height < 1 {
		return nil, errors.New("parameter error")
	}
	//Get the hash first
	hash, err := tx.Get(append(HeightPrefix, miscellaneous.E64func(height)...))
	if err != nil {
		return nil, err
	}
	//Then get the block through hash
	blockData, err := tx.Get(hash)
	if err != nil {
		return nil, err
	}

	return block.Deserialize(blockData)
}

// GetBlockByHeight get the block corresponding to the block height
func (bc *Blockchain) GetBlockByHash(hash []byte) (*block.Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.getBlockByHash(hash)
}

// getBlockByHeight get the block corresponding to the block height
func (bc *Blockchain) getBlockByHash(hash []byte) (*block.Block, error) {
	blockData, err := bc.db.Get(hash)
	if err != nil {
		return nil, err
	}

	return block.Deserialize(blockData)
}

// GetTransactionByHash get the transaction corresponding to the transaction hash
func (bc *Blockchain) GetTransactionByHash(hash []byte) (*transaction.FinishedTransaction, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.getTransactionByHash(hash)
}

// getTransactionByHash get the transaction corresponding to the transaction hash
func (bc *Blockchain) getTransactionByHash(hash []byte) (*transaction.FinishedTransaction, error) {
	Hi, err := bc.db.Get(hash)
	if err != nil {
		//logger.Error("failed to get hash", zap.Error(err))
		return nil, err
	}
	var txindex TxIndex
	err = json.Unmarshal(Hi, &txindex)
	if err != nil {
		logger.Error("Failed to unmarshal bytes", zap.Error(err))
		return nil, err
	}
	b, err := bc.getBlockByHeight(txindex.Height)
	if err != nil {
		logger.Error("failed to getblock height", zap.Error(err), zap.Uint64("height", txindex.Height))
		return nil, err
	}

	b.Transactions[txindex.Index].BlockNum = b.Height
	tx := &b.Transactions[txindex.Index]

	return *tx, nil
}

// NewBlock create a new block for the blockchain
func (bc *Blockchain) NewBlock(txs []*transaction.SignedTransaction, minaddr *common.Address) (*block.Block, error) {
	//logger.Info("start to new block")
	var height, prevHeight uint64
	var prevHash []byte
	var gasUsed *big.Int
	prevHeight, err := bc.GetMaxBlockHeight()
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		return nil, err
	}

	height = prevHeight + 1
	if height > InitHeight {
		prevHash, err = bc.GetHash(prevHeight)
		if err != nil {
			logger.Error("failed to get hash", zap.Error(err), zap.Uint64("previous height", prevHeight))
			return nil, err
		}
	} else {
		prevHash = block.GenesisHash
	}

	// init test (仅限测试使用)
	// if height == 1 {
	// 	setInitTest(bc)
	// 	txs = distrInitTest(txs)
	// 	txs = distrInitWaterTest(txs)
	// 	txs = distrInitWaterTest1(txs)
	// 	txs = distrInitWaterTest2(txs)
	// 	txs = distrInitWaterTest3(txs)
	// 	txs = distrInitWaterTest4(txs)
	// }

	// Currency distribution
	txs = distr(txs, minaddr, height)

	// Generate Merkel root, if there is no deal, calling GetMthash will painc
	txBytesList := make([][]byte, 0, len(txs))
	for i, tx := range txs {
		serialize, err := tx.Serialize()
		if err != nil {
			logger.Error("tx.serialize", zap.Error(err), zap.Int("index", i))
			return nil, err
		}
		txBytesList = append(txBytesList, serialize)
		// gasUsed += tx.Transaction.GasLimit * tx.Transaction.GasPrice
		gasUsed = new(big.Int).Add(tx.GasLimit, tx.GasPrice)
	}
	tree := merkle.New(sha256.New(), txBytesList)
	root := tree.GetMtHash()

	getRoot, err := getSnapRootLock(bc.db)
	if err != nil {
		logger.Error("failed to get getSnapRootLock", zap.Error(err))
		return nil, err

	}
	snapRoot := getRoot.Bytes()

	ftxs := make([]*transaction.FinishedTransaction, len(txs))
	for i, _ := range ftxs {
		ftxs[i] = transaction.NewFinishedTransaction(txs[i], gasUsed, height)
		ftxs[i].BlockNum = height
	}

	timestamp := uint64(time.Now().Unix())
	block := &block.Block{
		Height:       height,
		PrevHash:     prevHash,
		Transactions: ftxs,
		Root:         root,
		Version:      1,
		Timestamp:    timestamp,
		UsedTime:     0,
		Miner:        minaddr,
		SnapRoot:     snapRoot,
		//Difficulty:       big.NewInt(0),
		GlobalDifficulty: big.NewInt(0),
		Nonce:            1,
		GasLimit:         bc.ChainCfg.GasLimit,
		GasUsed:          gasUsed,
	}
	//block.SetHash()
	//	logger.Info("end to new block")
	return block, nil
}

// AddUncleBlock add uncle blocks to blockchain
func (bc *Blockchain) AddUncleBlock(block *block.Block) error {
	logger.Info("addUncleBlock", zap.Uint64("blockHeight", block.Height), zap.String("hash", hex.EncodeToString(block.Hash)))
	bc.mu.Lock()
	defer bc.mu.Unlock()

	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()

	// hash -> block
	data, err := block.Serialize()
	if err != nil {
		logger.Error("failed serialize block", zap.Error(err))
		return err
	}
	if err := DBTransaction.Set(block.Hash, data); err != nil {
		logger.Error("Failed to set block", zap.Error(err))
		return err
	}

	if err := DBTransaction.Commit(); err != nil {
		logger.Error("filed to commit db transaction", zap.Error(err))
	}

	logger.Info("End adduncleBlock", zap.Uint64("blockHeight", block.Height))
	return nil
}

// DeleteUncleBlock delete uncle blocks to blockchain
func (bc *Blockchain) DeleteUncleBlock(block *block.Block) error {
	logger.Info("DeleteUncleBlock", zap.Uint64("blockHeight", block.Height), zap.String("hash", hex.EncodeToString(block.Hash)))
	bc.mu.Lock()
	defer bc.mu.Unlock()

	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()

	// hash -> block
	/* 	data, err := block.Serialize()
	   	if err != nil {
	   		logger.Error("failed serialize block", zap.Error(err))
	   		return err
	   	} */
	if err := DBTransaction.Del(block.Hash); err != nil {
		logger.Error("Failed to set block", zap.Error(err))
		return err
	}

	if err := DBTransaction.Commit(); err != nil {
		logger.Error("filed to commit db transaction", zap.Error(err))
	}

	logger.Info("End DeleteUncleBlock", zap.Uint64("blockHeight", block.Height))
	return nil
}

func CompactToBig(compact uint32) *big.Int {
	mantissa := compact & 0x007fffff
	//  0010 0000 0111 1111 1111 1111 1111 1111
	//  0000 0000 0111 1111 1111 111 1111 1111
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)
	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}
	if isNegative {
		bn = bn.Neg(bn)
	}
	return bn
}

func BigToCompact(n *big.Int) uint32 {
	if n.Sign() == 0 {
		return 0
	}
	var mantissa uint32
	exponent := uint(len(n.Bytes()))
	if exponent <= 3 {
		mantissa = uint32(n.Bits()[0])
		mantissa <<= 8 * (3 - exponent)
	} else {
		tn := new(big.Int).Set(n)
		mantissa = uint32(tn.Rsh(tn, 8*(exponent-3)).Bits()[0])
	}
	if mantissa&0x00800000 != 0 {
		mantissa >>= 8
		exponent++
	}
	compact := uint32(exponent<<24) | mantissa
	if n.Sign() < 0 {
		compact |= 0x00800000
	}
	return compact
}

func (bc *Blockchain) rollState(rollroot common.Hash) error {
	sdb, err := updateNewStateByRoot(bc, rollroot)
	if err != nil {
		logger.Error("Failed to rollState", zap.Error(err))
		return err
	}
	bc.sdb = sdb
	bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))
	return nil
}

func (bc *Blockchain) AddLoseBlock1(blk *block.Block) error {

	data, err := blk.Serialize()
	if err != nil {
		logger.Error("Deserialize err", zap.Error(err))
		return err

	}

	DBTransaction := bc.NewTransaction()
	defer DBTransaction.Cancel()

	for index, tx := range blk.Transactions {
		logger.Info("block :", zap.String("hash", tx.HashToString()), zap.String("tx", tx.String()))
		if tx.Transaction.IsCoinBaseTransaction() {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("amount", tx.Amount.String()))
				return err
			}

			// } else if tx.Transaction.IsLockTransaction() || tx.Transaction.IsUnlockTransaction() {
			// 	txHash := tx.Hash()
			// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
			// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))

			// 		return err
			// 	}

			// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
			// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
			// 			zap.String("to address", tx.Transaction.To.String()), zap.Uint64("amount", tx.Transaction.Amount))

			// 		return err
			// 	}

			// 	// } else if tx.Transaction.IsPledgeTrasnaction() || tx.IsPledgeBreakTransaction() {

			// 	// 	txHash := tx.Hash()
			// 	// 	if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
			// 	// 		logger.Error("Failed to set transaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))

			// 	// 		return err
			// 	// 	}

		} else if tx.Transaction.IsEvmContractTransaction() {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))

				return err
			}

		} else {

			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))

				return err
			}

			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, blk.Height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))

				return err
			}

		}
	}

	if err := DBTransaction.Set(append(HeightPrefix, miscellaneous.E64func(blk.Height)...), blk.Hash); err != nil {
		logger.Error("Failed to set height and hash", zap.Error(err))

		return err
	}
	if err := DBTransaction.Set(blk.Hash, data); err != nil {
		logger.Error("Failed to set block", zap.Error(err))
		return err
	}

	if err := DBTransaction.Commit(); err != nil {
		logger.Error("commit db", zap.Error(err), zap.Uint64("block number", blk.Height))
		return err
	}
	logger.Info("commit lose block", zap.Uint64("height=", blk.Height))
	return nil
}

// AddBlock add blocks to blockchain
func (bc *Blockchain) AddBlock(block *block.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	rollroot, err := getSnapRoot(bc.db)
	if err != nil {
		logger.Error("AddBlock Failed, getSnapRoot", zap.Error(err))
		return err
	}
	var REVERT error = nil
	defer func() {
		BlockHeight = block.Height
		logger.SugarLogger.Info(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		logger.SugarLogger.Infof("block.Height:%d", block.Height)
		logger.SugarLogger.Infof("nonce:%d", block.Nonce)
		logger.SugarLogger.Infof("globalDifficulty:%d", BigToCompact(block.GlobalDifficulty))
		logger.SugarLogger.Infof("hash:%s", hex.EncodeToString(block.Hash))
		logger.SugarLogger.Infof("PrevHash:%s", hex.EncodeToString(block.PrevHash))
		logger.SugarLogger.Info("<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
		logger.InfoLogger.Printf(" Add block success    block Height=[%d]  hash =[%s]\n\n", block.Height, hex.EncodeToString(block.Hash))
		if REVERT != nil {
			logger.Error("AddBlock Failed,start to revert...", zap.Error(REVERT))
			err := bc.rollState(rollroot)
			if err != nil {
				logger.Error("rollState Failed", zap.Error(err))
			}
			return
		}
	}()
	// logger.Info("addBlock", zap.Uint64("blockHeight", block.Height))

	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()
	var height, prevHeight uint64
	// take out the block height
	prevHeight, err = bc.getMaxBlockHeight()
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		REVERT = err
		return err
	}

	/* if block.Height > InitHeight+1 {
		SnapRoothash, err := DBTransaction.Get(SnapRootKey)
		if err != nil {
			REVERT = err
			return err
		}
		startHash, err := factCommit(bc.sdb, true)
		if err != nil {
			logger.Error("Failed to set factCommit", zap.Error(err))
			REVERT = err
			return err
		}

		if !bytes.Equal(SnapRoothash, startHash.Bytes()) {
			logger.Error("snaproot is changed", zap.String("SnapRootkey-hash", hex.EncodeToString(SnapRoothash)),
				zap.String("nowsnaproothash", hex.EncodeToString(startHash.Bytes())))
			REVERT = fmt.Errorf("snaproot not equal,old root hash:%v,current root hash:%v", hex.EncodeToString(SnapRoothash), startHash)
			return REVERT
		}
	} */

	if prevHeight != 0 || block.Height != 0 {
		height = prevHeight + 1
		if block.Height != height {
			REVERT = fmt.Errorf("height error:current height=%d,commit height=%d", prevHeight, block.Height)
			return REVERT
		}
	}

	// height -> hash
	hash := block.Hash
	if err := DBTransaction.Set(append(HeightPrefix, miscellaneous.E64func(height)...), hash); err != nil {
		logger.Error("Failed to set height and hash", zap.Error(err))
		REVERT = err
		return err
	}

	// reset block height
	if err := DBTransaction.Set(HeightKey, miscellaneous.E64func(height)); err != nil {
		REVERT = fmt.Errorf("set hetght:%w", err)
		return err
	}

	logger.Info("set height", zap.Uint64("height", height))

	{
		//must set block into into evm at every addblock
		bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))
		miner := *block.Miner
		bc.evm.SetBlockInfo(block.Height, block.Timestamp, miner, block.GlobalDifficulty)
	}

	var blockGasU = new(big.Int)
	for index, tx := range block.Transactions {
		logger.Info("block :", zap.String("hash", tx.HashToString()), zap.String("tx", tx.String()))
		if tx.Transaction.IsCoinBaseTransaction() {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}

			tx.GasUsed = new(big.Int).Mul(tx.GasLimit, tx.GasPrice)
			blockGasU.Add(blockGasU, tx.GasUsed)
			if err := setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("from address", block.Miner.String()), zap.String("fee", tx.GasUsed.String()))
				REVERT = err
				return err
			}

			if err := bc.setToAccount(block, &tx.Transaction); err != nil {
				logger.Error("Failed to set account", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}
		} else if tx.Transaction.IsEvmContractTransaction() {
			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
				REVERT = err
				return err
			}

			gasLeft, err := bc.handleContractTransaction(block, DBTransaction, tx, index)
			if err != nil {
				logger.Error("Failed to HandleContractTransaction", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)))
				REVERT = err
				return err
			}

			evmcfg := bc.evm.GetConfig()
			// if evmcfg.GasLimit < gasLeft {
			if gasLeft.Cmp(big.NewInt(int64(evmcfg.GasLimit))) >= 0 {
				logger.Error("Failed to HandleContractTransaction", zap.Error(fmt.Errorf("hash[%v],evm gaslimit[%v] < gasLeft[%v]", transaction.HashToString(txHash), evmcfg.GasLimit, gasLeft)))
				REVERT = fmt.Errorf("error: hash[%v],evm gaslimit[%v] < gasLeft[%v]", transaction.HashToString(txHash), evmcfg.GasLimit, gasLeft)
				return REVERT
			}

			// tx.GasUsed = evmcfg.GasLimit - gasLeft
			// blockGasU += tx.GasUsed

			tx.GasUsed = new(big.Int).Sub(big.NewInt(int64(evmcfg.GasLimit)), gasLeft)

			if err := setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("hash", transaction.HashToString(txHash)), zap.String("gasUsed", tx.GasUsed.String()))
				REVERT = err
				return err
			}

			// update balance
			if err := setAccount(bc, tx); err != nil {
				logger.Error("Failed to set balance", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}
		} else {
			/*if tx.Transaction.Input != nil || !bytes.Equal(tx.Transaction.Input, []byte("")) || len(tx.Transaction.Input) != 0 {
				ipt := string(tx.Input)
				sc := parser.Parser([]byte(tx.Input))
				coAddr, _ := tx.Transaction.From.NewCommonAddr()
				e, err := exec.New(DBTransaction, bc.sdb, sc, coAddr.String())
				if err != nil {
					logger.Error("Failed to new exec", zap.String("script", ipt),
						zap.String("from address", tx.From.String()))
					ipt += " InsufficientBalance"
					tx.Input = []byte(ipt)
					block.Transactions[index] = tx
					REVERT = err
					return err
				}

				if e != nil {
					if err := e.Flush(DBTransaction, bc.sdb); err != nil {
						logger.Error("Failed to flush exec", zap.String("script", ipt),
							zap.String("from address", tx.From.String()))
						ipt += " ExecutionFailed"
						tx.Input = []byte(ipt)
						block.Transactions[index] = tx
						REVERT = err
						return err
					}
				}
			}*/

			txHash := tx.Hash()
			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.From.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}

			if err := setTxbyaddrKV(DBTransaction, tx.Transaction.To.Bytes(), txHash, height, uint64(index)); err != nil {
				logger.Error("Failed to set transaction", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}
			// update nonce,txs in block must be ordered
			nonce := tx.Transaction.Nonce + 1
			if err := setNonce(bc.sdb, *tx.Transaction.From, nonce); err != nil {
				logger.Error("Failed to set nonce", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}

			//err

			// tx.GasUsed = tx.Transaction.GasLimit * tx.Transaction.GasPrice
			// blockGasU += tx.GasUsed
			tx.GasUsed = new(big.Int).Mul(tx.GasLimit, tx.GasPrice)
			blockGasU.Add(blockGasU, tx.GasUsed)

			if err := setMinerFee(bc, *block.Miner, tx.GasUsed); err != nil {
				logger.Error("Failed to set Minerfee", zap.Error(err), zap.String("from address", block.Miner.String()), zap.String("fee", tx.GasUsed.String()))
				REVERT = err
				return err
			}

			// update balance
			if err := setAccount(bc, tx); err != nil {
				logger.Error("Failed to set balance", zap.Error(err), zap.String("from address", tx.Transaction.From.String()),
					zap.String("to address", tx.Transaction.To.String()), zap.String("amount", tx.Amount.String()))
				REVERT = err
				return err
			}
		}
	}

	t0 := time.Now()
	comHash, err := factCommit(bc.sdb, true)
	if err != nil {
		logger.Error("Failed to set factCommit", zap.Error(err))
		REVERT = err
		return err
	}
	logger.Info("sub factcommit", zap.Float64("second", time.Since(t0).Seconds()))

	/* 	if block.Height > InitHeight+1 {
	   		oldSnapRootkey, err := DBTransaction.Get(SnapRootKey)
	   		if err != nil {
	   			REVERT = err
	   			return err
	   		}

	   		logger.Info("AddBlock", zap.String("oldSnapRootkey", hex.EncodeToString(oldSnapRootkey)))
	   		logger.Info("AddBlock", zap.String("b.SnapRootkey", hex.EncodeToString(block.SnapRoot)))
	   		logger.Info("AddBlock", zap.String("newSnapRootkey", hex.EncodeToString(comHash.Bytes())))

	   		if bytes.Equal(oldSnapRootkey, block.SnapRoot) {
	   			block.SnapRoot = comHash.Bytes()
	   		} else if !bytes.Equal(comHash.Bytes(), block.SnapRoot) {
	   			REVERT = fmt.Errorf("SnapRoot not equal")
	   			return REVERT
	   		}
	   	} else if block.Height == InitHeight+1 {
	   		block.SnapRoot = comHash.Bytes()
	   	} else {
	   		REVERT = fmt.Errorf("block.height[%d]<initheight[%d]", block.Height, InitHeight)
	   		return REVERT
	   	} */

	if block.Height == 0 {
		block.SnapRoot = comHash.Bytes()
	} else {
		oldSnapRootkey, err := DBTransaction.Get(SnapRootKey)
		if err != nil {
			REVERT = err
			return err
		}

		logger.Info("AddBlock", zap.String("oldSnapRootkey", hex.EncodeToString(oldSnapRootkey)))
		logger.Info("AddBlock", zap.String("b.SnapRootkey", hex.EncodeToString(block.SnapRoot)))
		logger.Info("AddBlock", zap.String("newSnapRootkey", hex.EncodeToString(comHash.Bytes())))

		if bytes.Equal(oldSnapRootkey, block.SnapRoot) {
			block.SnapRoot = comHash.Bytes()
		} else if !bytes.Equal(comHash.Bytes(), block.SnapRoot) {
			REVERT = fmt.Errorf("SnapRoot not equal")
			return REVERT
		}
	}

	// hash -> block
	block.GasUsed = blockGasU
	data, err := block.Serialize()
	if err != nil {
		logger.Error("failed serialize block", zap.Error(err))
		REVERT = err
		return err
	}

	if err := DBTransaction.Set(hash, data); err != nil {
		logger.Error("Failed to set block", zap.Error(err))
		REVERT = err
		return err
	}

	if err := DBTransaction.Set(append(SnapRootPrefix, miscellaneous.E64func(height)...), comHash.Bytes()); err != nil {
		logger.Error("Failed to set height and hash", zap.Error(err))
		REVERT = err
		return err
	}

	DBTransaction.Set(SnapRootKey, comHash.Bytes())

	if err := DBTransaction.Commit(); err != nil {
		logger.Error("commit db", zap.Error(err), zap.Uint64("block number", block.Height))
		REVERT = err
		return err
	}

	/* 	time.Sleep(200 * time.Millisecond) */
	return nil
}

//ReorganizeChain reorganizes the block chain by disconnecting the
//nodes in the main list and connecting the nodes in the branch  list
//len(hashs) means main list need recover block lenght include tip  block,

func (bc *Blockchain) ReorganizeChain(hashs [][]byte, delHeight uint64) error {

	var errf int

	bc.mu.Lock()
	defer bc.mu.Unlock()

	db := bc.NewTransaction()
	defer db.Cancel()
	//snapshot
	root, err := getSnapRoot(bc.db)
	if err != nil {
		logger.Error("getSnapRoot   err", zap.Error(err))
		return err
	}

	defer func() {
		if errf != 0 {
			sdb, err := updateNewStateByRoot(bc, root)
			if err != nil {
				logger.Error("updateNewState   err", zap.Error(err))
				return
			}
			bc.sdb = sdb
			bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))
		}
	}()

	//len(hashs) number of blocks to be rolled back
	num := len(hashs)
	err = bc.DeleteTempBlockTest(delHeight, db)
	if err != nil {
		errf = -1
		logger.Error("DeleteTempBlock   err", zap.Error(err))
		return err
	}

	for num > 0 {
		num--
		block, err := bc.getBlockByHash(hashs[num])
		if err != nil {
			errf = -1
			return err
		}

		if err := bc.checkBlockRegular(block, bc.db, db); err != nil {
			errf = -1
			return err
		}
		// if err := difficultDetection(block, bc.db, db); err != nil {
		// 	errf = -1
		// 	return err
		// }

		err = bc.AddTempBlock(block, db)
		if err != nil {
			errf = -1
			logger.Error("ReorganizeChain.AddTempBlock err", zap.Error(err))
			return err
		}
	}

	if err := db.Commit(); err != nil {
		errf = -1
		logger.Error("commmit  err", zap.Error(err))
		return err
	}

	return nil
}

func updateDifficulty(height uint64, coinbaseAddr *common.Address, tx store.Transaction) (*big.Int, error) {
	h, err := getMaxBlockHeight(tx)
	if err != nil {
		return nil, err
	}

	b, err := getBlockByHeight(h, tx)
	if err != nil {
		return nil, err
	}

	if b.Height != height {
		return nil, fmt.Errorf("height error,last heighet: %d,incoming heighet: %d", b.Height, height)
	}

	gd := b.GlobalDifficulty

	// 0.计算下一个块的global难度
	// var head, tail *block.Block
	if height != 1 && height%10 == 1 {

		subTime := uint64(0)
		for i := uint64(0); i < 10; i++ {
			tmp, err := getBlockByHeight(height-i, tx)
			if err != nil {
				return nil, err
			}

			subTime += tmp.UsedTime
		}

		oldGlobalDifficultyBits := BigToCompact(b.GlobalDifficulty)
		newGlobalDifficultyBits := difficulty.CalcNextGlobalRequiredDifficulty(int64(0), int64(subTime), oldGlobalDifficultyBits)
		gd = CompactToBig(newGlobalDifficultyBits)

		logger.Info("check Difficulty", zap.Uint64("sub time", subTime), zap.Uint32("oldGlobalDifficultyBits", oldGlobalDifficultyBits), zap.Uint32("newGlobalDifficultyBits", newGlobalDifficultyBits))
	}

	return gd, nil
}

// func (bc *Blockchain) getBasePledge(sdb *state.StateDB, tx store.Transaction) (uint64, error) {
// 	total, pledgeTotal, err := getWholeNetworkPledgeAndNum(sdb, tx)
// 	if err != nil {
// 		return 0, err
// 	}

// 	bigTotal := big.NewInt(0).SetUint64(total)
// 	bigPledgeTotal := big.NewInt(0).SetUint64(pledgeTotal)

// 	var rate uint64
// 	if bigTotal.Cmp(big.NewInt(0)) != 0 {
// 		rate = bigPledgeTotal.Mul(bigPledgeTotal, big.NewInt(100)).Div(bigPledgeTotal, bigTotal).Uint64()
// 	} else {
// 		rate = 0
// 	}

// 	if rate > 100 {
// 		return 0, fmt.Errorf("rate(%d) is not in the range of 0~100", rate)
// 	}

// 	maxPledge := uint64(1200000000000000)
// 	basePledge := maxPledge * (100 - rate) / 100
// 	return basePledge, nil
// }

// func getBasePledge(sdb *state.StateDB, tx store.Transaction) (uint64, error) {
// 	total, pledgeTotal, err := getWholeNetworkPledgeAndNum(sdb, tx)
// 	if err != nil {
// 		return 0, err
// 	}

// 	bigTotal := big.NewInt(0).SetUint64(total)
// 	bigPledgeTotal := big.NewInt(0).SetUint64(pledgeTotal)

// 	var rate uint64
// 	if bigTotal.Cmp(big.NewInt(0)) != 0 {
// 		rate = bigPledgeTotal.Mul(bigPledgeTotal, big.NewInt(100)).Div(bigPledgeTotal, bigTotal).Uint64()
// 	} else {
// 		rate = 0
// 	}

// 	if rate > 100 {
// 		return 0, fmt.Errorf("rate(%d) is not in the range of 0~100", rate)
// 	}

// 	maxPledge := uint64(1200000000000000)
// 	basePledge := maxPledge * (100 - rate) / 100
// 	return basePledge, nil
// }

// DeleteBlock delete some blocks from the blockchain
// DeleteBlock(10):delete block data larger than 10, including 10
func (bc *Blockchain) DeleteBlock(height uint64) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	dbHeight, err := bc.getMaxBlockHeight()
	if err != nil {
		logger.Error("failed to get height", zap.Error(err))
		return err
	}

	if height > dbHeight {
		logger.SugarLogger.Infof("Wrong height to delete,[%v] should <= current height[%v]", height, dbHeight)
		return nil
	}

	for dH := dbHeight; dH >= height; dH-- {

		DBTransaction := bc.db.NewTransaction()

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

		if err = DBTransaction.Del(append(HeightPrefix, miscellaneous.E64func(block.Height)...)); err != nil {
			logger.Error("Failed to Del height and hash", zap.Error(err))
			return err
		}
		//hash -> block
		hash := block.Hash
		if err = DBTransaction.Del(hash); err != nil {
			logger.Error("Failed to Del block", zap.Error(err))
			return err
		}

		//previous set block into into evm
		previousbBlock, err := bc.getBlockByHeight(dH - 1)
		if err != nil {
			logger.Error("failed to get block", zap.Error(err))
			return err
		}
		previousMiner := *previousbBlock.Miner
		bc.evm.SetBlockInfo(previousbBlock.Height, previousbBlock.Timestamp, previousMiner, previousbBlock.GlobalDifficulty)

		DBTransaction.Set(SnapRootKey, sn)
		DBTransaction.Set(HeightKey, miscellaneous.E64func(dH-1))
		if err := DBTransaction.Commit(); err != nil {
			logger.Error("DBTransaction Commit err", zap.Error(err))
			return err
		}

		DBTransaction.Cancel()
	}

	/* {
		err := bc.rollBackRelease(height - 1)
		if err != nil {
			logger.Error("failed to rollBackRelease", zap.Error(err))
			return err
		}
	}
	*/
	//root, err := DBTransaction.Get(SnapRootKey)
	root, err := getSnapRoot(bc.db)
	if err != nil {
		logger.Error("failed to get SnapRootKey", zap.Error(err))
		return err
	}

	logger.SugarLogger.Info("delete end !!!!! snaproot=", root.String())
	sdb, err := updateNewStateByRoot(bc, root)
	if err != nil {
		return err
	}
	bc.sdb = sdb

	bc.evm = evm.NewEvm(bc.sdb, bc.ChainCfg.ChainId, bc.ChainCfg.GasLimit, new(big.Int).SetUint64(bc.ChainCfg.GasPrice))

	logger.Info("End delete")
	return nil
}

// distr coin out test
func distr(txs []*transaction.SignedTransaction, minaddr *common.Address, height uint64) []*transaction.SignedTransaction {
	total := GetMinerAmount(height)
	//	genesis, _ := address.NewAddrFromString(address.GenesisAddress)
	txm := transaction.NewTransaction()

	txm.Amount = total
	txm.From = &common.Address{}
	txm.To = minaddr
	txm.Type = transaction.CoinBaseTransaction
	txm.Nonce = height
	// txm := transaction.Transaction{
	// 	From: &common.Address{},
	// 	To:   minaddr,
	// 	//Nonce:  0,
	// 	Nonce:  height,
	// 	Amount: total,
	// 	Type:   transaction.CoinBaseTransaction,
	// }
	stxm := transaction.NewSignedTransaction(*txm, []byte{})
	txs = append(txs, stxm)
	return txs
}

// setTxbyaddrKV transaction data is stored by address and corresponding kV
func setTxbyaddrKV(DBTransaction store.Transaction, addr []byte, hash []byte, height, index uint64) error {
	DBTransaction.Mset(addr, hash, []byte(""))
	txindex := &TxIndex{
		Height: height,
		Index:  index,
	}

	tdex, err := json.Marshal(txindex)
	if err != nil {
		logger.Error("Failed Marshal txindex", zap.Error(err))
		return err
	}
	DBTransaction.Set(hash, tdex)

	return err
}

// deleteTxbyaddrKV delete transaction data by address and corresponding kV
func deleteTxbyaddrKV(DBTransaction store.Transaction, addr []byte, tx transaction.FinishedTransaction, index uint64) error {
	txHash := tx.Hash()
	err := DBTransaction.Mdel(addr, txHash)
	if err != nil {
		logger.Error("deleteTxbyaddrKV Mdel err ", zap.Error(err))
		return err
	}

	if err := DBTransaction.Del(txHash); err != nil {
		logger.Error("deleteTxbyaddrKV Del err ", zap.Error(err))
		return err
	}
	return err
}

// func (bc *Blockchain) getWholeNetworkNum() (uint64, error) {
// 	h, err := bc.getMaxBlockHeight()
// 	if err != nil {
// 		return 0, err
// 	}

// 	// //var total uint64 = 170000000000
// 	// x := height / 3153600
// 	// for i := 0; uint64(i) < x; i++ {
// 	// 	total = total * 5 / 10
// 	// }
// 	// return total
// 	var sum uint64
// 	var total uint64 = 100000000000
// 	for i := uint64(0); i < h/3153600; i++ {
// 		sum += total * 3153600
// 		total = total * 5 / 10
// 	}

// 	sum += total*(h%3153600) + 10000000000000666666
// 	return sum, nil
// }

// func getWholeNetworkNum(tx store.Transaction) (uint64, error) {
// 	h, err := getMaxBlockHeight(tx)
// 	if err != nil {
// 		return 0, err
// 	}

// 	var sum uint64
// 	var total uint64 = 100000000000
// 	for i := uint64(0); i < h/3153600; i++ {
// 		sum += total * 3153600
// 		total = total * 5 / 10
// 	}

// 	sum += total*(h%3153600) + 10000000000000666666
// 	return sum, nil
// }

// func (bc *Blockchain) GetWholeNetworkPledgeAndNum() (uint64, uint64, error) {
// 	bc.mu.Lock()
// 	defer bc.mu.Unlock()

// 	return bc.getWholeNetworkPledgeAndNum()
// }

// func (bc *Blockchain) getWholeNetworkPledgeAndNum() (uint64, uint64, error) {
// 	num, err := bc.getWholeNetworkNum()
// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	return num, getWholeNetWorkTotalPledge(bc.sdb), nil
// }

// func getWholeNetworkPledgeAndNum(sdb *state.StateDB, tx store.Transaction) (uint64, uint64, error) {
// 	num, err := getWholeNetworkNum(tx)
// 	if err != nil {
// 		return 0, 0, err
// 	}

// 	return num, getWholeNetWorkTotalPledge(sdb), nil
// }

func (bc *Blockchain) DifficultDetection(b *block.Block) error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	tx := bc.NewTransaction()
	defer tx.Cancel()
	return difficultDetection(b, bc.db, tx)
}

// func (bc *Blockchain) GetBasePledge() (uint64, error) {
// 	bc.mu.RLock()
// 	defer bc.mu.RUnlock()
// 	tx := bc.NewTransaction()
// 	defer tx.Cancel()

// 	return bc.getBasePledge(bc.sdb, tx)
// }

func difficultDetection(b *block.Block, db store.DB, tx store.Transaction) error {
	// 0.更新 difficulty
	gd, err := updateDifficulty(b.Height-1, b.Miner, tx)
	if err != nil {
		return err
	}

	// 1、验证global Difficulty
	if BigToCompact(gd) != BigToCompact(b.GlobalDifficulty) {
		//if gd.Cmp(b.GlobalDifficulty) != 0 {
		logger.Error("compre difficulty", zap.Uint32("global diffculty", BigToCompact(gd)),
			zap.Uint32("b.GlobalDiffculty", BigToCompact(b.GlobalDifficulty)))
		return fmt.Errorf("inconsistent global difficulty")
	}

	// 2、基于global Difficulty 验证miner Difficulty
	// var root common.Hash
	// sr, err := tx.Get(SnapRootKey)
	// if err == store.NotExist {
	// 	root = common.Hash{}
	// } else if err != nil {
	// 	return err
	// } else {
	// 	root = common.BytesToHash(sr)
	// }

	// cdb := bgdb.NewBadgerDatabase(db)
	// sdb := state.NewDatabase(cdb)
	// stdb, err := state.New(root, sdb, nil)
	// if err != nil {
	// 	logger.Error("failed to new state")
	// 	return err
	// }

	// pledge, err := getTotalPledge(stdb, b.Miner)
	// if err != nil {
	// 	return err
	// }

	// basePledge, err := getBasePledge(stdb, tx)
	// if err != nil {
	// 	return err
	// }

	// // minerTarget := calcNextLocalRequiredDifficultyCopy(gd, basePledge, pledge)
	// minerTarget, err := difficulty.NextMinerDifficulty(gd, pledge, basePledge)
	// if err != nil {
	// 	return err
	// }

	// if BigToCompact(minerTarget) != BigToCompact(b.Difficulty) {
	// 	return fmt.Errorf("inconsistent miner difficulty,miner:%s,block:%s", minerTarget.String(), b.Difficulty.String())
	// }

	// 3.计算hash验证难度
	newbig := difficulty.HashToBig(diffhash.Hash(b.MinerHash()))
	if newbig.Cmp(b.GlobalDifficulty) >= 0 {
		fmt.Println("check===================================")
		fmt.Printf("hash:%s,height:%d\n", hex.EncodeToString(b.Hash), b.Height)
		return fmt.Errorf("incorrect difficulty")
	}

	return nil
}

// func calcNextLocalRequiredDifficultyCopy(globalTarget *big.Int, basePledge, minerPledge uint64) *big.Int {
// 	//globalTarget := CompactToBig(GlobalBits)
// 	hashrateTarget := big.NewInt(0).Mul(globalTarget, big.NewInt(4))
// 	pledgeCoinbase := big.NewInt(0).Mul(globalTarget, big.NewInt(6))

// 	if minerPledge > basePledge {
// 		minerPledge = basePledge
// 	}

// 	a := big.NewInt(0).SetUint64((basePledge + minerPledge) * 100)
// 	a = a.Div(a, big.NewInt(0).SetUint64(basePledge))
// 	pleageTarget := pledgeCoinbase.Mul(pledgeCoinbase, a).Div(pledgeCoinbase, big.NewInt(100))

// 	minerTarget := hashrateTarget.Add(hashrateTarget, pleageTarget).Div(hashrateTarget, big.NewInt(10))

// 	return minerTarget
// }

// var powLimit = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))
// var Cycle = 10

// func calcNextGlobalRequiredDifficultyCopy(t0, t1 int64, globalBits uint32) uint32 {

// 	actualTimespan := t1 - t0
// 	adjustedTimespan := actualTimespan

// 	oldTarget := CompactToBig(globalBits)
// 	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
// 	targetTimeSpan := int64(Cycle)
// 	newTarget.Div(newTarget, big.NewInt(targetTimeSpan*10))

// 	if newTarget.Cmp(powLimit) > 0 {
// 		newTarget.Set(powLimit)
// 	}

// 	newTargetBits := BigToCompact(newTarget)
// 	return newTargetBits
// }
