package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	mr "math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"metachain/pkg/miner/hash"
)

var (
	num    int           = 1
	cpuNum int           = 1
	mode   string        = "number"
	hnum   time.Duration = time.Duration(10) * time.Minute
)

// ./1kh -m=time -cpu=100 -t=1_h

func init() {
	m := flag.String("m", "number", "calculation mode")
	n := flag.Int("n", 1, "hash numbers")
	cpu := flag.Int("cpu", 1, "cpu numbers")
	t := flag.String("t", "1_m", "几小时打印一次")
	flag.Parse()

	maxCpuNum := runtime.NumCPU()
	if *cpu > 0 && *cpu <= maxCpuNum {
		cpuNum = *cpu
	}

	mode = *m

	if mode == "number" {
		num = *n
		fmt.Printf("cpu num:%d \nhash num:%d \nmax cpu num:%d\n", cpuNum, num, maxCpuNum)
	} else {

		s := *t
		i := strings.IndexRune(s, '_')

		if len(s[:i]) == 0 || len(s[i+1:]) == 0 {
			panic("t invalid")
		}

		c, err := strconv.ParseInt(s[:i], 10, 32)
		if err != nil {
			panic(err)
		}

		switch s[i+1] {
		case 's':
			hnum = time.Duration(c) * time.Second
		case 'm':
			hnum = time.Duration(c) * time.Minute
		case 'h':
			hnum = time.Duration(c) * time.Hour
		default:
			panic("t invalid")
		}

		fmt.Printf("cpu num:%d \ntime num:%v \nmax cpu num:%d\n", cpuNum, hnum, maxCpuNum)
	}
}

func main21() {
	if mode == "number" {
		HashOfNumber()
	} else {
		HashOfTime(hnum)
	}
}

func HashOfNumber() {
	ch := make(chan []byte, 100)
	data := make([]byte, 1<<10)
	rand.Read(data)
	go dataToChan(num, data, ch)
	t := time.Now()
	hashCh := make(chan struct{}, 1)
	mutilHash(cpuNum, ch, hashCh)
	subt := time.Since(t)
	fmt.Printf("total time:%v\n", subt)
	fmt.Printf("1kh:%v\n", subt/time.Duration(num))
}

func HashOfTime(d time.Duration) {
	ch := make(chan []byte, 100)
	hashCh := make(chan struct{}, 1)
	data := make([]byte, 1<<10)
	rand.Read(data)
	go timetoChan(d, data, ch, hashCh)
	t := time.Now()

	mutilHash(cpuNum, ch, hashCh)
	subt := time.Since(t)
	fmt.Printf("total time:%v\n", subt)
	fmt.Printf("1kh:%v\n", subt/time.Duration(num))
}

func dataToChan(n int, data []byte, ch chan []byte) {

	for i := 0; i < n*1000; i++ {
		r := mr.Intn(1<<10 - 1)
		data[r]++

		ch <- data
	}

	close(ch)
}

func timetoChan(d time.Duration, data []byte, ch chan []byte, hashCh chan struct{}) {
	go func(d time.Duration) {
		ticker := time.NewTicker(d)
		n := uint64(0)

		for {
			select {
			case <-ticker.C:
				fmt.Printf("time:%v  number:%d\n", d, n)
				n = 0
			case <-hashCh:
				n++
			}
		}
	}(d)

	for {
		r := mr.Intn(1<<10 - 1)
		data[r]++

		ch <- data
	}
}

func mutilHash(cpuNum int, ch chan []byte, hashCh chan struct{}) {
	var wg sync.WaitGroup
	for i := 0; i < cpuNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				data, ok := <-ch
				if !ok {
					break
				}

				hash.Hash(data)
				hashCh <- struct{}{}
			}
		}()
	}

	wg.Wait()
}
