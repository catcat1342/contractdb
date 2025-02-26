package ads

// import (
// 	acc "contractdb/accumulator"
// 	"fmt"
// 	"log"
// 	"testing"
// )

// func TestCreateIndexSingleEqNonagg(t *testing.T) {
// 	var userInput string
// 	fmt.Printf("Please input table_name, index_column, dest0 dest1 ... to create the index\n")
// 	fmt.Scan("%s", &userInput)

// 	pk1, pk2 := acc.LoadPubkey(100)
// 	dbinfo := "ubuntu:ubuntu@tcp(localhost:3306)/contractdb"
// 	table := "table100"
// 	index := "name"
// 	dest := []string{"value"}
// 	ind := &IndexSingle{
// 		Table: table,
// 		Index: index,
// 		Dest:  dest,
// 	}
// 	// flag := "SingleEq"
// 	err := CreateIndexOnDBNonagg(dbinfo, ind, pk1, pk2)
// 	if err != nil {
// 		log.Printf("error: %v\n", err)
// 	}
// }

// func TestCreateIndexSingleRangeNonagg(t *testing.T) {
// 	var userInput string
// 	fmt.Printf("Please input table_name, index_column, dest0 dest1 ... to create the index\n")
// 	fmt.Scan("%s", &userInput)

// 	pk1, pk2 := acc.LoadPubkey(100)
// 	dbinfo := "ubuntu:ubuntu@tcp(localhost:3306)/contractdb"
// 	table := "table100"
// 	index := "rate"
// 	dest := []string{"value"}
// 	ind := &IndexSingle{
// 		Table: table,
// 		Index: index,
// 		Dest:  dest,
// 	}
// 	// flag := "SingleEq"
// 	err := CreateIndexOnDBNonagg(dbinfo, ind, pk1, pk2)
// 	if err != nil {
// 		log.Printf("error: %v\n", err)
// 	}
// }
