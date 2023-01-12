//Sync  Blockchain
package consensus

import (
	"encoding/hex"
	"math/big"
	"time"

	"metachain/pkg/block"
	"metachain/pkg/logger"

	"go.uber.org/zap"
)

//hash存在则丢弃块，
//prevhash 不存在，加入孤块列表
//prevhash 存在，则将块加入chain，判断加入当前块是否时最长链
//块加入后，hash在孤块列表,则延长当前链，筛选孤块列表中过期的块

//QUESTION
// 1.如何判断最长链的
// 2.孤块中超过了当前的最大块高，如何保证跟其他节点保持一致，
// 3.使用orphans中的块延长链时，如何确认最长链

const (
	MaxEqealBlockWeight = 10
	MaxExpiration       = time.Hour
	MaxOrphanBlocks     = 200
)

//ProcessBlock is management block function
func (b *BlockChain) ProcessBlock(newblock *block.Block, globalDifficulty *big.Int) bool {

	//newblcok hash is exist
	defer logger.Info(" ProcessBlock  end ", zap.Uint64("height", newblock.Height), zap.String("hash", hex.EncodeToString(newblock.Hash)))

	if b.BlockExists(newblock.Hash) {
		logger.SugarLogger.Info("Block is exist", hex.EncodeToString(newblock.Hash))
		return false
	}

	hash := BytesToHash(newblock.Hash)
	if _, exist := b.Oranphs[hash]; exist {
		logger.Info("orphan is exist", zap.String("hash", hex.EncodeToString(newblock.Hash)))
		return false
	}

	/* 	maxHeight, err := b.Bc.GetMaxBlockHeight()
	   	if err != nil {
	   		logger.SugarLogger.Error("GetBlockByHeight err", err)
	   		return false
	   	}

	   	if maxHeight == blockchain.InitHeight {
	   		logger.SugarLogger.Error("==========first blcok============", hex.EncodeToString(newblock.Hash))
	   		err = b.Bc.AddBlock(newblock)
	   		if err != nil {
	   			logger.SugarLogger.Error("first blcok addblock err", err)
	   			return false
	   		}
	   		return true
	   	} */

	//判断prevhash是否存在，
	if !b.BlockExists(newblock.PrevHash) {
		logger.Info("prevhash not exist")
		b.AddOrphanBlock(newblock)
		return false
	}

	//maybeAcceptBlock return longest chain flag
	succ, mainChain := b.maybeAcceptBlock(newblock)
	if !succ {
		return false
	}
	ok := b.ProcessOrphan(newblock)
	if ok {
		mainChain = ok
	}

	return mainChain
}

func bigToCompact(n *big.Int) uint32 {
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
