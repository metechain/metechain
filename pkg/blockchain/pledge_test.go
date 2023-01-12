package blockchain

/*
import (
	"fmt"
	"testing"

	"metachain/pkg/address"
	"metachain/pkg/storage/store/pb"

	"github.com/cockroachdb/pebble"
)

func TestSetTotalPledge(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		panic(err)
	}

	str1 := "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKp"
	addr1, err := address.StringToAddress(str1)
	if err != nil {
		t.Error(err)
	}

	err = b.SetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTotalPledge(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		panic(err)
	}

	str1 := "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKp"
	addr1, err := address.StringToAddress(str1)
	if err != nil {
		t.Error(err)
	}

	err = b.SetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}

	ttp, err := b.GetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("address:%v total pledge:%v\n", addr1, ttp)

}

func TestSetWholeNetWorkTotalPledge(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		panic(err)
	}

	err = b.SetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}
}

func TestGetWholeNetWorkTotalPledge(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		panic(err)
	}

	err = b.SetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}

	ttw, err := b.GetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf(" WholeNetWorkTotalPledge:%v\n", ttw)
}

func TestHandlePledgeBreakTransaction(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		panic(err)
	}

	str1 := "otK5XLQHTym83ygtAk6XanyYSioatrnGTm1jYtddAEVNNKp"
	addr1, err := address.StringToAddress(str1)
	if err != nil {
		t.Error(err)
	}

	err = b.SetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}

	err = b.SetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}

	ttp, err := b.GetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("address:%v total pledge:%v\n", addr1, ttp)

	wnt, err := b.GetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("GetWholeNetWorkTotalPledge:%v \n", wnt)

	breakNum, err := b.HandlePledgeBreakTransaction(addr1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("break amount:%v \n", breakNum)

	ttp2, err := b.GetTotalPledge(addr1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("address:%v total pledge left:%v\n", addr1, ttp2)

	wnt2, err := b.GetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("GetWholeNetWorkTotalPledge left:%v \n", wnt2)
}

func TestRelease(t *testing.T) {
	db, err := pebble.Open("pebble.db", &pebble.Options{})
	if err != nil {
		t.Error("open db:", err)
	}
	b, err := New(pb.New(db), &ChainConfig{})
	if err != nil {
		t.Errorf("New >>>>>>>>>%v", err)
	}

	from := "otK9sFhbjDdjEHvcdH6n9dtQws1m4ptsAWAy7DhqGdrUFai"
	addr, _ := address.StringToAddress(from)

	err = b.SetTotalPledge(addr)
	if err != nil {
		t.Error(err)
	}

	err = b.SetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}

	err = b.release(uint64(100000000), addr)
	if err != nil {
		t.Error(err)
	}

	totPledge, err := b.GetTotalPledge(addr)
	if err != nil {
		t.Error(err)
	}

	totNetPledge, err := b.GetWholeNetWorkTotalPledge()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("totPledge:%v,totNetPledge:%v\n", totPledge, totNetPledge)
}
*/
