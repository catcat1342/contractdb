package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// func VerifyMaxmin(
// 	db *sql.DB,
// 	q Query,
// 	res []MaxminRes1d,
// 	pk1 []bn.G1Affine,
// 	pk2 []bn.G2Affine,
// ) (bool, error) {
// 	// ind := IndexInfo{q.Table, q.Cond, q.Dest}
// 	// if CheckTableExist(db, []string{ind.authTable()}) {
// 	// 	if len(q.Cond) == 1 && len(q.Dest) == 1 {
// 	// 		return verifyMaxmin1c1d(db, q, res[0], pk1, pk2)
// 	// 	} else {
// 	// 		return false, fmt.Errorf("not support yet")
// 	// 	}
// 	// }

// 	for c := range q.Cond {
// 		for d := range q.Dest {
// 			ind := &IndexInfo{q.Table, q.Cond[c : c+1], q.Dest[d : d+1]}
// 			if !CheckTableExist(db, []string{ind.authTable()}) {
// 				return false, fmt.Errorf("lack index")
// 			}
// 		}
// 	}
// 	return verifyMaxmin1d(db, q, res[0], pk1, pk2)
// }

func verifyMaxmin1d(
	db *sql.DB,
	q Query,
	res MaxminRes1d,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {

	log.Printf("Verify max min query ...")

	newQ := MaxminToRange(q, fr.Vector{res.Result})

	it := 0 // indicate item
	for c := range newQ.Cond {
		var ver, empty bool
		var err error
		switch newQ.CondFlag[c] {
		case 0:
			ver, empty, err = verifyMiddleEq(newQ, res.IProof.DigestSet[c], res.DicItems[it], c, it, pk1)
			it += 1
		case 1, 2, 3, 4:
			ver, empty, err = verifyMiddleRange(newQ, res.IProof.DigestSet[c], res.DicItems[it], res.DicItems[it+1], c, it, pk1)
			it += 2
		default:
			return false, fmt.Errorf("invalid flag")
		}
		if !ver || err != nil {
			return false, err
		}
		if empty { // result should not be empty
			return false, fmt.Errorf("non-zero sum for empty result")
		}
	}

	fr := acc.ComputeAccG1(fr.Vector{res.Result}, pk1)
	if res.Result.IsZero() {
		fr.Set(&pk1[0])
	}
	if !acc.VerifyIntersection(fr, res.IProof, pk1) {
		return false, fmt.Errorf("verify intersection fail")
	}

	return true, nil
}

// // verify a single dest
// func verifyMaxmin1c1d(
// 	db *sql.DB,
// 	q Query,
// 	res MaxminRes1d,
// 	pk1 []bn.G1Affine,
// 	pk2 []bn.G2Affine,
// ) (bool, error) {
// 	log.Printf("Verify max min query ...")
// 	/* check sum, check dic key, check dic item */

// 	dicDigest, err := getDigest1c1d(db, q, 0, 0)
// 	if err != nil {
// 		return false, fmt.Errorf("error in getSumDigest: %v", err)
// 	}

// 	switch q.CondFlag[0] {
// 	case 0:
// 		item := res.DicItems[0]
// 		if !acc.VerifyDic(dicDigest, item, pk1) {
// 			return false, fmt.Errorf("verify dic item failed")
// 		}
// 		v := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
// 		if res.FR.Equal(&pk1[0]) { // empty
// 			if !(item.Key.Cmp(v) == -1 && v.Cmp(&item.Nxt) == -1) {
// 				return false, fmt.Errorf("verify dic item key failed")
// 			}
// 			log.Printf("... meet empty middle result, OK")
// 			return true, nil
// 		}
// 		if !item.Key.Equal(v) {
// 			return false, fmt.Errorf("verify dic item key failed")
// 		}
// 		val := item.Value.(bn.G1Affine)
// 		if !(&val).Equal(&res.FR) {
// 			return false, fmt.Errorf("verify item value and fR failed")
// 		}
// 		log.Printf("... verify dic ok\nOK\n")
// 		return true, nil
// 	case 1, 2, 3, 4:
// 		item1, item2 := res.DicItems[0], res.DicItems[1]
// 		k1, k2, n1, n2, val1, val2 := item1.Key, item2.Key, item1.Nxt, item2.Nxt, item1.Value.(bn.G2Affine), item2.Value.(bn.G2Affine)
// 		vl, vr := new(fr.Element).SetUint64(q.CondVal[0].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[0].([]uint64)[1])

// 		if !acc.VerifyDic(dicDigest, item1, pk1) || !acc.VerifyDic(dicDigest, item2, pk1) {
// 			return false, fmt.Errorf("verify dic item failed")
// 		}
// 		if res.FR.Equal(&pk1[0]) { // empty
// 			if !(k1.Equal(&item2.Key) && k1.Cmp(vl) == -1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) == -1) {
// 				return false, fmt.Errorf("verify dic item key failed")
// 			}
// 			log.Printf("... meet empty middle result, OK")
// 			return true, nil
// 		}
// 		if !(k1.Cmp(vl) == -1 && vl.Cmp(&n1) != 1) || !(k2.Cmp(vr) != 1 && vr.Cmp(&n2) == -1) {
// 			return false, fmt.Errorf("verify dic item key failed")
// 		}
// 		log.Printf("... verify dic keys ok\n")
// 		// e1 := bn.Pairing(item1.Value.(*bn.PointP), res.FR)
// 		// e2 := bn.Pairing(item2.Value.(*bn.PointP), pk1[0])
// 		p1 := []bn.G1Affine{res.FR, pk1[0]}
// 		p2 := []bn.G2Affine{val1, val2}
// 		p1[0].Neg(&p1[0])
// 		e1, err := bn.PairingCheck(p1, p2)
// 		if err != nil || !e1 {
// 			return false, fmt.Errorf("verify range difference failed")
// 		}
// 		log.Printf("... verify range difference ok\n")
// 		return true, nil
// 	default:
// 		return false, fmt.Errorf("invalid flag")
// 	}
// }
