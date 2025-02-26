package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"path/filepath"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/syndtr/goleveldb/leveldb"
)

func Verify(
	dbinfo string,
	q Query,
	res interface{},
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {
	if len(q.Cond) == 0 {
		return false, fmt.Errorf("no index in query")
	}

	db, err := sql.Open("mysql", dbinfo)
	if err != nil {
		return false, err
	}
	defer db.Close()

	dres := true
	switch q.DestType {
	case "sum":
		dres, err = verifySum1d(db, q, res.(SumRes1d), pk1, pk2)
	case "max", "min":
		dres, err = verifyMaxmin1d(db, q, res.(MaxminRes1d), pk1, pk2)
	default:
		dres, err = false, fmt.Errorf("verify do not support type %v", q.DestType)
	}
	if err != nil || !dres {
		return false, err
	}

	return true, nil

}

func verifySum1d(
	db *sql.DB,
	q Query,
	res SumRes1d,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {

	//log.Printf("verify sum query ...")
	ver, err := acc.VerifySum(res.FR, res.Sum, res.SumProof, pk1[0], pk2[0], pk2[1])
	if !ver || err != nil {
		return false, fmt.Errorf("verify sum failed: %v", err)
	}

	it := 0 // the dest index d is set to 0 in 1d verify
	for c := range q.Cond {
		var ver, empty bool
		var err error
		switch q.CondFlag[c] {
		case 0:
			ver, empty, err = verifyMiddleEq(q, res.IProof.DigestSet[c], res.DicItems[it], c, it, pk1)
			it += 1
		case 1, 2, 3, 4:
			ver, empty, err = verifyMiddleRange(q, res.IProof.DigestSet[c], res.DicItems[it], res.DicItems[it+1], c, it, pk1)
			it += 2
		default:
			return false, fmt.Errorf("invalid flag")
		}
		if !ver || err != nil {
			return false, err
		}
		if empty { // meet an empty middle result, the final result must be empty
			if res.FR.Equal(&pk1[0]) {
				return true, nil
			} else {
				return false, fmt.Errorf("non-zero sum for empty result")
			}
		}
	}
	// log.Printf("verify intersection ...")
	if !acc.VerifyIntersection(res.FR, res.IProof, pk1) {
		return false, fmt.Errorf("verify intersection fail")
	}

	return true, nil
}

func verifyMiddleEq(
	//db *sql.DB,
	q Query,
	//res SumRes1d,
	mdi bn.G1Affine,
	item acc.DicItem,
	c, it int,
	pk1 []bn.G1Affine,
) (bool, bool, error) {
	// log.Printf("verify middle query, cond: %v, condVal: %v ...", q.Cond[c], q.CondVal[c])

	dicDigest, err := getDigest1c1d(q, c)
	if err != nil {
		return false, false, fmt.Errorf("error in verifyMiddleEq: %v", err)
	}

	// mdi := res.IProof.DigestSet[c] // middle result set digest, mdi
	// item := res.DicItems[it]

	if !acc.VerifyDic(dicDigest, item, pk1) {
		return false, false, fmt.Errorf("verify dic item failed")
	}

	v := new(fr.Element).SetBytes([]byte(q.CondVal[c].(string)))
	if mdi.Equal(&pk1[0]) { // empty
		if !(item.Key.Cmp(v) == -1 && v.Cmp(&item.Nxt) == -1) {
			return false, false, fmt.Errorf("verify dic item key failed")
		}
		// log.Printf("... meet empty middle result, verify OK")
		return true, true, nil
	}
	if !item.Key.Equal(v) {
		return false, false, fmt.Errorf("verify dic item key failed")
	}
	val := item.Value.(bn.G1Affine)
	if !(&val).Equal(&mdi) {
		return false, false, fmt.Errorf("verify item value and fR failed")
	}
	return true, false, nil
}

func verifyMiddleRange(
	//db *sql.DB,
	q Query,
	// res SumRes1d,
	mdi bn.G1Affine,
	item1 acc.DicItem,
	item2 acc.DicItem,
	c, it int,
	pk1 []bn.G1Affine,
) (bool, bool, error) {
	// log.Printf("verify middle query, cond: %v, condVal: %v ...", q.Cond[c], q.CondVal[c])

	dicDigest, err := getDigest1c1d(q, c)
	if err != nil {
		return false, false, fmt.Errorf("error in verifyMiddleRange: %v", err)
	}

	//mdi := res.IProof.DigestSet[c] // middle result set digest, mdi
	//item1, item2 := res.DicItems[it], res.DicItems[it+1]
	vl, vr := new(fr.Element).SetUint64(q.CondVal[c].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[c].([]uint64)[1])
	k1, k2, n1, n2, val1, val2 := item1.Key, item2.Key, item1.Nxt, item2.Nxt, item1.Value.(bn.G2Affine), item2.Value.(bn.G2Affine)

	if !acc.VerifyDic(dicDigest, item1, pk1) || !acc.VerifyDic(dicDigest, item2, pk1) {
		return false, false, fmt.Errorf("verify dic item failed")
	}

	switch q.CondFlag[c] {
	case 1: // [vl,vr]
		if mdi.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) == -1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) == -1) {
				return false, false, fmt.Errorf("verify dic item key failed")
			}
			// log.Printf("... meet empty middle result, verify OK")
			return true, true, nil
		}
		if !(k1.Cmp(vl) == -1 && vl.Cmp(&n1) != 1) || !(k2.Cmp(vr) != 1 && vr.Cmp(&n2) == -1) {
			return false, false, fmt.Errorf("verify dic item key failed")
		}
	case 2: // (vl,vr]
		if mdi.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) == -1) {
				return false, false, fmt.Errorf("verify dic item key failed")
			}
			// log.Printf("... meet empty middle result, verify OK")
			return true, true, nil
		}
		if !(k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1) || !(k2.Cmp(vr) != 1 && vr.Cmp(&n2) == -1) {
			return false, false, fmt.Errorf("verify dic item key failed")
		}
	case 3: // [vl,vr)
		if mdi.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) == -1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
				return false, false, fmt.Errorf("verify dic item key failed")
			}
			// log.Printf("... meet empty middle result, verify OK")
			return true, true, nil
		}
		if !(k1.Cmp(vl) == -1 && vl.Cmp(&n1) != 1) || !(k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
			return false, false, fmt.Errorf("verify dic item key failed")
		}
	case 4: // (vl,vr)
		if mdi.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
				return false, false, fmt.Errorf("verify dic item key failed")
			}
			// log.Printf("... meet empty middle result, verify OK")
			return true, true, nil
		}
		if !(k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1) || !(k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
			return false, false, fmt.Errorf("verify dic item key failed")
		}
	}

	p1 := []bn.G1Affine{mdi, pk1[0]}
	p2 := []bn.G2Affine{val1, val2}
	p1[0].Neg(&p1[0])
	e1, err := bn.PairingCheck(p1, p2)
	if err != nil || !e1 {
		return false, false, fmt.Errorf("verify range difference failed")
	}
	return true, false, nil
}

func getDigest1c1d(q Query, c int) (bn.G1Affine, error) {
	// auth table: table_index_dest, digest is in the first line
	ind := &IndexInfo{q.Table, q.Cond[c : c+1], q.Dest, q.DestType}
	authtable := ind.AuthTable()

	authDB := filepath.Join(acc.BaseDir, "authdb", authtable)
	db, err := leveldb.OpenFile(authDB, nil)
	if err != nil {
		return bn.G1Affine{}, fmt.Errorf("get digest error: %v", err)
	}
	inf := acc.INF.Bytes()
	val, err := db.Get(inf[:], nil)
	db.Close()
	if err != nil {
		return bn.G1Affine{}, fmt.Errorf("get digest error: %v", err)
	}
	return acc.StringToG1Affine(string(val)), nil
}

func getDigest(ind IndexInfo) (bn.G1Affine, error) {
	authtable := ind.AuthTable()

	authDB := filepath.Join(acc.BaseDir, "authdb", authtable)
	db, err := leveldb.OpenFile(authDB, nil)
	if err != nil {
		return bn.G1Affine{}, fmt.Errorf("get digest error: %v", err)
	}
	inf := acc.INF.Bytes()
	val, err := db.Get(inf[:], nil)
	db.Close()
	if err != nil {
		return bn.G1Affine{}, fmt.Errorf("get digest error: %v", err)
	}
	return acc.StringToG1Affine(string(val)), nil
}
