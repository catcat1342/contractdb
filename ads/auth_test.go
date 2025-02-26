package ads

import (
	acc "contractdb/accumulator"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestTPCH6(t *testing.T) {
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer f.Close()
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	MAXN := 1<<15 + 100
	pk1, pk2 := acc.LoadPubkey(MAXN)

	ni := 10
	table := "tpch6" + strconv.Itoa(ni)
	dest := "result"
	destType := "sum"
	ind := IndexInfo{
		Table:    table,
		Dest:     dest,
		DestType: destType,
	}

	var start time.Time
	var elapsed time.Duration
	var err error

	ind.Cond = []string{"shipdate"}
	start = time.Now()
	err = CreateSingleCondIndex(DBINFO, ind, pk1, pk2)
	elapsed = time.Since(start)
	fmt.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
	log.Printf("setup runtime for Tpch6 shipdate and N=2^%v: %v\n", ni, elapsed)
	if err != nil {
		log.Printf("error: %v\n", err)
	}

	ind.Cond = []string{"discount"}
	start = time.Now()
	err = CreateSingleCondIndex(DBINFO, ind, pk1, pk2)
	elapsed = time.Since(start)
	fmt.Printf("setup runtime for Tpch6 discount and N=2^%v: %v\n", ni, elapsed)
	log.Printf("setup runtime for Tpch6 discount and N=2^%v: %v\n", ni, elapsed)
	if err != nil {
		log.Printf("error: %v\n", err)
	}

	ind.Cond = []string{"quantity"}
	start = time.Now()
	err = CreateSingleCondIndex(DBINFO, ind, pk1, pk2)
	elapsed = time.Since(start)
	fmt.Printf("setup runtime for Tpch6 quantity and N=2^%v: %v\n", ni, elapsed)
	log.Printf("setup runtime for Tpch6 quantity and N=2^%v: %v\n", ni, elapsed)
	if err != nil {
		log.Printf("error: %v\n", err)
	}

}

func TestCreateMultiCondIndex(t *testing.T) {
	runtime.GOMAXPROCS(52)

	ni := 18

	N := 1<<ni + 100
	pk1, pk2 := acc.LoadPubkey(N)

	conds := [][]string{
		{"name", "value"},
		{"name", "bank"},
		{"name", "bank", "value"},
		{"name", "bank", "addr"},
		{"name", "bank", "addr", "value"},
	}
	dest := "result"
	destType := "sum"

	table := "test" + strconv.Itoa(ni)
	ind := IndexInfo{
		Table:    table,
		Dest:     dest,
		DestType: destType,
	}
	//for _, c := range conds {
	for _, c := range conds[1:2] {
		ind.Cond = c
		fmt.Printf("create index on cond: %v...\n", c)
		err := CreateMultiCondIndex(DBINFO, ind, pk1, pk2)
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		fmt.Printf("ok\n")
	}

}

func TestQueryMultiCond(t *testing.T) {
	N := 1 << 12
	pk1, pk2 := acc.LoadPubkey(N)

	// for _, num := range []int{11, 20, 21, 30, 31} {
	// 	q := genQueryTest10(num)
	// 	res, err := QueryDBMulti(DBINFO, q, pk1, pk2)
	// 	if err != nil {
	// 		log.Printf("query db failed: %v\n", err)
	// 		return
	// 	}
	// 	sum := res.(SumRes1d).Sum
	// 	fmt.Printf("result: %v\n", sum.String())
	// 	ver, err := VerifyMulti(DBINFO, q, res, pk1, pk2)
	// 	if err != nil || !ver {
	// 		log.Printf("verify query failed: %v\n", err)
	// 	} else {
	// 		fmt.Printf("verify query: true\n")
	// 	}
	// }

	q := Query{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
		Cond:     []string{"name", "value"},
		CondFlag: []int{0, 1},
		CondVal:  []interface{}{"Name3948", []uint64{4500, 4800}},
	}
	res, err := QueryDBMulti(DBINFO, q, pk1, pk2)
	if err != nil {
		log.Printf("query db failed: %v\n", err)
		return
	}
	sum := res.(SumRes1d).Sum
	fmt.Printf("result: %v\n", sum.String())
	ver, err := VerifyMulti(DBINFO, q, res, pk1, pk2)
	if err != nil || !ver {
		log.Printf("verify query failed: %v\n", err)
	} else {
		fmt.Printf("verify query: true\n")
	}
}

func TestWriteLeveldb(t *testing.T) {
	var datalist []DBData
	N := 1 << 10
	for i := range N {
		datalist = append(datalist, DBData{[]byte("key" + strconv.Itoa(i)), []byte("value" + strconv.Itoa(i))})
	}

	log.Printf("writeLevelDB()")
	start := time.Now()
	writeLevelDB(filepath.Join(acc.BaseDir, "test_result", "test1"), datalist)
	elapsed := time.Since(start)
	log.Printf("writeLevelDB ok: %.2fs\n", elapsed.Seconds())
}

func TestCreateSingleCondIndex(t *testing.T) {
	N := 1 << 12
	pk1, pk2 := acc.LoadPubkey(N)
	table := "test10"
	cond := []string{"name", "bank", "addr", "value", "rate", "grade"}
	//cond := []string{"bank"}
	dest := "result"
	destType := "sum"
	ind := IndexInfo{
		Table:    table,
		Cond:     cond,
		Dest:     dest,
		DestType: destType,
	}
	// flag := "SingleEq"
	err := CreateSingleCondIndex(DBINFO, ind, pk1, pk2)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
}

func TestQuerySingleCond(t *testing.T) {
	N := 1<<12 + 10
	pk1, pk2 := acc.LoadPubkey(N)
	q := genQueryTest10(20)
	res, err := QueryDB(DBINFO, q, pk1, pk2)
	if err != nil {
		log.Printf("query db failed: %v\n", err)
		return
	}
	sum := res.(SumRes1d).Sum
	fmt.Printf("result: %v\n", sum.String())
	ver, err := Verify(DBINFO, q, res, pk1, pk2)
	if err != nil || !ver {
		fmt.Printf("verify query failed: %v\n", err)
	} else {
		fmt.Printf("verify query: true\n")
	}
}

func genQueryTest10(ind int) Query {
	q := Query{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	switch ind {
	case 01:
		q.Cond = []string{"value"}
		q.CondFlag = []int{1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}}
	case 02:
		q.Cond = []string{"value", "rate"}
		q.CondFlag = []int{1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}, []uint64{2000, 2500}}
	case 03:
		q.Cond = []string{"value", "rate", "grade"}
		q.CondFlag = []int{1, 1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}, []uint64{2000, 2500}, []uint64{3000, 3100}}
	case 10:
		q.Cond = []string{"name"}
		q.CondFlag = []int{0}
		q.CondVal = []interface{}{"Name3948"}
	case 11:
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}}
	case 12:
		q.Cond = []string{"name", "value", "rate"}
		q.CondFlag = []int{0, 1, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 13:
		q.Cond = []string{"name", "value", "rate", "grade"}
		q.CondFlag = []int{0, 1, 1, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	case 20:
		q.Cond = []string{"bank", "name"}
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Bank564", "Name3948"}
	case 21:
		q.Cond = []string{"bank", "name", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Bank564", "Name3948", []uint64{4000, 5000}}
	case 22:
		q.Cond = []string{"bank", "name", "value", "rate"}
		q.CondFlag = []int{0, 0, 1, 1}
		q.CondVal = []interface{}{"Bank564", "Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 23:
		q.Cond = []string{"bank", "name", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Bank564", "Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	case 30:
		q.Cond = []string{"addr", "bank", "name"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948"}
	case 31:
		q.Cond = []string{"addr", "bank", "name", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948", []uint64{4000, 5000}}
	case 32:
		q.Cond = []string{"addr", "bank", "name", "value", "rate"}
		q.CondFlag = []int{0, 0, 0, 1, 1}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 33:
		q.Cond = []string{"addr", "bank", "name", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	default:
		panic("invalid input type")
	}
	return q
}

func TestMaxmin(t *testing.T) {
	pk1, pk2 := acc.LoadPubkey(100)
	q := Query{
		Table:    "test10",
		Cond:     []string{"rate"},
		CondVal:  []interface{}{[]uint64{359744, 408005}},
		CondFlag: []int{1},
		Dest:     "value", // add keyid at first, use fixed sequence
		DestType: "min",
	}
	res, err := QueryDB(DBINFO, q, pk1, pk2)
	if err != nil {
		log.Printf("query db failed: %v\n", err)
		return
	}

	fmt.Printf("query result: %v\n", getResultStr(res))
	ver, err := Verify(DBINFO, q, res, pk1, pk2)
	fmt.Printf("ver: %v, err: %v\n", ver, err)
}
