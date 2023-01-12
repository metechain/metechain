package main

import (
	"fmt"
	"math/big"

	"metachain/pkg/miner"
)

func main() {

	a, _ := big.NewInt(0).SetString("479411865439304858028307985607831652788339911868488259049321322643236218470", 10)
	tmp := miner.BigToCompact(a)
	b := miner.CompactToBig(tmp)
	fmt.Println(a)
	fmt.Println(b)

	tmpb := miner.BigToCompact(b)
	fmt.Println(tmp)
	fmt.Println(tmpb)

	// data := make([]byte, 1<<10)
	// rand.Read(data)

	// target := miner.CompactToBig(uint32(0x20060f80))
	// sb := miner.CompactToBig(uint32(0x20060f80))
	// sp := sb.Div(sb, big.NewInt(10))
	// a := sp.Mul(sp, big.NewInt(16))

	// var aa, bb = 0, 0
	// for i := 0; i < 1000; i++ {
	// 	r := mr.Intn(1<<10 - 1)
	// 	data[r]++

	// 	bigi := miner.HashToBig(hash.Hash(data))
	// 	if bigi.Cmp(a) < 0 {
	// 		aa++
	// 	}

	// 	if bigi.Cmp(target) < 0 {
	// 		bb++
	// 	}
	// }

	// fmt.Printf("%s=%d \n%s=%d", a.String(), aa, target.String(), bb)
}
