package blockchain

/*
import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/state"
	"metachain/pkg/contract/exec"
	"metachain/pkg/contract/parser"
	"metachain/pkg/storage/store"
)

func newToken(CDBTransaction store.Transaction, sdb *state.StateDB, addr string) error {
	sc := parser.Parser([]byte("new \"cm\" 1000000000000 8"))
	e, err := exec.New(CDBTransaction, sdb, sc, addr)
	if err != nil {
		return err
	}
	e.Flush(CDBTransaction, sdb)
	return nil
}

func mintToken(CDBTransaction store.Transaction, sdb *state.StateDB, addr string) error {
	sc := parser.Parser([]byte("mint \"cm\" 1000000000000"))
	e, err := exec.New(CDBTransaction, sdb, sc, addr)
	if err != nil {
		return err
	}
	fmt.Printf("root : %x\n", e.Root())
	e.Flush(CDBTransaction, sdb)
	return nil
}

func getBalanceToken(sdb *state.StateDB, name string) error {
	b, err := exec.TokenBalance(sdb, "cm", name)
	if err != nil {
		return err
	}
	fmt.Printf("address A :%v\n", b)
	return nil
}

func initToken(bc *Blockchain) {
	DBTransaction := bc.db.NewTransaction()
	defer DBTransaction.Cancel()
	fmt.Println(">>>>>>>>>>>>>>>>>>>new")
	newToken(DBTransaction, bc.sdb, "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKn")
	fmt.Println(">>>>>>>>>>>>>>>>>>>mint")
	mintToken(DBTransaction, bc.sdb, "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKn")
	fmt.Println(">>>>>>>>>>>>>>>>>>>get")
	getBalanceToken(bc.sdb, "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKn")
	//priv: 2QKggfbFTc86k6s1oFoUvVBfNTx4qrUXv8epszGLPKSmbLByzLF1z9L47NCDRBt2kRDkH4Uw2Za4P7Gsd3nTmfrC
	if err := DBTransaction.Commit(); err != nil {
		return
	}

}
*/
