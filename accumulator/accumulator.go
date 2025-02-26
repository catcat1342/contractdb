package accumulator

import (
	"fmt"
	"log"
	"math/big"

	ecc "github.com/consensys/gnark-crypto/ecc"
	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	_ "github.com/go-sql-driver/mysql"
)

const ROUTINES = 8

type IntersectionProof struct {
	DigestSet []bn.G1Affine // digests of all sets f1=Acc(S1), ..., fk=Acc(Sk)
	Witness   []bn.G2Affine // digest pk2 of all differences fd1=Acc(D1)=Acc(S1-I), ...
	DigestGCD []bn.G1Affine // digests of Bezout coefficients fq1=q1,..., Sum(qi*di)=1, di is the Poly of Di
	// DigestI   bn.G1Affine   // digest of I
	/*
	 *	prove 1: e(wi, fI)=e(bn.G2, fi) ==> subset
	 *	prove 2: Mul(e(wi, fqi))=e(bn.G2,bn.G1) ==> Bezout Lemma, GCD of all difference sets should be 1
	 */
}

type SumProof struct {
	A0 fr.Element
	A1 fr.Element
	W1 bn.G1Affine
	W2 bn.G1Affine
}

func NewSetFromInt64(array []int64) fr.Vector {
	var set fr.Vector = make(fr.Vector, len(array))
	for i, v := range array {
		var nfr fr.Element
		nfr.SetInt64(v)
		set[i] = nfr
	}
	return set
}

func ComputeAccG1(set fr.Vector, pk1 []bn.G1Affine) bn.G1Affine {
	var p *Poly = SetToPoly(set)
	var acc bn.G1Affine
	acc.MultiExp(pk1[:p.Len()], *p, ecc.MultiExpConfig{NbTasks: ROUTINES})
	return acc
}

func ComputeAccG2(set fr.Vector, pk2 []bn.G2Affine) bn.G2Affine {
	var p *Poly = SetToPoly(set)
	var acc bn.G2Affine
	acc.MultiExp(pk2[:p.Len()], *p, ecc.MultiExpConfig{NbTasks: ROUTINES})
	return acc
}

func ComputePolyG1(p *Poly, pk1 []bn.G1Affine) bn.G1Affine {
	var acc bn.G1Affine
	acc.MultiExp(pk1[:p.Len()], *p, ecc.MultiExpConfig{NbTasks: ROUTINES})
	return acc
}

func ComputePolyG2(p *Poly, pk2 []bn.G2Affine) bn.G2Affine {
	var acc bn.G2Affine
	acc.MultiExp(pk2[:p.Len()], *p, ecc.MultiExpConfig{NbTasks: ROUTINES})
	return acc
}

func SetToPoly(set fr.Vector) *Poly {
	// the poly of an empty set is a 0-deg poly: 1
	// gauranteeing that the gcd of set S! and an empty set S2 is 1, rather than Poly(S)
	// because GCD(A,0)=A
	poly := new(Poly)
	if len(set) == 0 {
		return poly.SetOne()
	}
	poly.SetLen(len(set) + 1)          // P(x) = \sum (poly[i] \cdot x^i), degree=len(set)
	roots := make(fr.Vector, len(set)) // roots for P(x)

	for i, v := range set {
		roots[i].Neg(new(fr.Element).Inverse(&v)) // C(s)=\Mul(x_i^{-1}+s)
	}
	poly.BuildFromRoots(roots)
	return poly
}

func ProveSubset(A, I fr.Vector, pk2 []bn.G2Affine) bn.G2Affine {
	D := Difference(A, I)
	return ComputeAccG2(D, pk2)
}

func VerifySubset(fA, fI bn.G1Affine, wit bn.G2Affine) bool {
	// e(fA,G2)==e(fI,wit)
	_, _, _, G2 := bn.Generators()

	p1 := []bn.G1Affine{fA, fI}
	p2 := []bn.G2Affine{G2, wit}
	p1[0].Neg(&p1[0])

	e1, err := bn.PairingCheck(p1, p2)
	if err != nil {
		log.Panicf("pairing check error: %v\n", err)
		return false
	}
	return e1
}

// return A-I
func Difference(A, I fr.Vector) fr.Vector {
	countMap := make(map[string]int)
	var result fr.Vector
	for _, v := range I {
		countMap[v.String()] += 1
	}
	for _, v := range A {
		if countMap[v.String()] > 0 {
			countMap[v.String()] -= 1
		} else {
			result = append(result, v)
		}
	}
	return result
}

func Intersection(sets []fr.Vector) fr.Vector {

	// for i := range sets {
	// 	fmt.Printf("sets[%v]:", i)
	// 	for j := range sets[i] {
	// 		fmt.Printf(" %v", sets[i][j].String())
	// 	}
	// 	fmt.Printf("\n")
	// }

	n := len(sets)
	if n == 0 {
		return make(fr.Vector, 0)
	}
	if n == 1 {
		return sets[0]
	}
	I := make(fr.Vector, len(sets[0]))
	copy(I, sets[0])
	for i := 1; i < n; i++ {
		// I = Intersection2(I, sets[i])
		if I.Len() == 0 || sets[i].Len() == 0 {
			return make(fr.Vector, 0)
		}
		countMap := make(map[string]int)
		for _, v := range I {
			countMap[v.String()] += 1
		}
		var newI fr.Vector = make(fr.Vector, 0)
		for _, v := range sets[i] {
			if countMap[v.String()] > 0 {
				newI = append(newI, v)
				countMap[v.String()] -= 1
			}
		}
		I = newI
	}
	return I
}

func ProveIntersection(sets []fr.Vector, pk1 []bn.G1Affine, pk2 []bn.G2Affine) IntersectionProof {
	if len(sets) == 0 {
		return IntersectionProof{
			DigestSet: []bn.G1Affine{pk1[0]},
			Witness:   nil,
			DigestGCD: nil,
		}
	}

	if len(sets) == 1 {
		return IntersectionProof{
			DigestSet: []bn.G1Affine{ComputeAccG1(sets[0], pk1)},
			Witness:   nil,
			DigestGCD: nil,
		}
	}

	I := Intersection(sets)
	n := len(sets)
	//fI := ComputeAccG1(I, pk1)

	f := make([]bn.G1Affine, n) // digest of Si
	w := make([]bn.G2Affine, n) // witness of subsets

	D := make([]fr.Vector, n)    // difference sets
	d := make([]Poly, n)         // poly of differece sets
	q := make([]Poly, n)         // bezout coefficient for difference sets
	fq := make([]bn.G1Affine, n) // digest of qi

	for i := 0; i < n; i++ {
		f[i] = ComputeAccG1(sets[i], pk1)
		D[i] = Difference(sets[i], I)
		d[i] = *SetToPoly(D[i])
		w[i] = ComputePolyG2(&d[i], pk2)
		q[i].SetOne()
	}

	if len(sets) == 1 {
		// if only one set in sets, we just return its digest
		return IntersectionProof{
			DigestSet: f,
			Witness:   nil,
			DigestGCD: nil,
		}
	}

	// now, GCD of all di[i] is 1, and we should find qi[i] s.t. Sum(qi[i]*di[i])=1
	// we can use XGCD each for two items and update qi[i] iteratively
	dt, q0, q1, _ := XGCD(&d[0], &d[1])
	q[0].Set(q0)
	q[1].Set(q1)
	//fmt.Printf("dt=%v\n", dt.PrintString())
	//fmt.Printf("q[0]d[0]+q[1]d[1]=%v\n", q[0].Mul(d[0]).Add(q[1].Mul(d[1])).PrintString())

	for i := 2; i < n; i++ {
		dt1, t1, qi, _ := XGCD(dt, &d[i])
		q[i].Set(qi)
		//fmt.Printf("dt1=%v\n", dt1.PrintString())
		//fmt.Printf("t1*dt+q[%v]d[%v]=%v\n", i, i, t1.Mul(dt).Add(q[i].Mul(d[i])).PrintString())
		for j := 0; j < i; j++ {
			q[j].Mul(&q[j], t1)
		}
		dt = dt1
	}
	// update fq[i]
	for i := 0; i < n; i++ {
		fq[i] = ComputePolyG1(&q[i], pk1)
	}

	/*//check q: sum(q[i]d[i])=1
	for i := 0; i < n; i++ {
		fmt.Printf("q[%v]: %v\nd[%v]: %v\n", i, q[i].PrintString(), i, d[i].PrintString())
	}
	sumres := new(Poly).Mul(&q[0], &d[0])
	for i := 1; i < n; i++ {
		sumres.Add(sumres, new(Poly).Mul(&q[i], &d[i]))
	}
	fmt.Printf("sum q[i]d[i]: %v\n", sumres.PrintString())
	*/

	/* // test pairing
	fmt.Printf("w0: %v\nw1: %v\nw2: %v\n", w[0], w[1], w[2])
	fmt.Printf("fq0: %v\nfq1: %v\nfq2: %v\n", fq[0], fq[1], fq[2])

	testpair := bn.Pairing(w[0], fq[0])
	for i := 1; i < n; i++ {
		testpair = testpair.Mul(bn.Pairing(w[i], fq[i]))
	}
	fmt.Printf("testpair: %v\n", testpair)
	fmt.Printf("e(G2,G1): %v\n", bn.Pairing(pk2[0], pk1[0]))
	*/

	return IntersectionProof{
		DigestSet: f,
		Witness:   w,
		DigestGCD: fq,
	}
}

func VerifyIntersection(fI bn.G1Affine, proof IntersectionProof, pk1 []bn.G1Affine) bool {

	f, w, q := proof.DigestSet, proof.Witness, proof.DigestGCD

	n := len(f)
	if n == 0 || n == 1 {
		return fI.Equal(&f[0])
	}
	_, _, _, G2 := bn.Generators()
	for i := 0; i < n; i++ {
		var p1 []bn.G1Affine = make([]bn.G1Affine, 2)
		var p2 []bn.G2Affine = make([]bn.G2Affine, 2)
		p1[0].Set(&fI)
		p1[1].Set(&f[i])
		p2[0].Set(&w[i])
		p2[1].Set(&G2)
		p1[0].Neg(&p1[0])
		e1, err := bn.PairingCheck(p1, p2)
		if err != nil || !e1 {
			return false
		}
	}

	var p1 []bn.G1Affine = make([]bn.G1Affine, n+1)
	var p2 []bn.G2Affine = make([]bn.G2Affine, n+1)
	p1[0].Set(&pk1[0])
	p2[0].Set(&G2)
	for i := 1; i <= n; i++ {
		p1[i].Set(&q[i-1])
		p2[i].Set(&w[i-1])
	}
	p1[0].Neg(&p1[0])
	e1, err := bn.PairingCheck(p1, p2)
	if err != nil || !e1 {
		return false
	}
	return true
}

func ProveSum(S fr.Vector, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (fr.Element, bn.G1Affine, SumProof) {

	sum := new(fr.Element).SetZero()
	var acc bn.G1Affine = pk1[0]
	var proof SumProof

	if len(S) == 0 {
		return *sum, acc, proof
	}

	for i := range S {
		sum.Add(sum, &S[i])
	}

	poly := SetToPoly(S)
	n := len(S)
	a0 := (*poly)[0]
	a1 := (*poly)[1]

	var w1, w2 bn.G1Affine
	w1.MultiExp(pk1[:n], (*poly)[1:n+1], ecc.MultiExpConfig{NbTasks: ROUTINES})
	w2.MultiExp(pk1[:n-1], (*poly)[2:n+1], ecc.MultiExpConfig{NbTasks: ROUTINES})

	// log.Printf("sum=%v, a1/a0=%v\n", sum, new(fr.Element).Div(&a1, &a0))

	fS := ComputeAccG1(S, pk1)

	// e1 := bn.Pairing(pk2[1], w1)
	// e2 := bn.Pairing(pk2[0], fS.Add(pk1[0].Multiply(a0).Neg()))
	// log.Printf("e1==e2? %v\n", e1.Eq(e2))

	// e1 = bn.Pairing(pk2[1], w2)
	// e2 = bn.Pairing(pk2[0], w1.Add(pk1[0].Multiply(a1).Neg()))
	// log.Printf("e1==e2? %v\n", e1.Eq(e2))

	proof.A0.Set(&a0)
	proof.A1.Set(&a1)
	proof.W1.Set(&w1)
	proof.W2.Set(&w2)

	return *sum, fS, proof
}

func VerifySum(fS bn.G1Affine, sum fr.Element, proof SumProof, G1 bn.G1Affine, G2, sG2 bn.G2Affine) (bool, error) {

	if fS.Equal(&G1) {
		if sum.Equal(new(fr.Element).SetZero()) {
			return true, nil
		} else {
			return false, fmt.Errorf("non-zero sum for empty set")
		}
	}

	a0, a1, w1, w2 := proof.A0, proof.A1, proof.W1, proof.W2

	//fmt.Printf("sum=%v\nfS=%v\n", sum.String(), fS.String())
	//fmt.Printf("a0=%v\na1=%v\nw1=%v\nw2=%v\n", a0.String(), a1.String(), w1.String(), w2.String())

	if !sum.Equal(new(fr.Element).Div(&a1, &a0)) {
		return false, fmt.Errorf("sum != a1/a0")
	}

	// fS - a0*G1
	// w1 - a1*G1
	tmp1 := new(bn.G1Affine).ScalarMultiplication(&G1, a0.BigInt(new(big.Int)))
	tmp1.Sub(&fS, tmp1)
	tmp2 := new(bn.G1Affine).ScalarMultiplication(&G1, a1.BigInt(new(big.Int)))
	tmp2.Sub(&w1, tmp2)

	p1 := []bn.G1Affine{w1, *tmp1, w2, *tmp2}
	p2 := []bn.G2Affine{sG2, G2, sG2, G2}
	p1[0].Neg(&p1[0])
	p1[2].Neg(&p1[2])

	e1, err := bn.PairingCheck(p1, p2)
	if err != nil || !e1 {
		log.Panicf("pairing check error: %v\n", err)
		return false, fmt.Errorf("pairing check error: %v\n", err)
	}
	return true, nil
}
