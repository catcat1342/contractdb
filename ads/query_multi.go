package ads

import (
	"bytes"
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"sort"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
)

func QueryDBMulti(
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
		switch len(q.Cond) {
		case 2:
			result, err = querySum2c1d(db, q, pk1, pk2)
		case 3:
			result, err = querySum3c1d(db, q, pk1, pk2)
		case 4:
			result, err = querySum4c1d(db, q, pk1, pk2)
		}
	default:
		return nil, fmt.Errorf("invalid query type %v for multi-cond query", q.DestType)
	}
	if err != nil {
		return nil, err
	}

	return result, nil
}

func querySum2c1d(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (SumRes1d, error) {
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return SumRes1d{}, err
	}
	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)
	var dicItem []acc.DicItem
	var tempItem acc.DicItem
	var iterKey []byte

	// // generate dic item and wit by multi-cond index (layer-index)
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return SumRes1d{}, err
	// }
	// // reorder conditions in q according to ind
	// tempMap1 := make(map[string]interface{})
	// tempMap2 := make(map[string]int)
	// for i, c := range q.Cond {
	// 	tempMap1[c] = q.CondVal[i]
	// 	tempMap2[c] = q.CondFlag[i]
	// }
	// q.Cond = ind.Cond
	// for i, c := range q.Cond {
	// 	q.CondVal[i] = tempMap1[c]
	// 	q.CondFlag[i] = tempMap2[c]
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}

	var emptyLayer int
	if len(rset) == 0 { // find which layer generates the empty result
		res0, _ := querySet1d(db, q, 0) // query only q.Cond[0]
		if len(res0) == 0 {
			emptyLayer = 0
		} else {
			emptyLayer = 1
		}
	} else {
		emptyLayer = len(q.Cond) + 100 // not empty
	}

	authDB, err := leveldb.OpenFile(filepath.Join(acc.BaseDir, "authdb", ind.AuthTable()), nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return SumRes1d{}, err
	}
	iter := authDB.NewIterator(nil, nil)

	v0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	aq0 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v0}
	if emptyLayer == 0 {
		aq0.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq0, nil)
		iterKey = iterKey[0:32]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		iter.Release()
		authDB.Close()
		dicItem = append(dicItem, tempItem)
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq0.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq0, nil)
	dicItem = append(dicItem, tempItem)

	if q.CondFlag[1] == 0 {
		v1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
		aq1 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v1}
		if len(rset) == 0 {
			aq1.Flags = []int{0, 0}
		} else {
			aq1.Flags = []int{2, 0}
		}
		// do not need to clip iterKey for the bottom layer
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aq1, v0.Marshal())
		dicItem = append(dicItem, tempItem)
	} else {
		v1l := new(fr.Element).SetUint64(q.CondVal[1].([]uint64)[0])
		v1r := new(fr.Element).SetUint64(q.CondVal[1].([]uint64)[1])
		aqLeft := AuthQuery{AuthDB: ind.AuthTable(), Key: *v1l}
		aqRight := AuthQuery{AuthDB: ind.AuthTable(), Key: *v1r}
		switch q.CondFlag[1] {
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
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqLeft, v0.Marshal())
		dicItem = append(dicItem, tempItem)
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqRight, v0.Marshal())
		dicItem = append(dicItem, tempItem)
	}

	iter.Release()
	authDB.Close()

	res := SumRes1d{
		Sum:      sum,
		FR:       fR,
		SumProof: sumProof,
		DicItems: dicItem,
	}
	return res, nil
}

func querySum3c1d(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (SumRes1d, error) {
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return SumRes1d{}, err
	}
	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)
	var dicItem []acc.DicItem
	var tempItem acc.DicItem
	var iterKey []byte

	// // generate dic item and wit by multi-cond index (layer-index)
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return SumRes1d{}, err
	// }
	// // reorder conditions in q
	// tempMap1 := make(map[string]interface{})
	// tempMap2 := make(map[string]int)
	// for i, c := range q.Cond {
	// 	tempMap1[c] = q.CondVal[i]
	// 	tempMap2[c] = q.CondFlag[i]
	// }
	// q.Cond = ind.Cond
	// for i, c := range q.Cond {
	// 	q.CondVal[i] = tempMap1[c]
	// 	q.CondFlag[i] = tempMap2[c]
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}

	var emptyLayer int
	if len(rset) == 0 { // find which layer generates the empty result
		res0, _ := querySet1d(db, q, 0) // query only q.Cond[0]
		if len(res0) == 0 {
			emptyLayer = 0
		} else {
			newq := Query{Table: q.Table, Cond: q.Cond[0:2], CondVal: q.CondVal[0:2], CondFlag: q.CondFlag[0:2], Dest: q.Dest, DestType: q.DestType}
			res1, _ := querySet1d(db, newq, len(newq.Cond))
			if len(res1) == 0 {
				emptyLayer = 1
			} else {
				emptyLayer = 2
			}
		}
	} else {
		emptyLayer = len(q.Cond) + 100
	}

	authDB, err := leveldb.OpenFile(filepath.Join(acc.BaseDir, "authdb", ind.AuthTable()), nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return SumRes1d{}, err
	}
	iter := authDB.NewIterator(nil, nil)

	v0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	aq0 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v0}
	if emptyLayer == 0 {
		aq0.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq0, nil)
		iterKey = iterKey[0:32]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		iter.Release()
		authDB.Close()
		dicItem = append(dicItem, tempItem)
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq0.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq0, nil)
	dicItem = append(dicItem, tempItem)

	v1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
	aq1 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v1}
	if emptyLayer == 1 {
		aq1.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq1, v0.Marshal())
		iterKey = iterKey[0:64]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		dicItem = append(dicItem, tempItem)
		iter.Release()
		authDB.Close()
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq1.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq1, v0.Marshal())
	dicItem = append(dicItem, tempItem)

	if q.CondFlag[2] == 0 {
		v2 := new(fr.Element).SetBytes([]byte(q.CondVal[2].(string)))
		aq2 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v2}
		if len(rset) == 0 {
			aq2.Flags = []int{0, 0}
		} else {
			aq2.Flags = []int{2, 0}
		}
		// do not need to clip iterKey for the bottom layer
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aq2, append(v0.Marshal(), v1.Marshal()...))
		dicItem = append(dicItem, tempItem)
	} else {
		v2l := new(fr.Element).SetUint64(q.CondVal[2].([]uint64)[0])
		v2r := new(fr.Element).SetUint64(q.CondVal[2].([]uint64)[1])
		aqLeft := AuthQuery{AuthDB: ind.AuthTable(), Key: *v2l}
		aqRight := AuthQuery{AuthDB: ind.AuthTable(), Key: *v2r}
		switch q.CondFlag[2] {
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
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqLeft, append(v0.Marshal(), v1.Marshal()...))
		dicItem = append(dicItem, tempItem)
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqRight, append(v0.Marshal(), v1.Marshal()...))
		dicItem = append(dicItem, tempItem)
	}

	iter.Release()
	authDB.Close()

	res := SumRes1d{
		Sum:      sum,
		FR:       fR,
		SumProof: sumProof,
		DicItems: dicItem,
	}
	return res, nil
}

func querySum4c1d(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (SumRes1d, error) {
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return SumRes1d{}, err
	}
	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)
	var dicItem []acc.DicItem
	var tempItem acc.DicItem
	var iterKey []byte

	// // generate dic item and wit by multi-cond index (layer-index)
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return SumRes1d{}, err
	// }
	// // reorder conditions in q
	// tempMap1 := make(map[string]interface{})
	// tempMap2 := make(map[string]int)
	// for i, c := range q.Cond {
	// 	tempMap1[c] = q.CondVal[i]
	// 	tempMap2[c] = q.CondFlag[i]
	// }
	// q.Cond = ind.Cond
	// for i, c := range q.Cond {
	// 	q.CondVal[i] = tempMap1[c]
	// 	q.CondFlag[i] = tempMap2[c]
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}

	var emptyLayer int
	if len(rset) == 0 { // find which layer generates the empty result
		res0, _ := querySet1d(db, q, 0) // query only q.Cond[0]
		if len(res0) == 0 {
			emptyLayer = 0
		} else {
			newq := Query{Table: q.Table, Cond: q.Cond[0:2], CondVal: q.CondVal[0:2], CondFlag: q.CondFlag[0:2], Dest: q.Dest, DestType: q.DestType}
			res1, _ := querySet1d(db, newq, len(newq.Cond))
			if len(res1) == 0 {
				emptyLayer = 1
			} else {
				newq := Query{Table: q.Table, Cond: q.Cond[0:3], CondVal: q.CondVal[0:3], CondFlag: q.CondFlag[0:3], Dest: q.Dest, DestType: q.DestType}
				res2, _ := querySet1d(db, newq, len(newq.Cond))
				if len(res2) == 0 {
					emptyLayer = 2
				} else {
					emptyLayer = 3
				}
			}
		}
	} else {
		emptyLayer = len(q.Cond) + 100
	}

	authDB, err := leveldb.OpenFile(filepath.Join(acc.BaseDir, "authdb", ind.AuthTable()), nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return SumRes1d{}, err
	}
	iter := authDB.NewIterator(nil, nil)

	v0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	aq0 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v0}
	if emptyLayer == 0 {
		aq0.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq0, nil)
		iterKey = iterKey[0:32]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		iter.Release()
		authDB.Close()
		dicItem = append(dicItem, tempItem)
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq0.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq0, nil)
	dicItem = append(dicItem, tempItem)

	v1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
	aq1 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v1}
	if emptyLayer == 1 {
		aq1.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq1, v0.Marshal())
		iterKey = iterKey[0:64]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		dicItem = append(dicItem, tempItem)
		iter.Release()
		authDB.Close()
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq1.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq1, v0.Marshal())
	dicItem = append(dicItem, tempItem)

	v2 := new(fr.Element).SetBytes([]byte(q.CondVal[2].(string)))
	aq2 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v2}
	if emptyLayer == 2 {
		aq2.Flags = []int{0, 0} // hkey<key<hnxt
		_, iterKey, _ = queryAuthDBWithPrefix(iter, aq2, append(v0.Marshal(), v1.Marshal()...))
		iterKey = iterKey[0:96]
		iter.Seek(iterKey)
		tempItem = stringToDicItem(string(iter.Value()))
		if tempItem.Value == nil {
			return SumRes1d{}, fmt.Errorf("queryAuthDB error: required item not exist")
		}
		dicItem = append(dicItem, tempItem)
		iter.Release()
		authDB.Close()
		res := SumRes1d{
			Sum:      sum,
			FR:       fR,
			SumProof: sumProof,
			DicItems: dicItem,
		}
		return res, nil
	}
	aq2.Flags = []int{2, 0} // hkey=key<hnxt
	tempItem, _, _ = queryAuthDBWithPrefix(iter, aq2, append(v0.Marshal(), v1.Marshal()...))
	dicItem = append(dicItem, tempItem)

	if q.CondFlag[3] == 0 {
		v3 := new(fr.Element).SetBytes([]byte(q.CondVal[3].(string)))
		aq3 := AuthQuery{AuthDB: ind.AuthTable(), Key: *v3}
		if len(rset) == 0 {
			aq3.Flags = []int{0, 0}
		} else {
			aq3.Flags = []int{2, 0}
		}
		// do not need to clip iterKey for the bottom layer
		prefix := append(v0.Marshal(), v1.Marshal()...)
		prefix = append(prefix, v2.Marshal()...)
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aq3, prefix)
		dicItem = append(dicItem, tempItem)
	} else {
		v3l := new(fr.Element).SetUint64(q.CondVal[3].([]uint64)[0])
		v3r := new(fr.Element).SetUint64(q.CondVal[3].([]uint64)[1])

		// fmt.Printf("test, v3l=%v, v3r=%v\n", v3l.String(), v3r.String())

		aqLeft := AuthQuery{AuthDB: ind.AuthTable(), Key: *v3l}
		aqRight := AuthQuery{AuthDB: ind.AuthTable(), Key: *v3r}
		switch q.CondFlag[3] {
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
		prefix := append(v0.Marshal(), v1.Marshal()...)
		prefix = append(prefix, v2.Marshal()...)
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqLeft, prefix)
		//fmt.Printf("test: item3l.key=%v, item3l.Nxt=%v\n", tempItem.Key.String(), tempItem.Nxt.String())
		dicItem = append(dicItem, tempItem)
		tempItem, _, _ = queryAuthDBWithPrefix(iter, aqRight, prefix)
		//fmt.Printf("test: item3r.key=%v, item3r.Nxt=%v\n", tempItem.Key.String(), tempItem.Nxt.String())
		dicItem = append(dicItem, tempItem)
	}

	iter.Release()
	authDB.Close()

	res := SumRes1d{
		Sum:      sum,
		FR:       fR,
		SumProof: sumProof,
		DicItems: dicItem,
	}
	return res, nil
}

func queryAuthDBWithPrefix(iter iterator.Iterator, aq AuthQuery, prefix []byte) (dic acc.DicItem, iterKey []byte, err error) {

	hkey := aq.Key.Marshal()
	seekkey := append(prefix, hkey...)
	var val []byte
	iter.Seek(seekkey)

	switch aq.Flags[0] {
	case 0: // <
		if !iter.Valid() || bytes.Compare(iter.Key(), seekkey) >= 0 {
			if !iter.Prev() {
				return dic, iter.Key(), fmt.Errorf("queryAuthDB error: %v", err)
			}
		}
		val = iter.Value()
	case 1: // <=
		if !iter.Valid() || bytes.Compare(iter.Key(), seekkey) > 0 {
			if !iter.Prev() {
				return dic, iter.Key(), fmt.Errorf("queryAuthDB error: %v", err)
			}
		}
		val = iter.Value()
	case 2: // =
		if !iter.Valid() {
			return dic, iter.Key(), fmt.Errorf("queryAuthDB error: %v", err)
		}
		val = iter.Value()
	default:
		return dic, iter.Key(), fmt.Errorf("queryAuthDB error: invalid flag")
	}
	// parse val to dicItem
	dic = stringToDicItem(string(val))
	if dic.Value == nil {
		return dic, iter.Key(), fmt.Errorf("queryAuthDB error: required item not exist")
	}

	switch aq.Flags[1] {
	case 0: // k<nxt
		if !(aq.Key.Cmp(&dic.Nxt) == -1) {
			return acc.DicItem{}, iter.Key(), fmt.Errorf("queryAuthDB error: required item not exist")
		}
	case 1: // k<=nxt
		if !(aq.Key.Cmp(&dic.Nxt) != 1) {
			return acc.DicItem{}, iter.Key(), fmt.Errorf("queryAuthDB error: required item not exist")
		}
	case 2: // k==nxt
		if !(aq.Key.Equal(&dic.Nxt)) {
			return acc.DicItem{}, iter.Key(), fmt.Errorf("queryAuthDB error: required item not exist")
		}
	default:
		return acc.DicItem{}, iter.Key(), fmt.Errorf("queryAuthDB error: required item not exist")
	}

	return dic, iter.Key(), nil
}

func QueryToMultiCondInd(db *sql.DB, q Query) (IndexInfo, error) {
	type FlagCond struct {
		Cond  string
		Flag  int
		Count int
	}
	var eqconds []FlagCond
	var rangecond []FlagCond
	for i := range q.CondFlag {
		if q.CondFlag[i] == 0 {
			eqconds = append(eqconds, FlagCond{Cond: q.Cond[i], Flag: q.CondFlag[i]})
		} else {
			rangecond = append(rangecond, FlagCond{Cond: q.Cond[i], Flag: q.CondFlag[i]})
		}
	}
	if len(rangecond) > 1 {
		return IndexInfo{}, fmt.Errorf("multi-cond index do not exist")
	}

	for i := range eqconds {
		eqQuery := fmt.Sprintf("SELECT COUNT(DISTINCT %v) FROM %v;", eqconds[i].Cond, q.Table)
		rows, err := db.Query(eqQuery)
		if err != nil {
			return IndexInfo{}, fmt.Errorf("error db.Query: %v", err)
		}
		rows.Next()
		var count int
		err = rows.Scan(&count)
		if err != nil {
			return IndexInfo{}, fmt.Errorf("error db.Query: %v", err)
		}
		eqconds[i].Count = count

		defer rows.Close()
	}

	sort.Slice(eqconds, func(i, j int) bool {
		return eqconds[i].Count < eqconds[j].Count
	})

	// reorder conditions in ind
	conds := []string{}
	for i := range eqconds {
		conds = append(conds, eqconds[i].Cond)
	}
	if len(rangecond) != 0 {
		conds = append(conds, rangecond[0].Cond)
	}

	ind := IndexInfo{
		Table:    q.Table,
		Cond:     conds,
		Dest:     q.Dest,
		DestType: q.DestType,
	}
	return ind, nil
}
