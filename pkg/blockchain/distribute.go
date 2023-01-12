package blockchain

// import (
// 	"math/big"

// 	"metachain/pkg/address"
// )

// var (
// 	MinedKey      = []byte("MinedKey")
// 	TotalMinedKey = []byte("TotalMinedKey")
// )

// //set total mined pledge
// func (bc *Blockchain) setTotalMined(address address.Address, minedVal *big.Int) error {
// 	addr, err := address.NewCommonAddr()
// 	if err != nil {
// 		return err
// 	}
// 	bc.sdb.SetBalance(commonAddrToStoreAddr(addr, MinedKey), minedVal)
// 	return nil
// }

// // Get Total Mined pledge by address
// func (bc *Blockchain) GetTotalMined(address address.Address) (uint64, error) {
// 	bc.mu.Lock()
// 	defer bc.mu.Unlock()

// 	return bc.getTotalMined(address)
// }

// func (bc *Blockchain) getTotalMined(address address.Address) (uint64, error) {
// 	addr, err := address.NewCommonAddr()
// 	if err != nil {
// 		return 0, err
// 	}
// 	saddr := commonAddrToStoreAddr(addr, MinedKey)
// 	ttmBytes := bc.sdb.GetBalance(saddr)
// 	return ttmBytes.Uint64(), nil
// }

// //pledge 70% mined
// func (bc *Blockchain) handleMinedPledge(block *block.Block, miner address.Address, amount uint64) error {
// 	totMinedPledge, err := bc.getTotalMined(miner)
// 	if err != nil {
// 		return err
// 	}

// 	if block.Miner == miner {
// 		minedPledge := amount * MINEDDIVID / DIVIDEND
// 		totMinedPledge += minedPledge
// 		err := bc.setTotalMined(miner, Uint64ToBigInt(totMinedPledge))
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// //release mined pledge
// func (bc *Blockchain) releaseMined(block *block.Block) error {
// 	currHeight := block.Height
// 	var numbers uint64
// 	logger.Info("Start releaseMined", zap.String("currentHeight", fmt.Sprintf("%v", block.Height)))
// 	for control := PledgeCycle; control > 0; control-- {
// 		totalBlocksNum := control * DayMinedBlocks
// 		if currHeight <= uint64(totalBlocksNum) {
// 			continue
// 		}
// 		releaseBlock := currHeight - uint64(totalBlocksNum)

// 		miner, releaseValue, err := bc.getReleaseInfo(releaseBlock)
// 		if err != nil {
// 			logger.Error("getReleaseInfo error", zap.Error(err), zap.Uint64("height", releaseBlock))
// 			if err == store.NotExist {
// 				logger.Error("block not exist", zap.Uint64("height", releaseBlock))
// 				NotExistHeight = releaseBlock
// 				return fmt.Errorf(BlockNotExist)
// 			}
// 			return err
// 		}
// 		totMinedPledge, err := bc.getTotalMined(miner)
// 		if err != nil {
// 			logger.Error("getTotalMined error", zap.Error(err))
// 			return err
// 		}
// 		if totMinedPledge > 0 {

// 			if totMinedPledge >= releaseValue {
// 				totMinedPledge -= releaseValue
// 			}
// 			if totMinedPledge < releaseValue && totMinedPledge > 0 {
// 				totMinedPledge = 0
// 			}

// 			err = bc.setTotalMined(miner, Uint64ToBigInt(totMinedPledge))
// 			if err != nil {
// 				logger.Error("setTotalMined error", zap.Error(err))
// 				return err
// 			}
// 			numbers++
// 		}
// 	}
// 	logger.Info("End releaseMined", zap.String("success", fmt.Sprintf("currentHeight:%v,release times:%v", currHeight, numbers)))
// 	return nil
// }

// func (bc *Blockchain) getReleaseInfo(releaseH uint64) (address.Address, uint64, error) {
// 	var releaseValue uint64
// 	addr := address.Address{}
// 	b, err := bc.getBlockByHeight(releaseH)
// 	if err != nil {
// 		return addr, releaseValue, err
// 	}

// 	for _, tx := range b.Transactions {
// 		if tx.IsCoinBaseTransaction() {
// 			releaseValue = tx.Amount * MINEDDIVID / DIVIDEND / PledgeCycle
// 			addr = b.Miner
// 			break
// 		}
// 	}
// 	if releaseValue == 0 {
// 		return addr, releaseValue, fmt.Errorf("getReleaseInfo value failed!")
// 	}
// 	return addr, releaseValue, nil
// }
