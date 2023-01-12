package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"metachain/pkg/miner/hash"
)

const (
	Cycle = 10
)

var powLimit = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(1))

func calcNextRequiredDifficulty(t0, t1 int64, bits uint32) uint32 {

	actualTimespan := t1 - t0
	adjustedTimespan := actualTimespan

	oldTarget := CompactToBig(bits)
	newTarget := new(big.Int).Mul(oldTarget, big.NewInt(adjustedTimespan))
	targetTimeSpan := int64(Cycle)
	newTarget.Div(newTarget, big.NewInt(targetTimeSpan))

	if newTarget.Cmp(powLimit) > 0 {
		newTarget.Set(powLimit)
	}

	newTargetBits := BigToCompact(newTarget)
	return newTargetBits
}

func HashToBig(buf []byte) *big.Int {
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}
	return new(big.Int).SetBytes(buf[:])
}

func CompactToBig(compact uint32) *big.Int {
	mantissa := compact & 0x007fffff
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

func main() {
	i := 100

	ch := make(chan int, 1)
	start := make(chan int, 1)
	start <- 0
	go func() {
		/* 	num := <-start
		fmt.Println("num", num) */
		for ; i > 0; i-- {

			ch <- i
			fmt.Println("send ", i)
		}
		panic("net finish")
	}()

	for {

		recv := <-ch
		fmt.Println("recv", recv)
		time.Sleep(time.Second)
		/* 	start <- 0 */
		fmt.Println("recv finish", recv)
	}

}

func main1() {
	t := time.Now()
	data := make([]byte, 1<<10)
	bits := uint32(0x20000000)
	//bits := uint32(0x207fffff)
	target := CompactToBig(bits)
	cnt := 0
	for {

		_, err := rand.Read(data)
		if err != nil {
			log.Fatal(err)
		}
		if HashToBig(hash.Hash(data)).Cmp(target) <= 0 {
			fmt.Printf("sucess: %v\n", time.Now().Sub(t))
			cnt++
			if cnt == Cycle {
				bits = calcNextRequiredDifficulty(t.Unix(), time.Now().Unix(), bits)
				{
					fmt.Printf("begin reset %v: %v\n", time.Now().Sub(t), bits)
				}
				cnt = 0
				t = time.Now()
			}
		}
	}
	// ci()
}

// // 207fffff
// // 0010 0000 0111 1111 1111 1111 1111 1111

// func ci() {
// 	{
// 		data := make([]byte, 1<<10)
// 		_, err := rand.Read(data)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		t := time.Now()
// 		for i := 0; i < 100; i++ {
// 			sha256.Sum256(data)
// 			// hash.Hash(data)
// 		}
// 		fmt.Printf("process: %v\n", time.Now().Sub(t)/100)
// 	}
// }
