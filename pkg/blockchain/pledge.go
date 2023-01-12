package blockchain

/*
import (
	"math/big"

	"metachain/pkg/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"go.uber.org/zap"
)

var (
	PledgeKey               = []byte("PledgeInfo")
	TotalPledgeKey          = []byte("TotalPledgeKey")
	WholeNetWorkTotalPledge = []byte("WholeNetWorkTotalPledge")
	PledgeReleaseInfo       = []byte("PledgeReleaseInfo")
)

// Get address Total transaction Pledge
func (bc *Blockchain) GetTotalPledge(address common.Address) (uint64, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return getTotalPledge(bc.sdb, address)
}

//get whole network total pledge
func (bc *Blockchain) GetWholeNetWorkTotalPledge() (uint64, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return getWholeNetWorkTotalPledge(bc.sdb), nil
}

//add pledge into pldege pool
func setTotalPledge(sdb *state.StateDB, addr common.Address, totalpledge *big.Int) error {
	// addr, err := address.NewCommonAddr()
	// if err != nil {
	// 	return err
	// }
	sdb.SetBalance(commonAddrToStoreAddr(addr, PledgeKey), totalpledge)
	return nil
}

// Get address Total Pledge
func getTotalPledge(sdb *state.StateDB, addr common.Address) (uint64, error) {

	saddr := commonAddrToStoreAddr(addr, PledgeKey)
	ttpBytes := sdb.GetBalance(saddr)
	return ttpBytes.Uint64(), nil
}

// //handle pledge transaction
// func (bc *Blockchain) handlePledgeTransaction(blockHeight uint64, tx *transaction.Transaction) error {
// 	if tx.Type != transaction.PledgeTrasnaction {
// 		return fmt.Errorf("Not a Pledge Trasnaction hash:%v", transaction.HashToString(tx.Hash()))
// 	}
// 	availableBalance, err := bc.getAvailableBalance(tx.From)
// 	if availableBalance < tx.Amount {
// 		return fmt.Errorf("The pledge balance is insufficient available[%v] < amount[%v]", availableBalance, tx.Amount)
// 	}
// 	totPledge, err := getTotalPledge(bc.sdb, tx.From)
// 	if err != nil {
// 		return err
// 	}
// 	totNetPledge := getWholeNetWorkTotalPledge(bc.sdb)

// 	totPledge += tx.Amount
// 	totNetPledge += tx.Amount

// 	setWholeNetWorkTotalPledge(bc.sdb, totNetPledge)

// 	logger.Info("End handlePledgeTransaction", zap.String("info", fmt.Errorf("current height:%v,from:%v,pledge amount:%v,total pledge:%v,wholenet pledge:%v",
// 		blockHeight, tx.From, tx.Amount, totPledge, totNetPledge).Error()))

// 	return setTotalPledge(bc.sdb, tx.From, Uint64ToBigInt(totPledge))
// }

//get whole network total pledge
func getWholeNetWorkTotalPledge(sdb *state.StateDB) uint64 {
	addr := commonAddrToStoreAddr(common.Address{}, WholeNetWorkTotalPledge)
	return sdb.GetBalance(addr).Uint64()
}

//set whole network total pledge
func setWholeNetWorkTotalPledge(sdb *state.StateDB, total uint64) {
	sdb.SetBalance(commonAddrToStoreAddr(common.Address{}, WholeNetWorkTotalPledge), Uint64ToBigInt(total))
	return
}

//handle Pledge Break Transaction
// func (bc *Blockchain) handlePledgeBreakTransaction(block *block.Block, tx *transaction.Transaction) (uint64, error) {
// 	if tx.Type != transaction.PledgeBreakTransaction {
// 		return 0, fmt.Errorf("Not a PledgeBreakTransaction,hash:%v", transaction.HashToString(tx.Hash()))
// 	}
// 	totPledge, err := getTotalPledge(bc.sdb, tx.From)
// 	if err != nil {
// 		return 0, err
// 	}

// 	totNetPledge := getWholeNetWorkTotalPledge(bc.sdb)

// 	totNetPledge -= totPledge
// 	setWholeNetWorkTotalPledge(bc.sdb, totNetPledge)

// 	err = setTotalPledge(bc.sdb, tx.From, Uint64ToBigInt(0))
// 	if err != nil {
// 		return 0, err
// 	}
// 	logger.Info("End handlePledgeBreakTransaction", zap.String("info", fmt.Errorf("current height:%v,from:%v,pledge amount:%v,total pledge:%v,wholenet pledge:%v",
// 		block.Height, tx.From, tx.Amount, totPledge, totNetPledge).Error()))
// 	return totPledge * uint64(DestroyRatio) / uint64(DIVIDEND), nil
// }

// func (bc *Blockchain) releasePledge(blockHeight uint64) error {
// 	var numbers uint64
// 	logger.Info("Start releasePledge", zap.String("current height", fmt.Errorf("%v", blockHeight).Error()))
// 	for control := (2*PledgeCycle - 1); control >= PledgeCycle; control-- {
// 		tot := control * DayMinedBlocks

// 		if blockHeight <= uint64(tot) {
// 			continue
// 		}

// 		releaseH := blockHeight - uint64(tot)

// 		b, err := bc.getBlockByHeight(releaseH)
// 		if err != nil {
// 			return err
// 		}

// 		for _, tx := range b.Transactions {
// 			if tx.IsPledgeTrasnaction() {
// 				err := bc.release(tx.Amount, tx.From)
// 				if err != nil {
// 					return err
// 				}
// 				numbers++
// 			}
// 		}
// 	}
// 	logger.Info("End releasePledge", zap.String("release times:", fmt.Errorf("%v", numbers).Error()))
// 	return nil
// }

func (bc *Blockchain) release(amount uint64, from common.Address) error {
	releaseValue := amount / PledgeCycle

	totPledge, err := getTotalPledge(bc.sdb, from)
	if err != nil {
		logger.Error("getTotalPledge error", zap.Error(err))
		return err
	}

	totNetPledge := getWholeNetWorkTotalPledge(bc.sdb)

	if totPledge > 0 {
		if totPledge >= releaseValue {
			totPledge -= releaseValue
			totNetPledge -= releaseValue
		}
		if totPledge < releaseValue && totPledge > 0 {
			totPledge = 0
			totNetPledge -= totPledge
		}

		err = setTotalPledge(bc.sdb, from, Uint64ToBigInt(totPledge))
		if err != nil {
			logger.Error("setTotalPledge error", zap.Error(err))
			return err
		}

		setWholeNetWorkTotalPledge(bc.sdb, totNetPledge)
	}
	return nil
}

//==============================Test functions==========================================//
//Just use test
func (bc *Blockchain) SetTotalPledge(address common.Address) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	return setTotalPledge(bc.sdb, address, big.NewInt(100000000))
}

//Just use test
func (bc *Blockchain) SetWholeNetWorkTotalPledge() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	setWholeNetWorkTotalPledge(bc.sdb, uint64(100000000000))
	return nil
}

//just use test
func (bc *Blockchain) HandlePledgeBreakTransaction(addr common.Address) (uint64, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	totPledge, err := getTotalPledge(bc.sdb, addr)
	if err != nil {
		return 0, err
	}
	totNetPledge := getWholeNetWorkTotalPledge(bc.sdb)

	totNetPledge -= totPledge
	setWholeNetWorkTotalPledge(bc.sdb, totNetPledge)

	err = setTotalPledge(bc.sdb, addr, Uint64ToBigInt(0))
	if err != nil {
		return 0, err
	}

	return totPledge * uint64(DestroyRatio) / uint64(DIVIDEND), nil
}

//just use test
// func (bc *Blockchain) HandlePledgeTransaction(blockHeight uint64, tx *transaction.Transaction, releaseInfomap map[string]string) error {
// 	if tx.Type != transaction.PledgeTrasnaction {
// 		return fmt.Errorf("Not a Pledge Trasnaction hash:%v", transaction.HashToString(tx.Hash()))
// 	}
// 	availableBalance, err := bc.getAvailableBalance(tx.From)
// 	if availableBalance < tx.Amount {
// 		return fmt.Errorf("The pledge balance is insufficient available[%v] < amount[%v]", availableBalance, tx.Amount)
// 	}
// 	totPledge, err := getTotalPledge(bc.sdb, tx.From)
// 	if err != nil {
// 		return err
// 	}
// 	totNetPledge := getWholeNetWorkTotalPledge(bc.sdb)

// 	totPledge += tx.Amount
// 	totNetPledge += tx.Amount

// 	setWholeNetWorkTotalPledge(bc.sdb, totNetPledge)

// 	return setTotalPledge(bc.sdb, tx.From, Uint64ToBigInt(totPledge))
// }
*/
