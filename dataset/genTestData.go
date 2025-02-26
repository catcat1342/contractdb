package dataset

import (
	"contractdb/ads"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/exp/rand"
)

func GenTestData(ni int) { // n=2^ni
	n := 1 << ni
	baseDir := os.Getenv("PROJECT_ROOT")

	filename := "test_" + strconv.Itoa(ni) + ".csv"
	workloadfile := filepath.Join(baseDir, "dataset", "testDataset", filename)

	file, err := os.Create(workloadfile)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}

	writer := csv.NewWriter(file)

	// the title line
	writer.Write([]string{"keyid_BIGINT UNSIGNED", "name_VARCHAR(255)", "bank_VARCHAR(255)", "addr_VARCHAR(255)", "value_BIGINT UNSIGNED NOT NULL", "rate_BIGINT UNSIGNED NOT NULL", "grade_BIGINT UNSIGNED NOT NULL", "result_BIGINT UNSIGNED NOT NULL"})

	// name: 1<<12, value: 1<<12
	// bank: 1<<10, rate: 1<<10
	// addr: 1<<8, grade: 1<<8
	// result: no repeat
	k1, k2, k3 := 1<<12, 1<<10, 1<<8
	resultList := genRandUintStr(n, 1, n+10000)

	// randomly generate n lines data and write to workload.csv
	var keyid, name, bank, addr, value, rate, grade, result string
	for i := 0; i < n; i++ {
		keyid = strconv.Itoa(i + 1)
		name = "Name" + strconv.Itoa(rand.Intn(k1)+1)
		bank = "Bank" + strconv.Itoa(rand.Intn(k2)+1)
		addr = "Addr" + strconv.Itoa(rand.Intn(k3)+1)

		value = strconv.Itoa(rand.Intn(k1) + 1000)
		rate = strconv.Itoa(rand.Intn(k2) + 2000)
		grade = strconv.Itoa(rand.Intn(k3) + 3000)
		// ensure no duplicated on result
		result = resultList[i]

		writer.Write([]string{keyid, name, bank, addr, value, rate, grade, result})
	}
	writer.Flush()
	file.Close()

	log.Println("CSV file generated successfully.")

	// load to contractdb
	err = loadTestDataToSQL(workloadfile, ads.DBINFO, "test"+strconv.Itoa(ni))
	if err != nil {
		fmt.Printf("%v", err)
	}
}

func genRandUintStr(n int, v0, v1 int) []string {
	resultListMap := make(map[string]int)
	count := 0

	if v0 == 0 && v1 == 0 {
		var rval uint64
		var rstr string
		for count <= n {
			rval = rand.Uint64()
			if rval == 0 {
				continue
			}
			rstr = strconv.FormatUint(rval, 10)
			if _, exists := resultListMap[rstr]; !exists {
				resultListMap[rstr] = 1
				count++
			}
		}
	} else {
		var rval int
		var rstr string
		for count <= n {
			rval = rand.Intn(v1)
			if rval == 0 {
				continue
			}
			if rval < v0 {
				rval += v0
			}
			rstr = strconv.Itoa(rval)
			if _, exists := resultListMap[rstr]; !exists {
				resultListMap[rstr] = 1
				count++
			}
		}
	}

	resultList := []string{}
	for k := range resultListMap {
		resultList = append(resultList, k)
	}
	return resultList
}

func loadTestDataToSQL(workloadfile string, dbinfo string, tablename string) error {
	// open workload.csv
	file, err := os.Open(workloadfile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	log.Printf("open csv file: %v\n", workloadfile)
	defer file.Close()

	// read workload.csv
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %v", err)
	}
	// log.Printf("recordes len: %v\n", len(records))

	// connect with mysql
	db, err := sql.Open("mysql", dbinfo)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	} else {
		log.Println("connect to contractdb successfully")
	}
	defer db.Close()

	// create table with headers
	headers := records[0]

	log.Printf("tablename: %v\n", tablename)

	// if exist, drop it and create a new one
	if ads.CheckTableExist(db, []string{tablename}) {
		_, err = db.Exec("DROP TABLE " + tablename)
		if err != nil {
			log.Panicf("err during DROP TABLE: %v\n", err)
		}
		log.Printf("drop table %v\n", tablename)
	}
	createTableQuery := "CREATE TABLE IF NOT EXISTS " + tablename + " ("
	for _, header := range headers {
		parts := strings.Split(header, "_")
		createTableQuery += parts[0] + " " + parts[1] + ", "
	}
	createTableQuery = createTableQuery[:len(createTableQuery)-2] + ")"
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	log.Printf("create table %v successfully\n", tablename)

	// load data line by line
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO " + tablename + " VALUES (?,?,?,?,?,?,?,?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, record := range records[1:] {
		r0, _ := strconv.ParseUint(record[0], 10, 64)
		r4, _ := strconv.ParseUint(record[4], 10, 64)
		r5, _ := strconv.ParseUint(record[5], 10, 64)
		r6, _ := strconv.ParseUint(record[6], 10, 64)
		r7, _ := strconv.ParseUint(record[7], 10, 64)
		_, err := stmt.Exec(r0, record[1], record[2], record[3], r4, r5, r6, r7)
		if err != nil {
			return fmt.Errorf("failed to insert data: %v", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	log.Printf("create and clean %v, write %v to contractdb successfully.", tablename, workloadfile)

	return nil
}
