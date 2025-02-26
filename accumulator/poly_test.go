package accumulator

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	_ "github.com/go-sql-driver/mysql"
)

func TestPolyBasic(t *testing.T) {
	p1, p2 := new(Poly), new(Poly)
	p1.SetLen(4)
	(*p1)[0].SetUint64(2)
	(*p1)[1].SetUint64(1)
	(*p1)[2].SetUint64(5)
	(*p1)[3].SetUint64(3)
	p2.SetLen(2)
	(*p2)[0].SetUint64(1)
	(*p2)[1].SetUint64(1)
	fmt.Printf("p1: %v\n", p1.PrintString())
	fmt.Printf("p2: %v\n", p2.PrintString())

	p := new(Poly)
	fmt.Printf("p1+p2: %v\n", p.Add(p1, p2).PrintString())
	fmt.Printf("p1-p2: %v\n", p.Sub(p1, p2).PrintString())
	fmt.Printf("p2-p1: %v\n", p.Sub(p2, p1).PrintString())
	fmt.Printf("p1*p2: %v\n", p.Mul(p1, p2).PrintString())
	fmt.Printf("p2*p1: %v\n", p.Mul(p2, p1).PrintString())

	r := new(Poly)
	var err error
	p, r, err = PolyDivRem(p1, p2)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf("p1/p2, p: %v, r: %v\n", p.PrintString(), r.PrintString())
	}
}

func TestPolyBuild(t *testing.T) {
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer f.Close()
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	runtime.GOMAXPROCS(64)

	N := 1 << 19
	roots := make(fr.Vector, N)
	for i := range roots {
		roots[i].SetUint64(uint64(i) + 1)
	}
	p1 := new(Poly)

	start := time.Now()
	p1.BuildFromRoots(roots)
	elapsed := time.Since(start)

	fmt.Printf("N=%v, fft time: %v\n", N, elapsed)
}

func TestMatrix(t *testing.T) {
	m := NewMatrix()
	fmt.Printf("m: %v\n", m.PrintString())
}

func TestMulFFT(t *testing.T) {
	m, n := 1<<15, 1024
	p1, p2 := new(Poly), new(Poly)
	*p1 = Poly(make(fr.Vector, m))
	*p2 = Poly(make(fr.Vector, n))

	for i := range m {
		(*p1)[i].SetInt64(int64(i + 10000))
	}
	for i := range n {
		(*p2)[i].SetInt64(int64(i + 59587412))
	}

	m1, m2 := new(Poly), new(Poly)

	start := time.Now()
	m1.Mul(p1, p2)
	elapsed1 := time.Since(start)
	start = time.Now()
	m2.MulFFT(p1, p2)
	elapsed2 := time.Since(start)

	fmt.Printf("m=%v, n=%v, iter mul: %v, fft mul: %v\n", m, n, elapsed1, elapsed2)

}

func TestGCD(t *testing.T) {
	var p11, p12, p22 *Poly
	p11.SetInt([]int{1, 1, 1, 1})
	p12.SetInt([]int{2, 1, 1})
	// p22 := NewPolyInt([]int{2, 1, 1})
	p22.SetZero()

	var p1, p2 *Poly
	p1.Mul(p11, p12)
	p2.Mul(p11, p22)

	XGCD(p1, p2)
}
