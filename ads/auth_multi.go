package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	_ "github.com/go-sql-driver/mysql"
)

/*
 * support: SELECT sum(dest) FROM table WHERE cond[0]="XXX" and ...
 */
func CreateMultiCondIndex(dbinfo string, ind IndexInfo, pk1 []bn.G1Affine, pk2 []bn.G2Affine) error {
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
	// require that at most one flag is 1
	count0, count1 := 0, 0
	for i := range flags {
		if flags[i] == 0 {
			count0 += 1
		} else if flags[i] == 1 {
			count1 += 1
		} else {
			return fmt.Errorf("unsupported index type")
		}
	}
	if count1 > 1 {
		return fmt.Errorf("unsupported index type: too much range cond")
	}

	type FlagCond struct {
		Cond  string
		Flag  int
		Count int
	}
	var eqconds []FlagCond
	var rangecond FlagCond
	for i := range flags {
		if flags[i] == 0 {
			eqconds = append(eqconds, FlagCond{Cond: ind.Cond[i], Flag: flags[i]})
		} else {
			rangecond = FlagCond{Cond: ind.Cond[i], Flag: flags[i]}
		}
	}

	// query database, acquiring count(cond) for each cond
	for i := range eqconds {
		eqQuery := fmt.Sprintf("SELECT COUNT(DISTINCT %v) FROM %v;", eqconds[i].Cond, ind.Table)
		rows, err := db.Query(eqQuery)
		if err != nil {
			return fmt.Errorf("error db.Query: %v", err)
		}
		rows.Next()
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return fmt.Errorf("error db.Query: %v", err)
		}
		eqconds[i].Count = count

		defer rows.Close()
	}

	sort.Slice(eqconds, func(i, j int) bool {
		return eqconds[i].Count < eqconds[j].Count
	})

	// reorder sorted conditions in ind
	flags = []int{}
	conds := []string{}
	for i := range eqconds {
		flags = append(flags, eqconds[i].Flag)
		conds = append(conds, eqconds[i].Cond)
	}
	if count1 != 0 {
		flags = append(flags, rangecond.Flag)
		conds = append(conds, rangecond.Cond)
	}

	ind.Cond = conds
	switch len(ind.Cond) {
	case 2:
		return createLayerIndex2(db, ind, flags[len(ind.Cond)-1], pk1, pk2)
	case 3:
		return createLayerIndex3(db, ind, flags[len(ind.Cond)-1], pk1, pk2)
	case 4:
		return createLayerIndex4(db, ind, flags[len(ind.Cond)-1], pk1, pk2)
	default:
		return fmt.Errorf("multi-cond do not supported")
	}
}

func createLayerIndex2(db *sql.DB, ind IndexInfo, bottomFlag int, pk1 []bn.G1Affine, pk2 []bn.G2Affine) error {

	logFile, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_multi_details.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	logger.Printf("create layer index on %v\n", ind.Cond)

	start := time.Now()

	// check whether table exists in db, col exists in table
	if !CheckColumnExist(db, ind.Table, []string{ind.Cond[0], ind.Cond[1], ind.Dest}) {
		return fmt.Errorf("cannot create index on non-existing table or column")
	}
	// query all rows from table to get (cond, dest) tuples
	eqQuery := fmt.Sprintf("SELECT %v, %v, %v FROM %v;", ind.Cond[0], ind.Cond[1], ind.Dest, ind.Table)
	rows, err := db.Query(eqQuery)
	if err != nil {
		return fmt.Errorf("error db.Query: %v", err)
	}
	defer rows.Close()

	var queryRes = make(map[string]map[string]fr.Vector)
	if bottomFlag == 0 {
		var k0, k1 []byte
		var result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetBytes(k1).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str] = append(queryRes[k0str][k1str], *val)
		}
	} else {
		var k0 []byte
		var k1, result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetUint64(k1).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str] = append(queryRes[k0str][k1str], *val)
		}
	}

	elapsed := time.Since(start)
	logger.Printf("runtime for generating queryRes: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	var datalist []DBData
	var items0 []acc.Item

	var ks []string
	for k := range queryRes {
		ks = append(ks, k)
	}
	chNum := 10
	chTask := len(ks) / chNum
	if len(ks)%chNum != 0 {
		chTask += 1
	}
	chItem0 := make(chan []acc.Item, chNum)
	chDatalist := make(chan []DBData, chNum)
	var wg sync.WaitGroup
	for i := range chNum {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			start := i * chTask
			if start > len(ks) {
				return
			}
			end := start + chTask
			if end > len(ks) {
				end = len(ks)
			}

			items := []acc.Item{} // item0
			lists := []DBData{}   // datalist
			for j := start; j < end; j++ {
				k0 := ks[j]
				dicItems2, dicDigest2, _ := resToDicItemPlain(queryRes[k0], bottomFlag, pk1, pk2)
				fr0, _ := new(fr.Element).SetString(k0)
				prefix := fr0.Marshal()
				for _, it := range dicItems2 {
					k := it.Key.Marshal()
					v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
					lists = append(lists, DBData{Key: append(prefix, k...), Value: []byte(v)})
				}
				key, _ := new(fr.Element).SetString(k0)
				items = append(items, acc.Item{Key: *key, Value: dicDigest2})
			}
			chItem0 <- items
			chDatalist <- lists
		}(i)
	}
	go func() {
		wg.Wait()
		close(chItem0)
		close(chDatalist)
	}()

	for chItem := range chItem0 {
		items0 = append(items0, chItem...)
	}
	for chList := range chDatalist {
		datalist = append(datalist, chList...)
	}

	// for k0, map1 := range queryRes {
	// 	dicItems2, dicDigest2, err := resToDicItem(map1, bottomFlag, pk1, pk2)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// add items2 to datalist
	// 	fr0, _ := new(fr.Element).SetString(k0)
	// 	prefix := fr0.Marshal()
	// 	for _, it := range dicItems2 {
	// 		k := it.Key.Marshal()
	// 		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 		datalist = append(datalist, DBData{Key: append(prefix, k...), Value: []byte(v)})
	// 	}
	// 	// create a new item for items1 and append
	// 	key, _ := new(fr.Element).SetString(k0)
	// 	items0 = append(items0, acc.Item{Key: *key, Value: dicDigest2})
	// }

	elapsed = time.Since(start)
	logger.Printf("runtime for generating items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	// generate and write items0
	dicItems0, dicDigest0, err := acc.CreateDic(items0, pk1, pk2)
	if err != nil {
		return err
	}
	for _, it := range dicItems0 {
		k := it.Key.Marshal()
		// v: key|nxt|value|W
		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
		datalist = append(datalist, DBData{Key: k, Value: []byte(v)})
	}
	// record dicDigest with key=INF
	key := acc.INF.Marshal()
	v := acc.G1AffineToString(&dicDigest0)
	datalist = append(datalist, DBData{Key: key, Value: []byte(v)})

	elapsed = time.Since(start)
	logger.Printf("runtime for computing items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()
	// write to levelDB
	authfile := filepath.Join(acc.BaseDir, "authdb", ind.AuthTable())
	err = writeLevelDB(authfile, datalist)

	elapsed = time.Since(start)
	logger.Printf("runtime for writing authDB: %.2f s, len(datalist)=%v\n\n", elapsed.Seconds(), len(datalist))
	return err
}

func createLayerIndex3(db *sql.DB, ind IndexInfo, bottomFlag int, pk1 []bn.G1Affine, pk2 []bn.G2Affine) error {

	logFile, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_multi_details.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("create layer index on %v\n", ind.Cond)

	start := time.Now()

	// check whether table exists in db, col exists in table
	if !CheckColumnExist(db, ind.Table, []string{ind.Cond[0], ind.Cond[1], ind.Cond[2], ind.Dest}) {
		return fmt.Errorf("cannot create index on non-existing table or column")
	}
	// query all rows from table to get (cond, dest) tuples
	eqQuery := fmt.Sprintf("SELECT %v, %v, %v, %v FROM %v;", ind.Cond[0], ind.Cond[1], ind.Cond[2], ind.Dest, ind.Table)
	rows, err := db.Query(eqQuery)
	if err != nil {
		return fmt.Errorf("error db.Query: %v", err)
	}
	defer rows.Close()

	var queryRes = make(map[string]map[string]map[string]fr.Vector)
	if bottomFlag == 0 {
		var k0, k1, k2 []byte
		var result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &k2, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetBytes(k1).String()
			k2str := new(fr.Element).SetBytes(k2).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str]; !ok {
				queryRes[k0str][k1str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str][k2str] = append(queryRes[k0str][k1str][k2str], *val)
		}
	} else {
		var k0, k1 []byte
		var k2, result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &k2, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetBytes(k1).String()
			k2str := new(fr.Element).SetUint64(k2).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str]; !ok {
				queryRes[k0str][k1str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str][k2str] = append(queryRes[k0str][k1str][k2str], *val)
		}
	}

	elapsed := time.Since(start)
	logger.Printf("runtime for generating queryRes: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	var datalist []DBData
	var items0 []acc.Item

	var ks []string
	for k := range queryRes {
		ks = append(ks, k)
	}
	chNum := 10
	chTask := len(ks) / chNum
	if len(ks)%chNum != 0 {
		chTask += 1
	}
	chItem0 := make(chan []acc.Item, chNum)
	chDatalist := make(chan []DBData, chNum)
	var wg sync.WaitGroup
	for i := range chNum {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			start := i * chTask
			if start > len(ks) {
				return
			}
			end := start + chTask
			if end > len(ks) {
				end = len(ks)
			}

			items := []acc.Item{} // item0
			lists := []DBData{}   // datalist
			for j := start; j < end; j++ {
				k0 := ks[j]
				map1 := queryRes[k0]
				var items1 []acc.Item
				for k1, map2 := range map1 {
					dicItems2, dicDigest2, _ := resToDicItemPlain(map2, bottomFlag, pk1, pk2)

					// add items2 to datalist
					fr0, _ := new(fr.Element).SetString(k0)
					fr1, _ := new(fr.Element).SetString(k1)
					prefix := append(fr0.Marshal(), fr1.Marshal()...)
					for _, it := range dicItems2 {
						k := it.Key.Marshal()
						v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
						lists = append(lists, DBData{Key: append(prefix, k...), Value: []byte(v)})
					}

					// create a new item for items1 and append
					key, _ := new(fr.Element).SetString(k1)
					items1 = append(items1, acc.Item{Key: *key, Value: dicDigest2})
				}

				// generate and write items1
				dicItems1, dicDigest1, _ := acc.CreateDic(items1, pk1, pk2)

				// add items1 to datalist
				fr0, _ := new(fr.Element).SetString(k0)
				prefix := fr0.Marshal()
				for _, it := range dicItems1 {
					k := it.Key.Marshal()
					v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
					lists = append(lists, DBData{Key: append(prefix, k...), Value: []byte(v)})
				}
				// create a new item for items0 and append
				key, _ := new(fr.Element).SetString(k0)
				items = append(items, acc.Item{Key: *key, Value: dicDigest1})
			}
			chItem0 <- items
			chDatalist <- lists
		}(i)
	}
	go func() {
		wg.Wait()
		close(chItem0)
		close(chDatalist)
	}()

	for chItem := range chItem0 {
		items0 = append(items0, chItem...)
	}
	for chList := range chDatalist {
		datalist = append(datalist, chList...)
	}

	// for k0, map1 := range queryRes {
	// 	var items1 []acc.Item

	// 	for k1, map2 := range map1 {
	// 		dicItems2, dicDigest2, err := resToDicItem(map2, bottomFlag, pk1, pk2)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// add items2 to datalist
	// 		fr0, _ := new(fr.Element).SetString(k0)
	// 		fr1, _ := new(fr.Element).SetString(k1)
	// 		prefix := append(fr0.Marshal(), fr1.Marshal()...)
	// 		for _, it := range dicItems2 {
	// 			k := it.Key.Marshal()
	// 			v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 			datalist = append(datalist, DBData{Key: append(prefix, k...), Value: []byte(v)})
	// 		}

	// 		// create a new item for items1 and append
	// 		key, _ := new(fr.Element).SetString(k1)
	// 		items1 = append(items1, acc.Item{Key: *key, Value: dicDigest2})
	// 	}

	// 	// generate and write items1
	// 	dicItems1, dicDigest1, err := acc.CreateDic(items1, pk1, pk2)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// add items1 to datalist
	// 	fr0, _ := new(fr.Element).SetString(k0)
	// 	prefix := fr0.Marshal()
	// 	for _, it := range dicItems1 {
	// 		k := it.Key.Marshal()
	// 		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 		datalist = append(datalist, DBData{Key: append(prefix, k...), Value: []byte(v)})
	// 	}
	// 	// create a new item for items0 and append
	// 	key, _ := new(fr.Element).SetString(k0)
	// 	items0 = append(items0, acc.Item{Key: *key, Value: dicDigest1})
	// }

	elapsed = time.Since(start)
	logger.Printf("runtime for generating items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	// generate and write items0
	dicItems0, dicDigest0, err := acc.CreateDic(items0, pk1, pk2)
	if err != nil {
		return err
	}
	for _, it := range dicItems0 {
		k := it.Key.Marshal()
		// v: key|nxt|value|W
		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
		datalist = append(datalist, DBData{Key: k, Value: []byte(v)})
	}
	// record dicDigest with key=INF
	key := acc.INF.Marshal()
	v := acc.G1AffineToString(&dicDigest0)
	datalist = append(datalist, DBData{Key: key, Value: []byte(v)})

	elapsed = time.Since(start)
	logger.Printf("runtime for computing items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()
	// write index to levelDB
	authfile := filepath.Join(acc.BaseDir, "authdb", ind.AuthTable())
	err = writeLevelDB(authfile, datalist)

	elapsed = time.Since(start)
	logger.Printf("runtime for writing authDB: %.2f s, len(datalist)=%v\n\n", elapsed.Seconds(), len(datalist))
	return err
}

func createLayerIndex4(db *sql.DB, ind IndexInfo, bottomFlag int, pk1 []bn.G1Affine, pk2 []bn.G2Affine) error {

	logFile, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_multi_details.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("create layer index on %v\n", ind.Cond)

	start := time.Now()

	// check whether table exists in db, col exists in table
	if !CheckColumnExist(db, ind.Table, []string{ind.Cond[0], ind.Cond[1], ind.Cond[2], ind.Cond[3], ind.Dest}) {
		return fmt.Errorf("cannot create index on non-existing table or column")
	}
	// query all rows from table to get (cond, dest) tuples
	eqQuery := fmt.Sprintf("SELECT %v, %v, %v, %v, %v FROM %v;", ind.Cond[0], ind.Cond[1], ind.Cond[2], ind.Cond[3], ind.Dest, ind.Table)
	rows, err := db.Query(eqQuery)
	if err != nil {
		return fmt.Errorf("error db.Query: %v", err)
	}
	defer rows.Close()

	var queryRes = make(map[string]map[string]map[string]map[string]fr.Vector)
	if bottomFlag == 0 {
		var k0, k1, k2, k3 []byte
		var result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &k2, &k3, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetBytes(k1).String()
			k2str := new(fr.Element).SetBytes(k2).String()
			k3str := new(fr.Element).SetBytes(k3).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str]; !ok {
				queryRes[k0str][k1str] = make(map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str][k2str]; !ok {
				queryRes[k0str][k1str][k2str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str][k2str][k3str] = append(queryRes[k0str][k1str][k2str][k3str], *val)
		}
	} else {
		var k0, k1, k2 []byte
		var k3, result uint64
		for rows.Next() {
			err := rows.Scan(&k0, &k1, &k2, &k3, &result)
			if err != nil {
				return fmt.Errorf("error scanning row: %v", err)
			}
			k0str := new(fr.Element).SetBytes(k0).String()
			k1str := new(fr.Element).SetBytes(k1).String()
			k2str := new(fr.Element).SetBytes(k2).String()
			k3str := new(fr.Element).SetUint64(k3).String()
			val := new(fr.Element).SetUint64(result)

			if _, ok := queryRes[k0str]; !ok {
				queryRes[k0str] = make(map[string]map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str]; !ok {
				queryRes[k0str][k1str] = make(map[string]map[string]fr.Vector)
			}
			if _, ok := queryRes[k0str][k1str][k2str]; !ok {
				queryRes[k0str][k1str][k2str] = make(map[string]fr.Vector)
			}
			queryRes[k0str][k1str][k2str][k3str] = append(queryRes[k0str][k1str][k2str][k3str], *val)
		}
	}

	elapsed := time.Since(start)
	logger.Printf("runtime for generating queryRes: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	type DBDataStr struct {
		Key   string
		Value []byte
	}
	var datalist []DBDataStr
	var items0 []acc.Item

	var ks []string
	for k := range queryRes {
		ks = append(ks, k)
	}
	chNum := 10
	chTask := len(ks) / chNum
	if len(ks)%chNum != 0 {
		chTask += 1
	}
	chItem0 := make(chan []acc.Item, chNum)
	chDatalist := make(chan []DBDataStr, chNum)
	var wg sync.WaitGroup
	for i := range chNum {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			start := i * chTask
			if start > len(ks) {
				return
			}
			end := start + chTask
			if end > len(ks) {
				end = len(ks)
			}

			items := []acc.Item{}  // item0
			lists := []DBDataStr{} // datalist
			for j := start; j < end; j++ {
				k0 := ks[j]
				map1 := queryRes[k0]
				var items1 []acc.Item
				for k1, map2 := range map1 {
					var items2 []acc.Item
					for k2, map3 := range map2 {
						dicItems3, dicDigest3, _ := resToDicItem(map3, bottomFlag, pk1, pk2)
						// add dicItems3 to datalist
						fr0, _ := new(fr.Element).SetString(k0)
						fr1, _ := new(fr.Element).SetString(k1)
						fr2, _ := new(fr.Element).SetString(k2)
						prefix := append(fr0.Marshal(), fr1.Marshal()...)
						prefix = append(prefix, fr2.Marshal()...)
						for _, it := range dicItems3 {
							v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
							lists = append(lists, DBDataStr{string(append(prefix, it.Key.Marshal()...)), []byte(v)})

						}
						// create a new item for items2 and append items2
						key, _ := new(fr.Element).SetString(k2)
						items2 = append(items2, acc.Item{Key: *key, Value: dicDigest3})
					}
					// generate items2
					dicItems2, dicDigest2, _ := acc.CreateDic(items2, pk1, pk2)
					// add items2 to datalist
					fr0, _ := new(fr.Element).SetString(k0)
					fr1, _ := new(fr.Element).SetString(k1)
					prefix := append(fr0.Marshal(), fr1.Marshal()...)
					for _, it := range dicItems2 {
						k := it.Key.Marshal()
						v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
						lists = append(lists, DBDataStr{Key: string(append(prefix, k...)), Value: []byte(v)})

					}
					// create a new item for items0 and append
					key, _ := new(fr.Element).SetString(k1)
					items1 = append(items1, acc.Item{Key: *key, Value: dicDigest2})
				}
				// generate and write items1
				dicItems1, dicDigest1, _ := acc.CreateDic(items1, pk1, pk2)
				// add items1 to datalist
				fr0, _ := new(fr.Element).SetString(k0)
				prefix := fr0.Marshal()
				for _, it := range dicItems1 {
					k := it.Key.Marshal()
					v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
					lists = append(lists, DBDataStr{Key: string(append(prefix, k...)), Value: []byte(v)})
				}
				// create a new item for items0 and append
				key, _ := new(fr.Element).SetString(k0)
				items = append(items, acc.Item{Key: *key, Value: dicDigest1})
			}

			chItem0 <- items
			chDatalist <- lists
		}(i)
	}
	go func() {
		wg.Wait()
		close(chItem0)
		close(chDatalist)
	}()

	for chItem := range chItem0 {
		items0 = append(items0, chItem...)
	}
	for chList := range chDatalist {
		datalist = append(datalist, chList...)
	}

	// for k0, map1 := range queryRes {
	// 	var items1 []acc.Item

	// 	for k1, map2 := range map1 {
	// 		var items2 []acc.Item
	// 		for k2, map3 := range map2 {
	// 			dicItems3, dicDigest3, err := resToDicItem(map3, bottomFlag, pk1, pk2)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			// add dicItems3 to datalist
	// 			fr0, _ := new(fr.Element).SetString(k0)
	// 			fr1, _ := new(fr.Element).SetString(k1)
	// 			fr2, _ := new(fr.Element).SetString(k2)
	// 			prefix := append(fr0.Marshal(), fr1.Marshal()...)
	// 			prefix = append(prefix, fr2.Marshal()...)
	// 			for _, it := range dicItems3 {
	// 				v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 				datalist = append(datalist, DBDataStr{string(append(prefix, it.Key.Marshal()...)), []byte(v)})

	// 			}
	// 			// create a new item for items2 and append items2
	// 			key, _ := new(fr.Element).SetString(k2)
	// 			items2 = append(items2, acc.Item{Key: *key, Value: dicDigest3})
	// 		}
	// 		// generate items2
	// 		dicItems2, dicDigest2, err := acc.CreateDic(items2, pk1, pk2)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		// add items2 to datalist
	// 		fr0, _ := new(fr.Element).SetString(k0)
	// 		fr1, _ := new(fr.Element).SetString(k1)
	// 		prefix := append(fr0.Marshal(), fr1.Marshal()...)
	// 		for _, it := range dicItems2 {
	// 			k := it.Key.Marshal()
	// 			v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 			datalist = append(datalist, DBDataStr{Key: string(append(prefix, k...)), Value: []byte(v)})

	// 		}
	// 		// create a new item for items0 and append
	// 		key, _ := new(fr.Element).SetString(k1)
	// 		items1 = append(items1, acc.Item{Key: *key, Value: dicDigest2})
	// 	}
	// 	// generate and write items1
	// 	dicItems1, dicDigest1, err := acc.CreateDic(items1, pk1, pk2)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	// add items1 to datalist
	// 	fr0, _ := new(fr.Element).SetString(k0)
	// 	prefix := fr0.Marshal()
	// 	for _, it := range dicItems1 {
	// 		k := it.Key.Marshal()
	// 		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
	// 		datalist = append(datalist, DBDataStr{Key: string(append(prefix, k...)), Value: []byte(v)})
	// 	}
	// 	// create a new item for items0 and append
	// 	key, _ := new(fr.Element).SetString(k0)
	// 	items0 = append(items0, acc.Item{Key: *key, Value: dicDigest1})

	// }

	elapsed = time.Since(start)
	logger.Printf("runtime for generating items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()

	// generate and write items0
	dicItems0, dicDigest0, err := acc.CreateDic(items0, pk1, pk2)
	if err != nil {
		return err
	}
	for _, it := range dicItems0 {
		k := it.Key.Marshal()
		// v: key|nxt|value|W
		v := fmt.Sprintf("%v|%v|%v|%v", it.Key.String(), it.Nxt.String(), it.ValueString(), acc.G2AffineToString(&it.W))
		datalist = append(datalist, DBDataStr{Key: string(k), Value: []byte(v)})
	}
	// record dicDigest with key=INF
	key := acc.INF.Marshal()
	v := acc.G1AffineToString(&dicDigest0)
	datalist = append(datalist, DBDataStr{Key: string(key), Value: []byte(v)})

	elapsed = time.Since(start)
	logger.Printf("runtime for computing items0: %.2f s\n", elapsed.Seconds())
	start = time.Now()
	// write to levelDB
	var datalist1 = make([]DBData, len(datalist))
	for i := range datalist {
		datalist1[i] = DBData{[]byte(datalist[i].Key), datalist[i].Value}
	}
	authfile := filepath.Join(acc.BaseDir, "authdb", ind.AuthTable())
	err = writeLevelDB(authfile, datalist1)
	elapsed = time.Since(start)
	logger.Printf("runtime for writing authDB: %.2f s, len(datalist)=%v\n\n", elapsed.Seconds(), len(datalist1))

	return err
}
