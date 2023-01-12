package blockchain

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core/state"
)

func readStatedbDate(sdb *state.StateDB, fileName string) {
	//fileh, _ := os.OpenFile("testHash.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
	name := fmt.Sprintf("./sdbData/%v", fileName)

	os.Mkdir("sdbData", os.ModePerm)

	//	_ = IfNoFileToCreate(name)

	fileh, _ := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0664)
	defer fileh.Close()
	writer := bufio.NewWriter(fileh)
	_, err := writer.WriteString("statedb data:\n")
	if err != nil {
		fmt.Println("writer error")
	}
	a := sdb.RawDump(nil)
	for addr, c := range a.Accounts {
		data := fmt.Sprintf("[%v] n:%v b:%v\n", addr, c.Nonce, c.Balance)
		_, err := writer.WriteString(data)
		if err != nil {
			fmt.Println("writer error")
		}
	}
	writer.Flush()
}
