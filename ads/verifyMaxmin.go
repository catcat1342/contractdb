package ads

import (
	acc "contractdb/accumulator"
	"database/sql"
	"fmt"
	"log"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

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
