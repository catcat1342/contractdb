package performance

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	acc "contractdb/accumulator"
	"contractdb/ads"

	_ "github.com/go-sql-driver/mysql"
)

func CreateIndexTPCH6() {
	runtime.GOMAXPROCS(64)
	logFile, err := os.OpenFile("performance_tpch6.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	log.Printf("loading pub key")
	MAXN := 1<<20 + 100
	pk1, pk2 := acc.LoadPubkey(MAXN)
	for _, ni := range []int{14, 15, 16, 17, 18, 19, 20} {
		//for _, ni := range []int{14} {
		logger.Printf("ni=%v", ni)
		log.Printf("ni=%v", ni)

		table := "tpch6" + strconv.Itoa(ni)
		dest := "result"
		destType := "sum"
		ind := ads.IndexInfo{
			Table:    table,
			Dest:     dest,
			DestType: destType,
		}

		ind.Cond = []string{"shipdate"}
		start := time.Now()
		err := ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed := time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"discount"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//fmt.Printf("setup runtime for Tpch6 discount and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for Tpch6 discount and N=2^%v: %v\n", ni, elapsed)
		if err != nil {
			log.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"quantity"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//fmt.Printf("setup runtime for Tpch6 quantity and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for Tpch6 quantity and N=2^%v: %v\n", ni, elapsed)
		if err != nil {
			log.Printf("error: %v\n", err)
		}
	}
}

func CreateIndexTest() {
	runtime.GOMAXPROCS(64)
	logFile, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_index_eq.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)
	log.Printf("loading pub key")
	MAXN := 1<<20 + 100
	pk1, pk2 := acc.LoadPubkey(MAXN)
	for _, ni := range []int{15, 16, 17, 18, 19, 20} {
		// for _, ni := range []int{20} {
		logger.Printf("ni=%v", ni)
		log.Printf("ni=%v", ni)

		table := "test" + strconv.Itoa(ni)
		dest := "result"
		destType := "sum"
		ind := ads.IndexInfo{
			Table:    table,
			Dest:     dest,
			DestType: destType,
		}

		var start time.Time
		var elapsed time.Duration
		var err error

		ind.Cond = []string{"addr"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [addr->result] (eq 1<<8): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"bank"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [bank->result] (eq 1<<10): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"name"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [name->result] (eq 1<<12): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"grade"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [grade->result] (range 1<<8): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"rate"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [rate->result] (range 1<<10): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"value"}
		start = time.Now()
		err = ads.CreateSingleCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		//log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
		logger.Printf("setup runtime for test_%v [value->result] (range 1<<12): %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

	}
}

func CreateMultiIndexTest() {
	runtime.GOMAXPROCS(64)
	logFile, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_multi_index.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	logFile1, err := os.OpenFile(filepath.Join(acc.BaseDir, "test_result", "runtime_create_multi_details.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("open log file error: %v", err)
	}
	defer logFile.Close()
	logger1 := log.New(logFile1, "", log.LstdFlags)

	log.Printf("loading pub key")
	MAXN := 1<<20 + 100
	pk1, pk2 := acc.LoadPubkey(MAXN)
	for _, ni := range []int{15, 16, 17, 18, 19, 20} {
		//for _, ni := range []int{19} {
		logger.Printf("ni=%v", ni)
		logger1.Printf("ni=%v", ni)

		table := "test" + strconv.Itoa(ni)
		dest := "result"
		destType := "sum"
		ind := ads.IndexInfo{
			Table:    table,
			Dest:     dest,
			DestType: destType,
		}

		var start time.Time
		var elapsed time.Duration
		var err error

		ind.Cond = []string{"name", "value"}
		start = time.Now()
		err = ads.CreateMultiCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		logger.Printf("setup runtime for test_%v [(name,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		logger1.Printf("setup runtime for test_%v [(name,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"name", "bank"}
		start = time.Now()
		err = ads.CreateMultiCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		logger.Printf("setup runtime for test_%v [(name,bank)->result]: %ds\n", ni, int(elapsed.Seconds()))
		logger1.Printf("setup runtime for test_%v [(name,bank)->result]: %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"name", "bank", "value"}
		start = time.Now()
		err = ads.CreateMultiCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		logger.Printf("setup runtime for test_%v [(name,bank,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		logger1.Printf("setup runtime for test_%v [(name,bank,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"name", "bank", "addr"}
		start = time.Now()
		err = ads.CreateMultiCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		logger.Printf("setup runtime for test_%v [(name,bank,addr)->result]: %ds\n", ni, int(elapsed.Seconds()))
		logger1.Printf("setup runtime for test_%v [(name,bank,addr)->result]: %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

		ind.Cond = []string{"name", "bank", "addr", "value"}
		start = time.Now()
		err = ads.CreateMultiCondIndex(ads.DBINFO, ind, pk1, pk2)
		elapsed = time.Since(start)
		logger.Printf("setup runtime for test_%v [(name,bank,addr,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		logger1.Printf("setup runtime for test_%v [(name,bank,addr,value)->result]: %ds\n", ni, int(elapsed.Seconds()))
		if err != nil {
			logger.Printf("error: %v\n", err)
		}

	}
}
