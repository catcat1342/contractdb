package ads

import (
	"bytes"
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/syndtr/goleveldb/leveldb"
)

type Query struct {
	Table    string
	Cond     []string
	CondVal  []interface{} // string or uint
	CondFlag []int
	Dest     string
	DestType string // "sum", "min", "max", "count", "avg"
	// CondFlag: 0 for =, 1 for [vl,vr], 2 for (vl,vr], 3 for [vl, vr), 4 for (vl,vr)
	// < vr: (0,vr); <= vr: (0,vr]; > vl: (vl,INF); >=vl: [vl,INF)
}

type AuthQuery struct {
	AuthDB string
	Key    fr.Element
	Flags  []int // Flags[0] for Hkey; Flags[1] for Hnxt (0:<; 1:<=, 2:= 3:>=; 4:>)
}

type SumRes1d struct {
	Sum      fr.Element
	FR       bn.G1Affine
	SumProof acc.SumProof
	DicItems []acc.DicItem
	IProof   acc.IntersectionProof
}

func (q *Query) IsValid() (bool, error) {
	if len(q.Cond) != len(q.CondVal) || len(q.Cond) != len(q.CondFlag) {
		return false, fmt.Errorf("invalid q: length of index, CondVal, and CondFlag not match")
	}

	for i := range q.Cond {
		switch q.CondFlag[i] {
		case 0:
			if _, ok := q.CondVal[i].(string); !ok {
				return false, fmt.Errorf("invalid q: invalid CondVal (%v) for equivalent query", q.CondVal[i])
			}
		case 1, 2, 3, 4:
			qv, ok := q.CondVal[i].([]uint64)
			if !ok {
				return false, fmt.Errorf("invalid q: invalid CondVal (%v) for range query", q.CondVal[i])
			}
			if qv[0] > qv[1] {
				return false, fmt.Errorf("invalid q: invalid CondVal (%v,%v) for range query, left border should le right border", qv[0], qv[1])
			}
		default:
			return false, fmt.Errorf("invalid q: unspported flag %v", q.CondFlag[i])
		}
	}
	return true, nil
}

func QueryDB(
	dbinfo string,
	q Query,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (interface{}, error) {
	// each dest has a bn.FP||SumQueryProof as a result
	db, err := sql.Open("mysql", dbinfo)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// prepare: check the correctness of q
	if !CheckTableExist(db, []string{q.Table}) {
		return nil, fmt.Errorf("invalid q: table %v not exist", q.Table)
	}
	if ok, err := q.IsValid(); !ok {
		return nil, err
	}
	var result interface{}

	switch q.DestType {
	case "sum":
		result, err = querySum1d(db, q, pk1, pk2)
	case "max", "min":
		result, err = queryMaxmin1d(db, q, pk1, pk2)
	case "count":
		// return QueryDBCount(db, q, pk1, pk2)
		return nil, fmt.Errorf("invalid query type %v ", q.DestType)
	case "avg":
		// return QueryDBAvg(db, q, pk1, pk2)
		return nil, fmt.Errorf("invalid query type %v ", q.DestType)
	default:
		return nil, fmt.Errorf("invalid query type %v ", q.DestType)
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func QueryDBOutputTime(
	dbinfo string,
	q Query,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (interface{}, error, float64, float64) {
	// each dest has a bn.FP||SumQueryProof as a result
	db, err := sql.Open("mysql", dbinfo)
	if err != nil {
		return nil, err, 0, 0
	}
	defer db.Close()

	// prepare: check the correctness of q
	if !CheckTableExist(db, []string{q.Table}) {
		return nil, fmt.Errorf("invalid q: table %v not exist", q.Table), 0, 0
	}
	if ok, err := q.IsValid(); !ok {
		return nil, err, 0, 0
	}
	var result interface{}

	var sumTime, interTime float64

	switch q.DestType {
	case "sum":
		result, err, sumTime, interTime = querySum1dOutputTime(db, q, pk1, pk2)
	case "max", "min":
		result, err = queryMaxmin1d(db, q, pk1, pk2)
	case "count":
		// return QueryDBCount(db, q, pk1, pk2)
		return nil, fmt.Errorf("invalid query type %v ", q.DestType), 0, 0
	case "avg":
		// return QueryDBAvg(db, q, pk1, pk2)
		return nil, fmt.Errorf("invalid query type %v ", q.DestType), 0, 0
	default:
		return nil, fmt.Errorf("invalid query type %v ", q.DestType), 0, 0
	}
	if err != nil {
		return nil, err, 0, 0
	}

	return result, nil, sumTime, interTime
}

func querySum1d(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (SumRes1d, error) {
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return SumRes1d{}, err
	}
	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)

	var dicItem []acc.DicItem
	var middleSet []fr.Vector

	// for each cond, generate dic item and wit
	for c := range q.Cond {
		switch q.CondFlag[c] {
		case 0:
			mset, item, err := query1c1dEq(db, q, c)
			if err != nil {
				return SumRes1d{}, err
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, item)
		case 1, 2, 3, 4:
			mset, items, err := query1c1dRange(db, q, c)
			if err != nil {
				return SumRes1d{}, err
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, items...)
		default:
			return SumRes1d{}, fmt.Errorf("invalid flag")
		}
	}

	// fmt.Printf("Test rset == Intersect(middleSet)\n")
	// I := acc.Intersection(middleSet)
	// fmt.Printf("len(rset): %v, len(I): %v\n", len(rset), len(I))

	// itersect middle results
	iproof := acc.ProveIntersection(middleSet, pk1, pk2)
	// test
	// acc.VerifyIntersection(iproof, pk1)

	res := SumRes1d{
		Sum:      sum,
		FR:       fR,
		SumProof: sumProof,
		DicItems: dicItem,
		IProof:   iproof,
	}
	return res, nil
}

func querySum1dOutputTime(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (SumRes1d, error, float64, float64) {

	start := time.Now()
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return SumRes1d{}, err, 0, 0
	}
	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)
	elapsed := time.Since(start)
	sumTime := elapsed.Seconds()

	var dicItem []acc.DicItem
	var middleSet []fr.Vector

	// for each cond, generate dic item and wit
	start = time.Now()
	for c := range q.Cond {
		switch q.CondFlag[c] {
		case 0:
			mset, item, err := query1c1dEq(db, q, c)
			if err != nil {
				return SumRes1d{}, err, 0, 0
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, item)
		case 1, 2, 3, 4:
			mset, items, err := query1c1dRange(db, q, c)
			if err != nil {
				return SumRes1d{}, err, 0, 0
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, items...)
		default:
			return SumRes1d{}, fmt.Errorf("invalid flag"), 0, 0
		}
	}
	iproof := acc.ProveIntersection(middleSet, pk1, pk2)
	elapsed = time.Since(start)
	interTime := elapsed.Seconds()

	res := SumRes1d{
		Sum:      sum,
		FR:       fR,
		SumProof: sumProof,
		DicItems: dicItem,
		IProof:   iproof,
	}
	return res, nil, sumTime, interTime
}

func query1c1dEq(db *sql.DB, q Query, c int) (mset fr.Vector, item acc.DicItem, err error) {
	if q.CondFlag[c] != 0 {
		return mset, item, fmt.Errorf("invalid flag")
	}
	mset, err = querySet1d(db, q, c)
	if err != nil {
		return mset, item, err
	}

	ind := IndexInfo{q.Table, q.Cond[c : c+1], q.Dest, q.DestType}
	key := new(fr.Element).SetBytes([]byte(q.CondVal[c].(string)))

	aq := AuthQuery{
		AuthDB: ind.AuthTable(),
		Key:    *key,
	}
	if len(mset) == 0 {
		aq.Flags = []int{0, 0} // hkey<key<hnxt
	} else {
		aq.Flags = []int{2, 0} // hkey=key<hnxt
	}
	item, err = queryAuthDB(aq)
	if err != nil {
		return mset, item, fmt.Errorf("db.Query: %v", err)
	}

	return mset, item, nil
}

func query1c1dRange(db *sql.DB, q Query, c int) (mset fr.Vector, item []acc.DicItem, err error) {
	if q.CondFlag[c] != 1 && q.CondFlag[c] != 2 && q.CondFlag[c] != 3 && q.CondFlag[c] != 4 {
		return nil, nil, fmt.Errorf("invalid flag")
	}
	mset, err = querySet1d(db, q, c)
	if err != nil {
		return nil, nil, err
	}
	vl, vr := new(fr.Element).SetUint64(q.CondVal[c].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[c].([]uint64)[1])

	ind := &IndexInfo{q.Table, q.Cond[c : c+1], q.Dest, q.DestType}

	aqLeft := AuthQuery{AuthDB: ind.AuthTable(), Key: *vl}
	aqRight := AuthQuery{AuthDB: ind.AuthTable(), Key: *vr}

	switch q.CondFlag[c] {
	case 1: // [vl,vr] aqLeft: itemkey<vl<=itemnxt, aqRight: itemkey<=vr<itemnxt
		aqLeft.Flags = []int{0, 1}
		aqRight.Flags = []int{1, 0}
	case 2: // (vl,vr] aqLeft: itkey<=vl<itnxt, aqRight: itkey<=vr<itnxt
		aqLeft.Flags = []int{1, 0}
		aqRight.Flags = []int{1, 0}
	case 3: // [vl,vr) aqLeft: itkey<vl<=itnxt, aqRight: itkey<vr<=itnxt
		aqLeft.Flags = []int{0, 1}
		aqRight.Flags = []int{0, 1}
	case 4: // (vl,vr) aqLeft: itkey<=vl<itnxt, aqRight: itkey<vr<=itnxt
		aqLeft.Flags = []int{1, 0}
		aqRight.Flags = []int{0, 1}
	}

	itemLeft, err := queryAuthDB(aqLeft)
	if err != nil {
		return nil, nil, fmt.Errorf("load authDB error, unsupported type")
	}
	itemRight, err := queryAuthDB(aqRight)
	if err != nil {
		return nil, nil, fmt.Errorf("load authDB error, unsupported type")
	}

	item = append(item, itemLeft, itemRight)
	return mset, item, nil
}

// query the cc-th condition
// if cc==len(q.Cond), query the intersection on all conditions
func querySet1d(db *sql.DB, q Query, cc int) (rset fr.Vector, err error) {
	var selectQuery string
	switch q.DestType {
	case "sum": // sum: query all results to generate sum proof, do not query sum directly
		selectQuery = fmt.Sprintf("SELECT %v FROM %v WHERE ", q.Dest, q.Table)
	case "max":
		selectQuery = fmt.Sprintf("SELECT MAX(%v) FROM %v WHERE ", q.Dest, q.Table)
	case "min":
		selectQuery = fmt.Sprintf("SELECT MIN(%v) FROM %v WHERE ", q.Dest, q.Table)
	default:
		return fr.Vector{}, fmt.Errorf("querySet1d do not support type %v yet", q.DestType)
	}

	for i := 0; i < len(q.Cond); i++ { // generate condition
		if cc < len(q.Cond) && cc != i {
			continue
		}
		switch q.CondFlag[i] {
		case 0:
			selectQuery += fmt.Sprintf("%v = \"%v\"", q.Cond[i], q.CondVal[i].(string))
		case 1: // [vl,vr]
			vleft, vright := q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1]
			selectQuery += fmt.Sprintf("%v BETWEEN %v AND %v", q.Cond[i], vleft, vright)
		case 2: // (vl,vr]
			vleft, vright := q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1]
			selectQuery += fmt.Sprintf("%v > %v AND %v <= %v", q.Cond[i], vleft, q.Cond[i], vright)
		case 3: // [vl,vr)
			vleft, vright := q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1]
			selectQuery += fmt.Sprintf("%v >= %v AND %v < %v", q.Cond[i], vleft, q.Cond[i], vright)
		case 4: // (vl,vr)
			vleft, vright := q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1]
			selectQuery += fmt.Sprintf("%v > %v AND %v < %v", q.Cond[i], vleft, q.Cond[i], vright)
		default:
			return fr.Vector{}, fmt.Errorf("invalid flag type of query")
		}
		if cc >= len(q.Cond) && i < len(q.Cond)-1 {
			selectQuery += " AND "
		}
	}

	//log.Printf("query database: [%v] ...\n", selectQuery)
	rows, err := db.Query(selectQuery)
	if err != nil {
		return fr.Vector{}, fmt.Errorf("error in db.Query: %v", err)
	}
	defer rows.Close()

	var rval []byte
	for rows.Next() {
		err := rows.Scan(&rval)
		if err != nil {
			return fr.Vector{}, fmt.Errorf("error scanning row: %v", err)
		}
		if len(rval) == 0 {
			return fr.Vector{}, nil
		}
		rval, err := new(fr.Element).SetString(string(rval))
		if err != nil {
			return fr.Vector{}, fmt.Errorf("fr.Element setstring error: %v", err)
		}
		rset = append(rset, *rval)
	}
	return rset, nil
}

func queryAuthDB(aq AuthQuery) (dic acc.DicItem, err error) {
	authDB := filepath.Join(acc.BaseDir, "authdb", aq.AuthDB)
	db, err := leveldb.OpenFile(authDB, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return dic, err
	}

	hkey := aq.Key.Bytes()
	var val []byte
	iter := db.NewIterator(nil, nil)
	iter.Seek(hkey[:])

	switch aq.Flags[0] {
	case 0: // <
		if !iter.Valid() || bytes.Compare(iter.Key(), hkey[:]) >= 0 {
			if !iter.Prev() {
				return dic, fmt.Errorf("queryAuthDB error")
			}
		}
		val = iter.Value()
	case 1: // <=
		if !iter.Valid() || bytes.Compare(iter.Key(), hkey[:]) > 0 {
			if !iter.Prev() {
				return dic, fmt.Errorf("queryAuthDB error")
			}
		}
		val = iter.Value()
	case 2: // =
		if !iter.Valid() {
			return dic, fmt.Errorf("queryAuthDB error")
		}
		val = iter.Value()
	default:
		return dic, fmt.Errorf("queryAuthDB error: invalid flag")
	}
	iter.Release()
	db.Close()
	// parse val to dicItem
	dic = stringToDicItem(string(val))
	if dic.Value == nil {
		return dic, fmt.Errorf("queryAuthDB error: required item not exist")
	}

	switch aq.Flags[1] {
	case 0: // k<nxt
		if !(aq.Key.Cmp(&dic.Nxt) == -1) {
			return acc.DicItem{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
	case 1: // k<=nxt
		if !(aq.Key.Cmp(&dic.Nxt) != 1) {
			return acc.DicItem{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
	case 2: // k==nxt
		if !(aq.Key.Equal(&dic.Nxt)) {
			return acc.DicItem{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
	default:
		return acc.DicItem{}, fmt.Errorf("queryAuthDB error: required item not exist")
	}

	return dic, nil
}

func stringToDicItem(val string) acc.DicItem {
	if val == "" {
		return acc.DicItem{}
	}
	ss := strings.Split(val, "|")
	if len(ss) <= 1 {
		return acc.DicItem{}
	}
	var item acc.DicItem
	key, _ := new(fr.Element).SetString(ss[0])
	nxt, _ := new(fr.Element).SetString(ss[1])
	item.Key = *key
	item.Nxt = *nxt

	if len(ss[2]) < 200 {
		item.Value = acc.StringToG1Affine(ss[2])
	} else {
		item.Value = acc.StringToG2Affine(ss[2])
	}

	if len(ss) >= 4 { // val may have no witness field
		item.W = acc.StringToG2Affine(ss[3])
	}

	//fmt.Printf("item: (%x,%x) %v\n", item.Key.Marshal(), item.Nxt.Marshal(), item.ValueString())
	return item
}

func queryAuthTable(db *sql.DB, authQuery string, flag int) (dic acc.DicItem, err error) {
	rows, err := db.Query(authQuery)
	if err != nil {
		return acc.DicItem{}, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	switch flag {
	case 0:
		var hkey, hnxt uint64
		var vx, vy, dx, dy, wx0, wx1, wy0, wy1 []byte
		rows.Next()
		err = rows.Scan(&hkey, &hnxt, &vx, &vy, &dx, &dy, &wx0, &wx1, &wy0, &wy1)
		if err != nil {
			return acc.DicItem{}, fmt.Errorf("load mysql line error: %v", err)
		}
		val, w := new(bn.G1Affine), new(bn.G2Affine)
		val.X.SetBytes(vx)
		val.Y.SetBytes(vy)
		w.X.A0.SetBytes(wx0)
		w.X.A1.SetBytes(wx1)
		w.Y.A0.SetBytes(wy0)
		w.Y.A1.SetBytes(wy1)
		dic = acc.DicItem{
			Key:   *new(fr.Element).SetUint64(hkey),
			Nxt:   *new(fr.Element).SetUint64(hnxt),
			Value: *val,
			W:     *w,
		}
		return dic, nil
	case 1, 2, 3, 4:
		var hkey, hnxt uint64
		var vx0, vx1, vy0, vy1, dx, dy, wx0, wx1, wy0, wy1 []byte
		rows.Next()
		err = rows.Scan(&hkey, &hnxt, &vx0, &vx1, &vy0, &vy1, &dx, &dy, &wx0, &wx1, &wy0, &wy1)
		if err != nil {
			return acc.DicItem{}, fmt.Errorf("load mysql line error: %v", err)
		}
		val, w := new(bn.G2Affine), new(bn.G2Affine)
		val.X.A0.SetBytes(vx0)
		val.X.A1.SetBytes(vx1)
		val.Y.A0.SetBytes(vy0)
		val.Y.A1.SetBytes(vy1)
		w.X.A0.SetBytes(wx0)
		w.X.A1.SetBytes(wx1)
		w.Y.A0.SetBytes(wy0)
		w.Y.A1.SetBytes(wy1)
		dic = acc.DicItem{
			Key:   *new(fr.Element).SetUint64(hkey),
			Nxt:   *new(fr.Element).SetUint64(hnxt),
			Value: *val,
			W:     *w,
		}
	default:
		return acc.DicItem{}, fmt.Errorf("invalid flag")
	}
	return dic, nil
}

func getResultStr(res interface{}) string {
	str := ""
	switch v := res.(type) {
	case SumRes1d:
		sum := v.Sum
		if sum.IsZero() {
			str += "NULL,"
		} else {
			str += sum.String() + ","
		}
	case MaxminRes1d:
		maxmin := v.Result
		if maxmin.IsZero() {
			str += "NULL,"
		} else {
			str += maxmin.String() + ","
		}
	default:
		return ""
	}
	return str[:len(str)-1]
}
