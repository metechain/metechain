package blockchain

/*
import (
	"math/big"
	"testing"

	"metachain/pkg/address"
	"metachain/pkg/crypto"
	"metachain/pkg/crypto/sigs"
	_ "metachain/pkg/crypto/sigs/ed25519"
	"metachain/pkg/logger"
	"metachain/pkg/storage/store/pb"
	"metachain/pkg/transaction"

	"github.com/cockroachdb/pebble"
	"github.com/mr-tron/base58"
)

// init Initialize logger,no exception occurs when the logger package is called
func init() {
	logger.InitLogger(logger.DefaultConfig())
}

// address and private key of some tests
var (
	testAddrA = "otKBvWNrk4Dgi2yeHf2ZEdmHxXadBtdEqAVAmwH1qBbePP2"
	testAddrB = "otK31qswtCNPJEFtfYUXq2rFKwGSfLUfwcPBEWz9UAAbone"
	testAddrC = "otK1Luhvd2MJCBTTBzPRQeBLt2KEXaFxiuTZLr1LZ5FhqMz"
	testPrivA = "2DiXYY3moSdNtwCzCfTMqZgpkX2EAwQyKMZe3BFPc9MvKpx5sDcgDkHS6kGzqnJTaUJULg3VNBVxb8MaFBHHLRLQ"
	testPrivB = "34UGBume6QGsngEeyw1t1yqBMW6g7ChkUAhCMRQ56a2DZetL6tvZv5QrNrvHoZ5Xp9tePuL6x2qVDmeCfT4XARTc"
	testPrivC = "34wPHn8jFFeeyBCr5jNx7S5M3eSLjMnqCiQXcqK6s6oNUEqCNdg2vyVxpwWesTyXQFq3iq9Dw1NSrG9b6gjMC7Ut"
)

// TestNewDB tests whether a blockchain object can be created normally
func TestNewDB(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Errorf("opendb  error>>>>>>>>>%v", err)
	}
	t.Log("start test New()")
	cfg := &ChainConfig{}
	// cfg, _ := config.LoadConfig()

	bc, err := New(pb.New(db), cfg)
	if err != nil {
		t.Errorf("opendb  error>>>>>>>>>%v", err)
	}
	t.Logf(" result blockchain %v", bc)
}

// TestGetBalance tests whether the given address can correctly obtain the balance
func TestGetBalance(t *testing.T) {
	bc := createBlockchain()
	testAddrA, _ := address.NewAddrFromString(testAddrA)
	balA, err := bc.GetBalance(testAddrA)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("testAddrA: %v  balA: %v", testAddrA, balA)
}

// TestGetFreezeBalance tests whether the given address can correctly obtain the frozen balance
// func TestGetFreezeBalance(t *testing.T) {
// 	bc := createBlockchain()
// 	addr, _ := address.NewAddrFromString(testAddrA)
// 	balA, err := bc.GetFreezeBalance(addr)
// 	if err != nil {
// 		t.Errorf("error>>>>>>>>>%v", err)
// 	}
// 	t.Logf("testAddrA: %v  FreeBalA: %v", testAddrA, balA)
// }

// TestGetNonce tests whether the nonce of the address can be obtained normally
func TestGetNonce(t *testing.T) {
	bc := createBlockchain()
	addr, _ := address.NewAddrFromString(testAddrA)
	nonA, err := bc.GetNonce(addr)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("testAddrA: %v  NonA: %v", testAddrA, nonA)
}

//TestNewAddBlock tests creates a new block and adds it to the blockchain
func TestNewAddBlock(t *testing.T) {
	bc := createBlockchain()
	//	address.CurrentNetWork = Mainnet

	var txs []*transaction.SignedTransaction
	from, err := address.NewAddrFromString(testAddrA)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	to, _ := address.NewAddrFromString(testAddrB)
	minaddr, _ := address.NewAddrFromString(testAddrC)
	nonce, _ := bc.getNonce(from)

	//	balance := miscellaneous.E64func(100000000)
	//	_ = setBalance(bc.sdb, from.Bytes(), balance)

	tmp := &transaction.Transaction{
		Version:  1,
		Type:     transaction.LockTransaction,
		Nonce:    nonce,
		From:     from,
		To:       to,
		Amount:   big.NewInt(1000000),
		GasLimit: big.NewInt(10),
		GasPrice: big.NewInt(10),
	}
	priv, _ := base58.Decode(testPrivA)
	siganature, err := sigs.Sign(crypto.TypeED25519, priv, tmp.SignHash())
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}

	stx := transaction.SignedTransaction{
		Transaction: *tmp,
		Signature:   *siganature,
	}

	addA, _ := base58.Decode(testAddrA[3:])
	if err := sigs.Verify(siganature, addA, tmp.SignHash()); err != nil {
		t.Log("verify error")
		t.Errorf("error>>>>>>>>>return")
	}

	txs = append(txs, &stx)
	block, err := bc.NewBlock(txs, minaddr)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf(">>>>>>>>>newblock %v", block)

	// data,err := block.Serialize()

	err = bc.AddBlock(block)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf(">>>>>>>>>addBlock\n")
}

// TestDeleteBlock tests whether an exception occurs when deleting the specified block
func TestDeleteBlock(t *testing.T) {
	bc := createBlockchain()
	var delHeight uint64 = 2
	err := bc.DeleteBlock(delHeight)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
}

// TestGetMaxBlockHeight tests whether the maximum block height can be obtained normally
func TestGetMaxBlockHeight(t *testing.T) {
	bc := createBlockchain()
	maxh, err := bc.GetMaxBlockHeight()
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("MaxBlockHeight: %v", maxh)
}

// TestGetHash tests whether the hash corresponding to the block height can be obtained normally
func TestGetHash(t *testing.T) {
	bc := createBlockchain()
	var testHeight uint64 = 1
	bhash, err := bc.GetHash(testHeight)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("blockHash: %v", bhash)

}

// TestGetBlockByHeight tests whether the corresponding block data can be obtained normally through the block height
func TestGetBlockByHeight(t *testing.T) {
	bc := createBlockchain()

	var testHeight uint64 = 33
	blo, err := bc.GetBlockByHeight(testHeight)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("block: %v", blo)
}

// TestGetTransactionByHash tests whether transaction data can be obtained through hash normally
func TestGetTransactionByHash(t *testing.T) {
	bc := createBlockchain()
	testHash := []byte("ef5261c62bf22e0704f2da781da001b6678987099c0c47e89c08d20996822587")
	tx, err := bc.GetTransactionByHash(testHash)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("txa : %v ", tx)
}

// TestGetHeight tests whether the current block height can be obtained normally
// TestGetSnapRoot tests whether the current snapshot can be obtained normally
func TestGetSnapRoot(t *testing.T) {
	bc := createBlockchain()
	root, err := getSnapRoot(bc.db)
	if err != nil {
		t.Errorf("error>>>>>>>>>%v", err)
	}
	t.Logf("root: %v", root)
}

// newBlockchain create blockchain objects for testing
func createBlockchain() *Blockchain {
	db, _ := pebble.Open("pebble.db", &pebble.Options{})
	cfg := &ChainConfig{}
	bc, _ := New(pb.New(db), cfg)
	return bc
	} */
