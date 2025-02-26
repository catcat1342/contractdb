package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"
)

// IndexSingle is the index of single column query
// index is the condition column, dest is the required column
type IndexSingle struct {
	Table string
	Index string
	Dest  []string
}

type Row [][]byte // single row in the result

func (r Row) Bytes() []byte { // 8-byte concact, used for compute hash
	var res []byte
	for _, ri := range r {
		diff := len(ri) - acc.SHORTBYTES
		if diff < 0 {
			pad := make([]byte, -diff)
			ri = append(pad, ri...)
		} else if diff > 0 {
			ri = ri[diff:]
		}
		res = append(res, ri...)
	}
	return res
}

func CheckTableExist(db *sql.DB, tables []string) bool {
	// check whether table exists in db, col exists in table
	var currentDB string
	err := db.QueryRow("SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		log.Panicf("failed to get current database name: %v", err)
		return false
	}

	for _, table := range tables {
		var tableCount int

		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '%v' AND table_name = '%s'", currentDB, table)).Scan(&tableCount)
		if err != nil {
			return false
		}
		if tableCount == 0 {
			return false
		}
	}
	return true
}

func CheckColumnExist(db *sql.DB, table string, cols []string) bool {
	// check whether table exists in db, col exists in table
	var currentDB string
	err := db.QueryRow("SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		log.Panicf("failed to get current database name: %v", err)
		return false
	}

	var tableCount int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '%v' AND table_name = '%s'", currentDB, table)).Scan(&tableCount)
	if err != nil {
		return false
	}
	if tableCount == 0 {
		return false
	}

	for _, col := range cols {
		var columnCount int
		err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = '%v' AND table_name = '%s' AND column_name = '%s'", currentDB, table, col)).Scan(&columnCount)
		if err != nil {
			return false
		}
		if columnCount == 0 {
			return false
		}
	}

	return true
}
