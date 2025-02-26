package ads

/* QueryDBMaxmin is the query interface of ContractDB database
 * Input: a dbinfo, the query condition, the public keys pk1 and pk2
 * Output: query results, related dic items and their witnesses, intersection proofs, error
 *
 * select sum(dest1),sum(dest2) from table T where cond1 and cond2 ...
 * Use cases: result, dicItem, dicWit, iProof, err := QueryDB(dbinfo, q, pk1, pk2)
 */

// func queryCount1d(
// 	db *sql.DB,
// 	q Query,
// 	pk1 []bn.G1Affine,
// 	pk2 []bn.G2Affine,
// ) ([]SumRes1d, error) {

// 	rset, err := querySet1d(db, q, len(q.Cond), 0)
// 	if err != nil {
// 		return SumRes1d{}, err
// 	}
// 	sum, fR, sumProof := acc.ProveSum(rset, pk1, pk2)
// 	var dicItem []acc.DicItem
// 	var middleSet []fr.Vector

// 	// for each cond, generate dic item and wit
// 	for c := range q.Cond {
// 		switch q.CondFlag[c] {
// 		case 0:
// 			mset, item, err := query1c1dEq(db, q, c, 0)
// 			if err != nil {
// 				return SumRes1d{}, err
// 			}
// 			middleSet = append(middleSet, mset)
// 			dicItem = append(dicItem, item)
// 		case 1, 2, 3, 4:
// 			mset, items, err := query1c1dRange(db, q, c, 0)
// 			if err != nil {
// 				return SumRes1d{}, err
// 			}
// 			middleSet = append(middleSet, mset)
// 			dicItem = append(dicItem, items...)
// 		default:
// 			return SumRes1d{}, fmt.Errorf("invalid flag")
// 		}
// 	}
// }

// func QueryDBAvg(
// 	db *sql.DB,
// 	q Query,
// 	pk1 []bn.G1Affine,
// 	pk2 []bn.G2Affine,
// ) ([]SumRes1d, error) {

// 	// 1. check whether multi_index exist
// 	ind := &IndexInfo{q.Table, q.Cond, q.Dest}
// 	if CheckTableExist(db, []string{ind.authTable()}) {
// 		if len(q.Cond) == 1 && len(q.Dest) == 1 {
// 			return querySum1c1d(db, q, pk1, pk2)
// 		}
// 		return nil, fmt.Errorf("optimized index combination not supported yet")
// 	}

// 	// 2. if index not exist, check whether split index exist
// 	for i := range q.Cond {
// 		for j := range q.Dest {
// 			ind := &IndexInfo{q.Table, q.Cond[i : i+1], q.Dest[j : j+1]}
// 			if !CheckTableExist(db, []string{ind.authTable()}) {
// 				return nil, fmt.Errorf("both index and split index not exist")
// 			}
// 		}
// 	}
// 	return querySum1d(db, q, pk1, pk2)
// }
