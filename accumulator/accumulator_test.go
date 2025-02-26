package accumulator

import (
	"fmt"
	"testing"

	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestComputeAcc(t *testing.T) {
	n := 800000
	var set fr.Vector = make(fr.Vector, n)
	for i := range set {
		set[i].SetInt64(int64(i + 5000000))
	}

	p := SetToPoly(set)

	pk1, pk2 := LoadPubkey(n + 1)
	// acc1 := ComputeAccG1(set, pk1)
	// acc2 := ComputeAccG2(set, pk2)
	acc1 := ComputePolyG1(p, pk1)
	acc2 := ComputePolyG2(p, pk2)

	fmt.Printf("acc1: %v\nacc2: %v\n", acc1.String(), acc2.String())
}

func TestSubset(t *testing.T) {
	var set1, set2 fr.Vector
	m, n := 20, 10
	set1 = make(fr.Vector, m)
	set2 = make(fr.Vector, n)
	for i := range m {
		set1[i].SetInt64(int64(i + 1))
	}
	for i := range n {
		set2[i].SetInt64(int64(i + 1))
	}
	pk1, pk2 := LoadPubkey(1000)
	f1, f2 := ComputeAccG1(set1, pk1), ComputeAccG1(set2, pk1)

	wit := ProveSubset(set1, set2, pk2)

	check := VerifySubset(f1, f2, wit)
	fmt.Printf("check subset: %v\n", check)
}

func NewSet(array []int) fr.Vector {
	set := make(fr.Vector, len(array))
	for i := range array {
		set[i].Set(new(fr.Element).SetInt64(int64(array[i])))
	}
	return set
}

func TestIntersection(t *testing.T) {
	S1 := NewSet([]int{1, 2, 3, 4, 5})
	S2 := NewSet([]int{3, 4, 5, 6, 7, 8})
	S3 := NewSet([]int{4, 5, 6, 7, 8, 9, 10})

	I := Intersection([]fr.Vector{S1, S2, S3})
	fmt.Printf("I: %v\n", I)

	pk1, pk2 := LoadPubkey(20)
	iproof := ProveIntersection([]fr.Vector{S1, S2, S3}, pk1, pk2)
	ver := VerifyIntersection(ComputeAccG1(I, pk1), iproof, pk1)
	fmt.Printf("verify intersection: %v\n", ver)
}

func TestSum(t *testing.T) {
	pk1, pk2 := LoadPubkey(100)
	fmt.Printf("sG2: %v\n", pk2[1].String())
	S := NewSet([]int{1, 2, 3, 4, 5})

	sum, fS, proof := ProveSum(S, pk1, pk2)
	// fmt.Printf("sum=%v\nfS=%v\n", sum.String(), fS.String())
	// fmt.Printf("a0=%v\na1=%v\nw1=%v\nw2=%v\n", proof.A0.String(), proof.A1.String(), proof.W1.String(), proof.W2.String())

	_, err := VerifySum(fS, sum, proof, pk1[0], pk2[0], pk2[1])
	if err != nil {
		fmt.Printf("false, %v\n", err)
	} else {
		fmt.Printf("true\n")
	}
}
