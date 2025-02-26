package performance

import (
	"fmt"
	"log"
	"time"

	acc "contractdb/accumulator"
	"contractdb/ads"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/exp/rand"
)

func TestProveSum() {
	// test cost of ProveSum
	N := 1 << 13
	pk1, pk2 := acc.LoadPubkey(N)

	randListMap := make(map[uint64]int)
	for i := 0; i < N; i++ {
		rval := rand.Uint64()
		if _, exists := randListMap[rval]; !exists {
			randListMap[rval] = 1
		}
	}
	var randList []uint64
	for k := range randListMap {
		randList = append(randList, k)
	}

	for _, ni := range []int{5, 6, 7, 8, 9, 10, 11, 12} {
		n := 1 << ni
		offset := rand.Intn(len(randList) - n)
		S := make(fr.Vector, n)
		for i := range n {
			S[i].SetUint64(randList[i+offset])
		}

		start := time.Now()
		acc.ProveSum(S, pk1, pk2)
		elapsed := time.Since(start)
		fmt.Printf("Query time for single equivalent and N=2^%v: %v\n", ni, elapsed)
	}
}

func RuntimeQueryWithIntersection() {
	fmt.Print("loading public key ...\n")
	N := 1<<20 + 10
	pk1, pk2 := acc.LoadPubkey(N)

	// warmup
	fmt.Printf("warm up ... \n")
	q := genQueryTest20(1)
	ads.QueryDB(ads.DBINFO, q, pk1, pk2)
	ads.QueryDB(ads.DBINFO, q, pk1, pk2)
	fmt.Printf("finished \n ")

	var repeat int

	for _, ind := range []int{1, 2, 3, 10, 11, 12, 13, 20, 21, 22, 23, 30, 31, 32, 33} {

		if ind == 2 || ind == 3 {
			repeat = 1
		} else {
			repeat = 20
		}

		fmt.Printf("test query runtime for k_eq=%v and k_range=%v ...\n", ind/10, ind%10)

		q := genQueryTest20(ind)

		var res interface{}
		var err error
		var sumTime, interTime, sumTime1, interTime1 float64
		sumTime, interTime = 0, 0

		for i := 0; i < repeat; i++ {
			res, err, sumTime1, interTime1 = ads.QueryDBOutputTime(ads.DBINFO, q, pk1, pk2)
			sumTime += sumTime1
			interTime += interTime1
		}

		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		sum := res.(ads.SumRes1d).Sum
		fmt.Printf("sumTime=%.2fs, interTime=%.2f, query result: %v\n", sumTime/float64(repeat), interTime/float64(repeat), sum.String())

		fmt.Printf("test verify runtime for k_eq=%v and k_range=%v ...\n", ind/10, ind%10)
		var ver bool
		start := time.Now()
		for i := 0; i < repeat; i++ {
			ver, err = ads.Verify(ads.DBINFO, q, res, pk1, pk2)
		}
		elapsed := time.Since(start)
		fmt.Printf("runtime=%.2fs, verify result: %v with error %v\n", elapsed.Seconds()/float64(repeat), ver, err)
	}
}

func RuntimeQueryWithMultiCond() {
	fmt.Print("loading public key ...\n")
	N := 1<<20 + 10
	pk1, pk2 := acc.LoadPubkey(N)

	log.Printf("warmup")
	q := genQueryMultiTest20(0)
	ads.QueryDBMulti(ads.DBINFO, q, pk1, pk2)
	log.Printf("ok")
	repeat := 20

	for _, ind := range []int{0, 1, 2, 3, 4} {
		//for _, ind := range []int{0, 0} {
		fmt.Printf("test query runtime for multi-cond index %v ...\n", ind)
		q := genQueryMultiTest20(ind)
		start := time.Now()
		var res interface{}
		var err error
		for i := 0; i < repeat; i++ {
			res, err = ads.QueryDBMulti(ads.DBINFO, q, pk1, pk2)
		}
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("err: %v\n", err)
			return
		}
		sum := res.(ads.SumRes1d).Sum
		fmt.Printf("runtime=%.2fs, query result: %v\n", elapsed.Seconds()/float64(repeat), sum.String())

		fmt.Printf("test verify runtime for multi-cond index %v ...\n", ind)
		start = time.Now()

		var ver bool
		for i := 0; i < repeat; i++ {
			ver, err = ads.VerifyMulti(ads.DBINFO, q, res, pk1, pk2)
		}
		elapsed = time.Since(start)

		fmt.Printf("runtime=%.2fs, verify result: %v with error %v\n", elapsed.Seconds()/float64(repeat), ver, err)
	}
}

func genQueryMultiTest20(ind int) ads.Query {
	q := ads.Query{
		Table:    "test20",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	// to control middle result set of range query is around 8000
	// let range size of (value, rate, grade) be (30, 7, 1)
	switch ind {
	case 0: // name,value
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name2951", []uint64{3000, 3030}}
	case 1: // bank,name
		q.Cond = []string{"bank", "name"}
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Bank532", "Name2951"}
	case 2: // bank,name,value
		q.Cond = []string{"bank", "name", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Bank532", "Name2951", []uint64{4570, 5000}}
	case 3: // addr,bank,name
		q.Cond = []string{"addr", "bank", "name"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Addr129", "Bank532", "Name2951"}
	case 4: // addr,bank,name,value
		q.Cond = []string{"addr", "bank", "name", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Addr129", "Bank532", "Name2951", []uint64{4570, 5000}}
	default:
		panic("invalid input type")
	}
	return q
}

func genQueryMultiTest19(ind int) ads.Query {
	q := ads.Query{
		Table:    "test19",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	// to control middle result set of range query is around 8000
	// let range size of (value, rate, grade) be (30, 7, 1)
	switch ind {
	case 0: // name,value
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name2529", []uint64{3370, 3400}}
	case 1: // bank,name
		q.Cond = []string{"bank", "name"}
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Bank810", "Name2529"}
	case 2: // bank,name,value
		q.Cond = []string{"bank", "name", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Bank810", "Name2529", []uint64{3370, 3400}}
	case 3: // addr,bank,name
		q.Cond = []string{"addr", "bank", "name"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Addr100", "Bank810", "Name2529"}
	case 4: // addr,bank,name,value
		q.Cond = []string{"addr", "bank", "name", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Addr100", "Bank810", "Name2529", []uint64{3370, 3400}}
	default:
		panic("invalid input type")
	}
	return q
}

func genQueryTest20(ind int) ads.Query {
	q := ads.Query{
		Table:    "test20",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	// to control middle result set of range query is around 8000
	// let range size of (value, rate, grade) be (30, 7, 1)
	switch ind {
	case 1: //01
		q.Cond = []string{"value"}
		q.CondFlag = []int{1}
		q.CondVal = []interface{}{[]uint64{1000, 1030}}
	case 2: //02
		q.Cond = []string{"value", "rate"}
		q.CondFlag = []int{1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 1030}, []uint64{2000, 2007}}
	case 3: //03
		q.Cond = []string{"value", "rate", "grade"}
		q.CondFlag = []int{1, 1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 1030}, []uint64{2000, 2007}, []uint64{3000, 3001}}
	case 10:
		q.Cond = []string{"name"}
		q.CondFlag = []int{0}
		q.CondVal = []interface{}{"Name2951"}
	case 11:
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name2951", []uint64{3000, 3030}}
	case 12:
		q.Cond = []string{"name", "value", "rate"}
		q.CondFlag = []int{0, 1, 1}
		q.CondVal = []interface{}{"Name2951", []uint64{3000, 3030}, []uint64{2343, 2350}}
	case 13:
		q.Cond = []string{"name", "value", "rate", "grade"}
		q.CondFlag = []int{0, 1, 1, 1}
		q.CondVal = []interface{}{"Name2951", []uint64{3000, 3030}, []uint64{2343, 2350}, []uint64{3087, 3088}}
	case 20:
		q.Cond = []string{"name", "bank"} // addr and bank are wrong
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Name2951", "Bank532"}
	case 21:
		q.Cond = []string{"name", "bank", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", []uint64{4570, 5000}}
	case 22:
		q.Cond = []string{"name", "bank", "value", "rate"}
		q.CondFlag = []int{0, 0, 1, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", []uint64{4570, 5000}, []uint64{2113, 2120}}
	case 23:
		q.Cond = []string{"name", "bank", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", []uint64{4570, 5000}, []uint64{2113, 2120}, []uint64{3244, 3245}}
	case 30:
		q.Cond = []string{"name", "bank", "addr"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Name2951", "Bank532", "Addr129"}
	case 31:
		q.Cond = []string{"name", "bank", "bank", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", "Addr129", []uint64{4570, 5000}}
	case 32:
		q.Cond = []string{"name", "bank", "addr", "value", "rate"}
		q.CondFlag = []int{0, 0, 0, 1, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", "Addr129", []uint64{4570, 5000}, []uint64{2113, 2120}}
	case 33:
		q.Cond = []string{"name", "bank", "addr", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Name2951", "Bank532", "Addr129", []uint64{4570, 5000}, []uint64{2113, 2120}, []uint64{3244, 3245}}
	default:
		panic("invalid input type")
	}
	return q
}
