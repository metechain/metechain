package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"metachain/pkg/block"
	"metachain/pkg/storage/store"

	"golang.org/x/crypto/sha3"
)

func (bc *Blockchain) CheckBlockRegular(b *block.Block) error {
	tx := bc.NewTransaction()
	defer tx.Cancel()
	return bc.checkBlockRegular(b, bc.db, tx)
}

func (bc *Blockchain) checkBlockRegular(b *block.Block, db store.DB, tx store.Transaction) error {
	if len(b.Transactions) > 101 {
		return errors.New("Too many transactions")
	}

	// checkout Difficulty
	if err := difficultDetection(b, db, tx); err != nil {
		return err
	}

	for _, tx := range b.Transactions {
		if tx.IsCoinBaseTransaction() {
			continue
		}

		if err := checkGas(tx.GasLimit, tx.GasPrice); err != nil {
			return err
		}
	}

	//checkNonceMp := make(map[address.Address]uint64)
	return nil
}

func checkBlockHash(b *block.Block) error {
	copyB := *b
	copyB.GasUsed = new(big.Int)
	copyB.Hash = []byte{}
	data, err := copyB.Serialize()
	if err != nil {
		return err
	}

	hash := sha3.Sum256(data)
	if !bytes.Equal(hash[:], b.Hash) {
		return fmt.Errorf("hash not equal")
	}
	return nil
}

func checkGas(gasLimit, gasPrice *big.Int) error {
	// if gasLimit*gasPrice < MINGASLIMIT || gasLimit*gasPrice > MAXGASLIMIT {
	// 	return fmt.Errorf("gas is too small or too big,gas limit:%d gas price:%d", gasLimit, gasPrice)
	// }
	gasUsed := new(big.Int).Mul(gasLimit, gasPrice)
	if gasUsed.Cmp(Limit.MaxGasLimit()) == 1 || gasUsed.Cmp(Limit.MinGasLimit()) == -1 {
		return fmt.Errorf("gas is too small or too big,gas limit:%s gas price:%s", gasLimit.String(), gasPrice.String())
	}
	return nil
}
