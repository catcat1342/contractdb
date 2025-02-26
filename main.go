package main

import (
	"bufio"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"

	acc "contractdb/accumulator"
	"contractdb/dataset"
	perf "contractdb/performance"

	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	_ "github.com/go-sql-driver/mysql"
)

func InitPubkey() {
	var s fr.Element
	sint, _ := new(big.Int).SetString("43247289572873481232976519249476583910", 10)
	s.SetBigInt(sint)

	n := 10000000
	acc.GenPubKey(&s, n)
}

func GenTestDataAll() {
	for _, ni := range []int{14, 15, 16, 17, 18, 19, 20} {
		//for _, ni := range []int{14} {
		dataset.GenTestData(ni)
		log.Printf("generate workload for N=2^%v ok\n", ni)
	}
}

func main() {
	fmt.Printf("print a number to run a function\n 1. InitPubkey(): generate 10000000 public keys and store in authdb/pubkey\n 2. GenTestDataAll(): generate all dataset used for performance evaluation\n 3. CreateIndexTest(): evaluate index creation performance\n 4. RuntimeQueryWithIntersection(): evaluate query performance\n 5. CreateMultiIndexTest(): evaluate multi-cond index creation performance\n 6. RuntimeQueryWithMultiCond(): evaluate query runtime with multi-cond indexes\n ")
	reader := bufio.NewReader(os.Stdin)
	inputLine, _, err := reader.ReadLine()
	if err != nil {
		fmt.Printf("invalid input: %v\n", err)
		return
	}
	input, err := strconv.Atoi(string(inputLine))
	if err != nil {
		fmt.Printf("invalid input: %v\n", err)
		return
	}

	switch input {
	case 1:
		InitPubkey()
	case 2:
		GenTestDataAll()
	case 3:
		fmt.Printf("this function may require several hours\n")
		perf.CreateIndexTest()
	case 4:
		perf.RuntimeQueryWithIntersection()
	case 5:
		fmt.Printf("this function may require several hours\n")
		perf.CreateMultiIndexTest()
	case 6:
		perf.RuntimeQueryWithMultiCond()
	default:
		fmt.Printf("invalid input")
		return
	}

}
