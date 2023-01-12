package txpool

/*
import (
	"math/big"
	"testing"

	"metachain/pkg/address"
	_ "metachain/pkg/logger"
	"metachain/pkg/transaction"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestStHeappush(t *testing.T) {
	assert := assert.New(t)
	q := newQueue()
	p1, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(p1.PublicKey)
	p2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(p2.PublicKey)
	p3, _ := crypto.GenerateKey()
	addr3 := crypto.PubkeyToAddress(p3.PublicKey)
	srcList := []transaction.SignedTransaction{
		0:  {Transaction: transaction.Transaction{From: addr1, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(3)}},
		1:  {Transaction: transaction.Transaction{From: addr2, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(8)}},
		2:  {Transaction: transaction.Transaction{From: addr3, Nonce: 2, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(3)}},
		3:  {Transaction: transaction.Transaction{From: addr1, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(5)}},
		4:  {Transaction: transaction.Transaction{From: addr2, Nonce: 2, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(6)}},
		5:  {Transaction: transaction.Transaction{From: addr3, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(1)}},
		6:  {Transaction: transaction.Transaction{From: addr1, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(7)}},
		7:  {Transaction: transaction.Transaction{From: addr2, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(4)}},
		8:  {Transaction: transaction.Transaction{From: addr3, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(9)}},
		9:  {Transaction: transaction.Transaction{From: addr3, Nonce: 4, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(1)}},
		10: {Transaction: transaction.Transaction{From: addr3, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(10)}},
	}

	//sort
	{

		inputIdx := []int{0, 1, 2, 3, 4, 5}
		for _, idx := range inputIdx {
			q.push(srcList[idx])
		}

		outputIdx := []int{1, 4, 3, 0, 5, 2}
		for _, idx := range outputIdx {
			maybe := q.pop()
			assert.Equal(srcList[idx].GasCap(), maybe.GasCap())
		}
	}

	// update
	{
		inputIdx := []int{0, 1, 2, 3, 4, 5, 6}
		for _, idx := range inputIdx {
			q.push(srcList[idx])
		}

		outputIdx := []int{1, 6, 4, 0, 5, 2}
		for _, idx := range outputIdx {
			maybe := q.pop()
			assert.Equal(srcList[idx].GasCap(), maybe.GasCap())
		}
	}

	// remove
	{
		inputIdx := []int{0, 1, 2, 3, 4, 5, 6}
		for _, idx := range inputIdx {
			q.push(srcList[idx])
		}

		q.remove(2)
		outputIdx := []int{1, 6, 4, 5, 2}
		for _, idx := range outputIdx {
			maybe := q.pop()
			assert.Equal(srcList[idx].GasCap(), maybe.GasCap())
		}
	}

	// update and remove
	{
		inputIdx := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		for _, idx := range inputIdx {
			q.push(srcList[idx])
		}

		q.remove(2)

		outputIdx := []int{10, 8, 1, 6, 4, 7, 0, 9}
		for _, idx := range outputIdx {
			maybe := q.pop()
			assert.Equal(srcList[idx].GasCap(), maybe.GasCap())
		}
	}
}

func TestFindTxByHash(t *testing.T) {
	assert := assert.New(t)
	q := newQueue()

	addr1, _ := address.NewAddrFromString("otKG7B6mFNNHHLNqLCbFMhz2bbwPJuRv4fMbthXJtcaCAJq")
	addr2, _ := address.NewAddrFromString("otKFVuKvsDLUb5zWMutcroqs8WiocjgmWuF55WE4GYvfhvA")
	addr3, _ := address.NewAddrFromString("otK4uXfcTtYYRfFprzxuxzAqqgjx2nTdKUw1WdzybQ2ukn6")
	srcList := []transaction.SignedTransaction{
		0:  {Transaction: transaction.Transaction{From: addr1, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(3)}},
		1:  {Transaction: transaction.Transaction{From: addr2, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(8)}},
		2:  {Transaction: transaction.Transaction{From: addr3, Nonce: 2, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(3)}},
		3:  {Transaction: transaction.Transaction{From: addr1, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(5)}},
		4:  {Transaction: transaction.Transaction{From: addr2, Nonce: 2, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(6)}},
		5:  {Transaction: transaction.Transaction{From: addr3, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(1)}},
		6:  {Transaction: transaction.Transaction{From: addr1, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(7)}},
		7:  {Transaction: transaction.Transaction{From: addr2, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(4)}},
		8:  {Transaction: transaction.Transaction{From: addr3, Nonce: 3, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(9)}},
		9:  {Transaction: transaction.Transaction{From: addr3, Nonce: 4, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(1)}},
		10: {Transaction: transaction.Transaction{From: addr3, Nonce: 1, Amount: big.NewInt(1), GasLimit: big.NewInt(1), GasFeeCap: big.NewInt(1), GasPrice: big.NewInt(10)}},
	}

	for _, st := range srcList {
		q.push(st)
	}

	hash := srcList[1].HashToString()
	t.Log("1 hash:", srcList[1].HashToString())
	addr, nonce, err := q.getAddrAndNonceByHash(hash)
	assert.NoError(err)

	rst, err := q.getIndex(addr, nonce)
	assert.NoError(err)

	q.remove(rst.idx)
	_, _, err = q.getAddrAndNonceByHash(hash)
	assert.NoError(err)
}
*/
