package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type MaxminRes1d struct {
	Result   fr.Element // the max or min value
	DicItems []acc.DicItem
	IProof   acc.IntersectionProof
}

func queryMaxmin1d(db *sql.DB, q Query, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (MaxminRes1d, error) {
	rset, err := querySet1d(db, q, len(q.Cond))
	if err != nil {
		return MaxminRes1d{}, err
	}
	var dicItem []acc.DicItem
	var middleSet []fr.Vector
	// fmt.Printf("result: %v\n", rset[0].String())

	newQ := MaxminToRange(q, rset)

	// for each cond, generate dic item and wit
	for c := range newQ.Cond {
		switch newQ.CondFlag[c] {
		case 0:
			mset, item, err := query1c1dEq(db, newQ, c)
			if err != nil {
				return MaxminRes1d{}, err
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, item)
		case 1, 2, 3, 4:
			mset, items, err := query1c1dRange(db, newQ, c)
			if err != nil {
				return MaxminRes1d{}, err
			}
			middleSet = append(middleSet, mset)
			dicItem = append(dicItem, items...)
		default:
			return MaxminRes1d{}, fmt.Errorf("invalid flag")
		}
	}

	iproof := acc.ProveIntersection(middleSet, pk1, pk2)

	res := MaxminRes1d{
		DicItems: dicItem,
		IProof:   iproof,
	}
	if len(rset) == 0 {
		res.Result = *new(fr.Element).SetZero()
	} else {
		res.Result = rset[0]
	}

	return res, nil
}

// convert a 1-d manxmin query to range query
func MaxminToRange(q Query, result fr.Vector) Query {

	if len(result) == 0 || result[0].IsZero() {
		return Query{
			Table:    q.Table,
			Cond:     q.Cond,
			CondVal:  q.CondVal,
			CondFlag: q.CondFlag,
			Dest:     q.Dest,
			DestType: "sum",
		}
	}
	newQ := Query{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: "sum",
	}
	newQ.Cond = make([]string, len(q.Cond))
	copy(newQ.Cond, q.Cond)
	newQ.CondVal = make([]interface{}, len(q.CondVal))
	copy(newQ.CondVal, q.CondVal)
	newQ.CondFlag = make([]int, len(q.CondFlag))
	copy(newQ.CondFlag, q.CondFlag)
	// add dest to cond
	newQ.Cond = append(newQ.Cond, q.Dest)
	switch q.DestType {
	case "max": // >= rset[0]  [rset[0],INF) -> flag=3
		newQ.CondVal = append(newQ.CondVal, []uint64{result[0].Uint64(), acc.INF.Uint64()})
		newQ.CondFlag = append(newQ.CondFlag, 3)
	case "min": // >= rset[0]  (MIN, rset[0]] -> flag=2
		newQ.CondVal = append(newQ.CondVal, []uint64{acc.MIN.Uint64(), result[0].Uint64()})
		newQ.CondFlag = append(newQ.CondFlag, 2)
	default:
		return Query{}
	}
	return newQ
}
