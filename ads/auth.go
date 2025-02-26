package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	_ "github.com/go-sql-driver/mysql"
	"github.com/syndtr/goleveldb/leveldb"
)

const DBINFO = "ubuntu:ubuntu@tcp(localhost:3306)/contractdb"
const AUTH_ROUTINE = 7

// multi-cond, single-dest
type IndexInfo struct {
	Table    string
	Cond     []string
	Dest     string
	DestType string // (sum, max, min), count, select
}

func (ind *IndexInfo) AuthTable() string {
	authtable := ""
	switch ind.DestType {
	case "sum", "max", "min":
		authtable = ind.Table + "_" + strings.Join(ind.Cond, "") + "_" + ind.Dest
	case "count":
		authtable = ind.Table + "_COUNT_" + strings.Join(ind.Cond, "") + "_" + ind.Dest
	case "select":
		authtable = ind.Table + "_SELECT_" + strings.Join(ind.Cond, "") + "_" + ind.Dest
	default:
		return ""
	}

	return authtable
}

type DBData struct {
	Key   []byte
	Value []byte
}

/*
 * support: SELECT sum(dest) FROM table WHERE index="XXX"
 */
func CreateSingleCondIndex(dbinfo string, ind IndexInfo, pk1 []bn.G1Affine, pk2 []bn.G2Affine) error {
	// connect with mysql
	db, err := sql.Open("mysql", dbinfo)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	var currentDB string
	err = db.QueryRow("SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		return fmt.Errorf("failed to get current database name: %v", err)
	}

	flags, err := getCondFlag(ind, currentDB, db)
	if err != nil {
		return err
	}

	for i := range ind.Cond {
		flag := flags[i]
		log.Printf("create index on [%v]->[%v]\n", ind.Cond[i], ind.Dest)

		queryRes, err := queryAll1c1d(db, ind, flags, i)
		if err != nil {
			return err
		}
		dicItems, dicDigest, err := resToDicItem(queryRes, flag, pk1, pk2)
		if err != nil {
			return err
		}
		log.Printf("... writing to authDB")

		//test
		// for i := range 10 {
		// 	fmt.Printf("dic item: (%v,%v) %v\n", dicItems[i].Key.String(), dicItems[i].Nxt.String(), dicItems[i].ValueString())
		// }

		// write dic to level db
		newInd := IndexInfo{Table: ind.Table, Cond: ind.Cond[i : i+1], Dest: ind.Dest, DestType: ind.DestType}
		authfile := filepath.Join(acc.BaseDir, "authdb", newInd.AuthTable())
		// turn dicItems and dicDigest to []DBData
		var datalist []DBData
		for _, it := range dicItems {
			key := it.Key.Bytes()
			// v: key|nxt|value|W
			v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
			datalist = append(datalist, DBData{Key: key[0:32], Value: []byte(v)})
		}
		// record dicDigest with key=INF
		key := acc.INF.Bytes()
		v := acc.G1AffineToString(&dicDigest)
		datalist = append(datalist, DBData{Key: key[:], Value: []byte(v)})
		// write to levelDB
		err = writeLevelDB(authfile, datalist)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeLevelDB(dbfile string, datalist []DBData) error {

	_, err := os.Stat(dbfile)
	// log.Printf("authfile: %v, err: %v", authfile, err)
	if err == nil || os.IsExist(err) {
		log.Printf("AuthDB %v exists, delete it and create a new one.", dbfile)
		os.RemoveAll(dbfile)
	}
	authDB, err := leveldb.OpenFile(dbfile, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return err
	}
	batch := new(leveldb.Batch)

	batchSize := 1 << 15
	batchCount := 0
	for i, it := range datalist {
		batch.Put(it.Key, it.Value)
		batchCount++

		if batchCount >= batchSize || i == len(datalist)-1 {
			err = authDB.Write(batch, nil)
			if err != nil {
				log.Fatalf("Failed to write batch: %v", err)
				return err
			}
			batch.Reset()
			batchCount = 0
		}
	}
	authDB.Close()
	return nil
}

func resToDicItemPlain(
	queryRes map[string]fr.Vector,
	flag int,
	//ks fr.Vector,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (dicItems []acc.DicItem, dicDigest bn.G1Affine, err error) {
	var ks fr.Vector
	count := 0
	for k := range queryRes {
		key, _ := new(fr.Element).SetString(k)
		ks = append(ks, *key)
		count += len(queryRes[k])
	}

	var items []acc.Item
	switch flag {
	case 0:
		var item acc.Item
		for i := range ks {
			item.Key.Set(&ks[i])
			item.Value = acc.ComputeAccG1(queryRes[ks[i].String()], pk1)
			items = append(items, item)
		}
	case 1:
		// sort key
		sort.Slice(ks, func(i, j int) bool {
			return ks[i].Cmp(&ks[j]) == -1
		})

		curPoly := new(acc.Poly).SetOne()
		newPoly := new(acc.Poly)
		var item acc.Item
		for i := range ks {
			newPoly = acc.SetToPoly(queryRes[ks[i].String()])
			curPoly.Mul(curPoly, newPoly)
			item = acc.Item{Key: ks[i], Value: acc.ComputePolyG2(curPoly, pk2)}
			items = append(items, item)
		}

	default:
		return dicItems, dicDigest, fmt.Errorf("invalid flag")
	}
	return acc.CreateDic(items, pk1, pk2)
}

func resToDicItem(
	queryRes map[string]fr.Vector,
	flag int,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (dicItems []acc.DicItem, dicDigest bn.G1Affine, err error) {

	//log.Printf("... generating dic items\n")
	var ks fr.Vector
	count := 0
	for k := range queryRes {
		key, _ := new(fr.Element).SetString(k)
		ks = append(ks, *key)
		count += len(queryRes[k])
	}

	if count <= 5 {
		return resToDicItemPlain(queryRes, flag, pk1, pk2)
	}

	// compute accumulator, record in chResult
	var items []acc.Item
	switch flag { // generate items
	case 0:
		chNum := AUTH_ROUTINE
		chTask := len(ks) / chNum
		if len(ks)%chNum != 0 {
			chTask += 1
		}
		chResult := make(chan []acc.Item, chNum)
		var wgAcc sync.WaitGroup
		for i := range chNum {
			wgAcc.Add(1)
			go func(i int) {
				defer wgAcc.Done()
				chItems := []acc.Item{}
				start := i * chTask
				end := start + chTask
				if end > len(queryRes) {
					end = len(queryRes)
				}
				item := acc.Item{}
				for j := start; j < end; j++ {
					item.Key.Set(&ks[j])
					item.Value = acc.ComputeAccG1(queryRes[ks[j].String()], pk1)
					chItems = append(chItems, item)
				}
				chResult <- chItems
			}(i)
		}
		go func() {
			wgAcc.Wait()
			close(chResult)
		}()

		for res := range chResult {
			items = append(items, res...)
		}
	case 1:
		items = resToDicItemRange(queryRes, ks, pk2)
	default:
		return dicItems, dicDigest, fmt.Errorf("invalid flag for create single-column auth")
	}

	//log.Printf("... computing dic witnesses\n")
	return acc.CreateDic(items, pk1, pk2)
}

func resToDicItemRange(
	queryRes map[string]fr.Vector,
	ks fr.Vector,
	pk2 []bn.G2Affine,
) (items []acc.Item) {
	// sort key
	sort.Slice(ks, func(i, j int) bool {
		return ks[i].Cmp(&ks[j]) == -1
	})

	// compute dic item based on subset poly
	// log.Printf("... ... compute accumulator")
	type Result struct {
		ProcessID int
		Data      []acc.Item
	}

	chNum := AUTH_ROUTINE
	chResult := make(chan Result, chNum)
	var wg sync.WaitGroup

	runtime.GOMAXPROCS(52)
	// logFile, err := os.OpenFile("performance_range_details.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatalf("open log file error: %v", err)
	// }
	// defer logFile.Close()
	// logger := log.New(logFile, "", log.LstdFlags)
	// logger.Printf("len(ks): %v", len(ks))

	chStart := make([]int, chNum+1)
	if len(ks) <= 260 {
		chStart = []int{0, 60, 100, 140, 180, 210, 240, 260}
	} else if len(ks) > 1000 && len(ks) <= 1100 {
		chStart = []int{0, 350, 500, 630, 750, 850, 950, 1100}
	} else if len(ks) > 2000 && len(ks) <= 2600 {
		chStart = []int{0, 800, 1200, 1500, 1800, 2100, 2350, 2600}
	} else if len(ks) > 4000 && len(ks) <= 4100 {
		chStart = []int{0, 1250, 1950, 2400, 2850, 3280, 3650, 4100}
	} else { // no optimization
		avg_task := len(ks) / chNum
		if len(ks)%chNum != 0 {
			avg_task += 1
		}
		for i := range chStart {
			chStart[i] = avg_task * i
		}
	}

	for i := range chNum {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			start := chStart[i]
			end := chStart[i+1]
			if start > len(ks) {
				// log.Printf("rountine %v do not run", i)
				return
			}
			if end > len(ks) {
				end = len(ks)
			}
			//logger.Printf("rountine %v run ...", i)
			//log.Printf("rountine %v run ...", i)
			//stime := time.Now()

			chItem := make([]acc.Item, end-start)

			var curset fr.Vector
			var newsets []fr.Vector
			for j := 0; j < start; j++ {
				curset = append(curset, queryRes[ks[j].String()]...)
			}
			for j := start; j < end; j++ {
				newsets = append(newsets, queryRes[ks[j].String()])
			}
			curPoly := acc.SetToPoly(curset)
			newPoly := new(acc.Poly)
			for j := start; j < end; j++ {
				newPoly = acc.SetToPoly(newsets[j-start])
				curPoly.Mul(curPoly, newPoly)
				chItem[j-start] = acc.Item{Key: ks[j], Value: acc.ComputePolyG2(curPoly, pk2)}
			}

			chResult <- Result{i, chItem}
			//ftime := time.Since(stime)
			//logger.Printf("rountine %v finished, task=[%v,%v), runtime=%v", i, start, end, ftime)
			//log.Printf("rountine %v finished, task=[%v,%v], total_elem=%v, runtime=%v", i, start, end, count, ftime)
		}(i)
	}
	go func() {
		wg.Wait()
		close(chResult)
	}()

	var results []Result
	for res := range chResult {
		results = append(results, res)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].ProcessID < results[j].ProcessID
	})
	for _, res := range results {
		items = append(items, res.Data...)
	}

	//logger.Printf("\n\n")
	return items
}

func queryAll1c1d(db *sql.DB, ind IndexInfo, flags []int, cc int) (res map[string]fr.Vector, err error) {
	table, cond, dest := ind.Table, ind.Cond[cc], ind.Dest
	// check whether table exists in db, col exists in table
	if !CheckColumnExist(db, table, []string{cond, dest}) {
		return nil, fmt.Errorf("cannot create index on non-existing table or column")
	}

	// query all rows from table to get (cond, dest) tuples
	eqQuery := fmt.Sprintf("SELECT %v, %v FROM %v;", cond, dest, table)
	rows, err := db.Query(eqQuery)
	if err != nil {
		return nil, fmt.Errorf("error db.Query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if len(columns) != 2 || err != nil {
		return nil, fmt.Errorf("error rows.Columns(): %v", err)
	}

	switch flags[cc] {
	case 0:
		var itemKey []byte
		var itemVal uint64
		res = make(map[string]fr.Vector)
		for rows.Next() {
			err := rows.Scan(&itemKey, &itemVal)
			if err != nil {
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
			k := new(fr.Element).SetBytes(itemKey).String()
			v := new(fr.Element).SetUint64(itemVal)
			res[k] = append(res[k], *v)
		}
	case 1:
		var itemKey uint64
		var itemVal uint64
		res = make(map[string]fr.Vector)
		for rows.Next() {
			err := rows.Scan(&itemKey, &itemVal)
			if err != nil {
				return nil, fmt.Errorf("error scanning row: %v", err)
			}
			k := new(fr.Element).SetUint64(itemKey).String()
			v := new(fr.Element).SetUint64(itemVal)
			res[k] = append(res[k], *v)
		}
	default:
		return nil, fmt.Errorf("invalid flag type")
	}

	return res, nil
}

func getCondFlag(ind IndexInfo, currentDB string, db *sql.DB) (flags []int, err error) {
	for i := range ind.Cond {
		// find flag
		var flag int
		selectQuery := fmt.Sprintf("SELECT COLUMN_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA=\"%v\" AND TABLE_NAME=\"%v\" AND COLUMN_NAME=\"%v\"", currentDB, ind.Table, ind.Cond[i])
		rows, err := db.Query(selectQuery)
		if err != nil {
			return []int{}, fmt.Errorf("error in db.Query: %v", err)
		}
		defer rows.Close()
		var indexType string
		if rows.Next() {
			err := rows.Scan(&indexType)
			if err != nil {
				return []int{}, fmt.Errorf("error scanning row: %v", err)
			}
		} else {
			return []int{}, fmt.Errorf("no results found")
		}
		switch {
		case strings.HasPrefix(indexType, "bigint") || strings.HasPrefix(indexType, "int"):
			flag = 1
		case strings.HasPrefix(indexType, "varchar") || strings.HasPrefix(indexType, "char") || strings.HasPrefix(indexType, "text"):
			flag = 0
		default:
			return []int{}, fmt.Errorf("unsupported index type: %s", indexType)
		}
		flags = append(flags, flag)

	}
	return flags, nil
}
