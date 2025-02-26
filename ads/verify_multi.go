package ads

import (
	acc "contractdb/accumulator"
	"fmt"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func VerifyMulti(
	dbinfo string,
	q Query,
	res interface{},
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {
	if len(q.Cond) == 0 {
		return false, fmt.Errorf("no index in query")
	}
	var err error

	// db, err := sql.Open("mysql", dbinfo)
	// if err != nil {
	// 	return false, err
	// }
	// defer db.Close()

	dres := true
	switch q.DestType {
	case "sum":
		switch len(q.Cond) {
		case 2:
			dres, err = verifySum2c1d(q, res.(SumRes1d), pk1, pk2)
		case 3:
			dres, err = verifySum3c1d(q, res.(SumRes1d), pk1, pk2)
		case 4:
			dres, err = verifySum4c1d(q, res.(SumRes1d), pk1, pk2)
		default:
			return false, fmt.Errorf("verify do not support")
		}

	default:
		dres, err = false, fmt.Errorf("verify do not support type %v", q.DestType)
	}
	if err != nil || !dres {
		return false, err
	}

	return true, nil

}

func verifySum2c1d(
	q Query,
	res SumRes1d,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {

	//log.Printf("Verify sum query ...")

	ver, err := acc.VerifySum(res.FR, res.Sum, res.SumProof, pk1[0], pk2[0], pk2[1])
	if !ver || err != nil {
		return false, fmt.Errorf("verify sum failed: %v", err)
	}
	//log.Printf("... ok\n")

	// verify dic items
	//log.Printf("Verify dic items ...")

	// // get digest0 from authDB
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return false, err
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}
	digest0, err := getDigest(ind)
	if err != nil {
		return false, err
	}
	item0 := res.DicItems[0]
	// check whether item exists in authDB
	if !acc.VerifyDic(digest0, item0, pk1) { // e(item.ToElement(), digest)==e(item.W, G2)
		return false, fmt.Errorf("verify layer0 dic item failed")
	}
	// check whether item key and nxt satisfies query conditions
	condVal0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	if len(res.DicItems) == 1 { // emptyLayer==0
		if !(item0.Key.Cmp(condVal0) == -1 && condVal0.Cmp(&item0.Nxt) == -1) {
			return false, fmt.Errorf("verify layer0 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-0\n")
		return true, nil
	} else {
		if !item0.Key.Equal(condVal0) {
			return false, fmt.Errorf("verify layer0 item key failed")
		}
	}

	digest1 := item0.Value.(bn.G1Affine)

	if q.CondFlag[1] == 0 {
		item1 := res.DicItems[1]
		condVal1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
		if res.FR.Equal(&pk1[0]) { // empty result
			if !(item1.Key.Cmp(condVal1) == -1 && condVal1.Cmp(&item1.Nxt) == -1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			//log.Printf("... ok, empty result in layer-2\n")
			return true, nil
		}
		if !item1.Key.Equal(condVal1) {
			return false, fmt.Errorf("verify layer1 item key failed")
		}
		val := item1.Value.(bn.G1Affine)
		if !(&val).Equal(&res.FR) {
			return false, fmt.Errorf("verify item value and fR failed")
		}

	} else {
		item1l := res.DicItems[1]
		item1r := res.DicItems[2]
		if !acc.VerifyDic(digest1, item1l, pk1) || !acc.VerifyDic(digest1, item1r, pk1) {
			return false, fmt.Errorf("verify layer1 dic item failed")
		}
		vl, vr := new(fr.Element).SetUint64(q.CondVal[1].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[1].([]uint64)[1])

		_, err := verifyBottomRange(1, vl, vr, res.FR, item1l, item1r, pk1)
		if err != nil {
			return false, err
		}
	}
	//log.Printf("... ok\n")
	return true, err
}

func verifySum3c1d(
	q Query,
	res SumRes1d,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {

	//log.Printf("Verify sum query ...")

	ver, err := acc.VerifySum(res.FR, res.Sum, res.SumProof, pk1[0], pk2[0], pk2[1])
	if !ver || err != nil {
		return false, fmt.Errorf("verify sum failed: %v", err)
	}
	//log.Printf("... ok\n")

	// verify dic items
	//log.Printf("Verify dic items ...")

	// // get digest0 from authDB
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return false, err
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}
	digest0, err := getDigest(ind)
	if err != nil {
		return false, err
	}
	item0 := res.DicItems[0]
	// check whether item exists in authDB
	if !acc.VerifyDic(digest0, item0, pk1) { // e(item.ToElement(), digest)==e(item.W, G2)
		return false, fmt.Errorf("verify layer0 dic item failed")
	}
	// check whether item key and nxt satisfies query conditions
	condVal0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	if len(res.DicItems) == 1 { // emptyLayer==0
		if !(item0.Key.Cmp(condVal0) == -1 && condVal0.Cmp(&item0.Nxt) == -1) {
			return false, fmt.Errorf("verify layer0 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-0\n")
		return true, nil
	} else {
		if !item0.Key.Equal(condVal0) {
			return false, fmt.Errorf("verify layer0 item key failed")
		}
	}

	digest1 := item0.Value.(bn.G1Affine)
	item1 := res.DicItems[1]
	if !acc.VerifyDic(digest1, item1, pk1) {
		return false, fmt.Errorf("verify layer1 dic item failed")
	}
	condVal1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
	if len(res.DicItems) == 2 { // emptyLayer==1
		if !(item1.Key.Cmp(condVal1) == -1 && condVal1.Cmp(&item1.Nxt) == -1) {
			return false, fmt.Errorf("verify layer1 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-1\n")
		return true, nil
	} else {
		if !item1.Key.Equal(condVal1) {
			return false, fmt.Errorf("verify layer1 item key failed")
		}
	}

	digest2 := item1.Value.(bn.G1Affine)

	if q.CondFlag[2] == 0 {
		item2 := res.DicItems[2]
		condVal2 := new(fr.Element).SetBytes([]byte(q.CondVal[2].(string)))
		if res.FR.Equal(&pk1[0]) { // empty result
			if !(item2.Key.Cmp(condVal2) == -1 && condVal2.Cmp(&item2.Nxt) == -1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			//log.Printf("... ok, empty result in layer-2\n")
			return true, nil
		}
		if !item2.Key.Equal(condVal2) {
			return false, fmt.Errorf("verify layer1 item key failed")
		}
		val := item2.Value.(bn.G1Affine)
		if !(&val).Equal(&res.FR) {
			return false, fmt.Errorf("verify item value and fR failed")
		}

	} else {
		item2l := res.DicItems[2]
		item2r := res.DicItems[3]
		if !acc.VerifyDic(digest2, item2l, pk1) || !acc.VerifyDic(digest2, item2r, pk1) {
			return false, fmt.Errorf("verify layer1 dic item failed")
		}
		vl, vr := new(fr.Element).SetUint64(q.CondVal[2].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[2].([]uint64)[1])

		_, err := verifyBottomRange(1, vl, vr, res.FR, item2l, item2r, pk1)
		if err != nil {
			return false, err
		}
	}
	//log.Printf("... ok\n")
	return true, err
}

func verifySum4c1d(
	q Query,
	res SumRes1d,
	pk1 []bn.G1Affine,
	pk2 []bn.G2Affine,
) (bool, error) {

	//log.Printf("Verify sum query ...")

	ver, err := acc.VerifySum(res.FR, res.Sum, res.SumProof, pk1[0], pk2[0], pk2[1])
	if !ver || err != nil {
		return false, fmt.Errorf("verify sum failed: %v", err)
	}
	//log.Printf("... ok\n")

	// verify dic items
	//log.Printf("Verify dic items ...")

	// // get digest0 from authDB
	// ind, err := QueryToMultiCondInd(db, q)
	// if err != nil {
	// 	return false, err
	// }
	ind := IndexInfo{
		Table:    q.Table,
		Dest:     q.Dest,
		DestType: q.DestType,
		Cond:     q.Cond,
	}
	digest0, err := getDigest(ind)
	if err != nil {
		return false, err
	}
	item0 := res.DicItems[0]
	// check whether item exists in authDB
	if !acc.VerifyDic(digest0, item0, pk1) { // e(item.ToElement(), digest)==e(item.W, G2)
		return false, fmt.Errorf("verify layer0 dic item failed")
	}
	// check whether item key and nxt satisfies query conditions
	condVal0 := new(fr.Element).SetBytes([]byte(q.CondVal[0].(string)))
	if len(res.DicItems) == 1 { // emptyLayer==0
		if !(item0.Key.Cmp(condVal0) == -1 && condVal0.Cmp(&item0.Nxt) == -1) {
			return false, fmt.Errorf("verify layer0 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-0\n")
		return true, nil
	} else {
		if !item0.Key.Equal(condVal0) {
			return false, fmt.Errorf("verify layer0 item key failed")
		}
	}

	digest1 := item0.Value.(bn.G1Affine)
	item1 := res.DicItems[1]
	if !acc.VerifyDic(digest1, item1, pk1) {
		return false, fmt.Errorf("verify layer1 dic item failed")
	}
	condVal1 := new(fr.Element).SetBytes([]byte(q.CondVal[1].(string)))
	if len(res.DicItems) == 2 { // emptyLayer==1
		if !(item1.Key.Cmp(condVal1) == -1 && condVal1.Cmp(&item1.Nxt) == -1) {
			return false, fmt.Errorf("verify layer1 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-1\n")
		return true, nil
	} else {
		if !item1.Key.Equal(condVal1) {
			return false, fmt.Errorf("verify layer1 item key failed")
		}
	}

	digest2 := item1.Value.(bn.G1Affine)
	item2 := res.DicItems[2]
	if !acc.VerifyDic(digest2, item2, pk1) {
		return false, fmt.Errorf("verify layer2 dic item failed")
	}
	condVal2 := new(fr.Element).SetBytes([]byte(q.CondVal[2].(string)))
	if len(res.DicItems) == 3 { // emptyLayer==1
		if !(item2.Key.Cmp(condVal2) == -1 && condVal2.Cmp(&item2.Nxt) == -1) {
			return false, fmt.Errorf("verify layer2 dic item key failed")
		}
		//log.Printf("... ok, empty result in layer-2\n")
		return true, nil
	} else {
		if !item2.Key.Equal(condVal2) {
			return false, fmt.Errorf("verify layer2 item key failed")
		}
	}

	digest3 := item2.Value.(bn.G1Affine)
	if q.CondFlag[3] == 0 {
		item3 := res.DicItems[3]
		condVal3 := new(fr.Element).SetBytes([]byte(q.CondVal[3].(string)))
		if res.FR.Equal(&pk1[0]) { // empty result
			if !(item3.Key.Cmp(condVal3) == -1 && condVal3.Cmp(&item3.Nxt) == -1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			//log.Printf("... ok, empty result in layer-3\n")
			return true, nil
		}
		if !item3.Key.Equal(condVal3) {
			return false, fmt.Errorf("verify layer3 item key failed")
		}
		val := item3.Value.(bn.G1Affine)
		if !(&val).Equal(&res.FR) {
			return false, fmt.Errorf("verify layer3 item value and fR failed")
		}

	} else {
		item3l := res.DicItems[3]
		item3r := res.DicItems[4]
		if !acc.VerifyDic(digest3, item3l, pk1) || !acc.VerifyDic(digest3, item3r, pk1) {
			return false, fmt.Errorf("verify layer3 dic item failed")
		}
		vl, vr := new(fr.Element).SetUint64(q.CondVal[3].([]uint64)[0]), new(fr.Element).SetUint64(q.CondVal[3].([]uint64)[1])

		ver, err := verifyBottomRange(1, vl, vr, res.FR, item3l, item3r, pk1)
		if err != nil || !ver {
			return false, fmt.Errorf("verify difference failed")
		}
	}
	//log.Printf("... ok\n")
	return true, err
}

func verifyBottomRange(
	bottomFlag int,
	vl, vr *fr.Element,
	resAcc bn.G1Affine,
	item1, item2 acc.DicItem,
	pk1 []bn.G1Affine,
) (bool, error) {
	k1, k2, n1, n2, val1, val2 := item1.Key, item2.Key, item1.Nxt, item2.Nxt, item1.Value.(bn.G2Affine), item2.Value.(bn.G2Affine)

	switch bottomFlag {
	case 1: // [vl,vr]
		if resAcc.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) == -1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) == -1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			return true, nil
		}
		if !(k1.Cmp(vl) == -1 && vl.Cmp(&n1) != 1) || !(k2.Cmp(vr) != 1 && vr.Cmp(&n2) == -1) {
			return false, fmt.Errorf("verify dic item key failed")
		}
	case 2: // (vl,vr]
		if resAcc.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) == -1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			return true, nil
		}
		if !(k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1) || !(k2.Cmp(vr) != 1 && vr.Cmp(&n2) == -1) {
			return false, fmt.Errorf("verify dic item key failed")
		}
	case 3: // [vl,vr)
		if resAcc.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) == -1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			return true, nil
		}
		if !(k1.Cmp(vl) == -1 && vl.Cmp(&n1) != 1) || !(k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
			return false, fmt.Errorf("verify dic item key failed")
		}
	case 4: // (vl,vr)
		if resAcc.Equal(&pk1[0]) { // empty
			if !(k1.Equal(&k2) && k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1 && k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
				return false, fmt.Errorf("verify dic item key failed")
			}
			return true, nil
		}
		if !(k1.Cmp(vl) != 1 && vl.Cmp(&n1) == -1) || !(k2.Cmp(vr) == -1 && vr.Cmp(&n2) != 1) {
			return false, fmt.Errorf("verify dic item key failed")
		}
	}

	p1 := []bn.G1Affine{resAcc, pk1[0]}
	p2 := []bn.G2Affine{val1, val2}
	p1[0].Neg(&p1[0])
	e1, err := bn.PairingCheck(p1, p2)
	if err != nil || !e1 {
		return false, fmt.Errorf("verify range difference failed")
	}
	return true, nil
}
