package accumulator

import (
	"fmt"
	"math"

	"github.com/consensys/gnark-crypto/ecc"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
	_ "github.com/go-sql-driver/mysql"
)

const K0 = 7 // crossover = 1 << K0
// N=1000000, K0=7 is the best choice
const MulK0 = 10
const FFT_TASKS = 8

type Poly fr.Vector // P(x) = \sum (Coeffs[i] \cdot x^i)

func (p *Poly) Len() int {
	return fr.Vector(*p).Len()
}

type MatrixP2 struct {
	Items [][]Poly
}

func (p *Poly) Set(p1 *Poly) *Poly {
	*p = Poly(make(fr.Vector, p1.Len()))
	copy(*p, *p1)
	return p
}

func (p *Poly) SetLen(l int) *Poly {
	*p = Poly(make(fr.Vector, l))
	return p
}

func (p *Poly) SetInt(array []int) *Poly {
	*p = Poly(make(fr.Vector, len(array)))
	for i := range array {
		(*p)[i].SetInt64(int64(array[i]))
	}
	return p
}

func (p *Poly) SetZero() *Poly {
	*p = Poly(make(fr.Vector, 1))
	(*p)[0].SetZero()
	return p
}

func (p *Poly) SetOne() *Poly {
	*p = Poly(make(fr.Vector, 1))
	(*p)[0].SetOne()
	return p
}

func (p *Poly) Equal(other *Poly) bool {
	if p.Len() != other.Len() {
		return false
	}
	for i := range *p {
		if !(*p)[i].Equal(&(*other)[i]) {
			return false
		}
	}
	return true
}

func (p *Poly) IsZero() bool {
	zero := fr.NewElement(0)
	if p == nil || p.Len() == 0 || (p.Len() == 1 && (*p)[0].Equal(&zero)) {
		return true
	}
	return false
}

func (p *Poly) Normalize() *Poly {
	if p == nil {
		p.SetZero()
		return p
	}
	zero := fr.NewElement(0)
	for p.Len() > 0 && (*p)[p.Len()-1].Equal(&zero) {
		*p = (*p)[:p.Len()-1]
	}
	return p
}

func (p *Poly) PrintString() string {
	if p == nil || p.Len() == 0 {
		return "0"
	}
	var res string
	for i := p.Len() - 1; i >= 1; i-- {
		one := fr.NewElement(1)
		if (*p)[i].Equal(&one) {
			res += fmt.Sprintf("x^%v + ", i)
		} else {
			res += fmt.Sprintf("%vx^%v + ", (*p)[i].String(), i)
		}

	}
	res += fmt.Sprintf("%v", (*p)[0].String())
	return res
}

// p=x+y  p.Add(x,y)
func (p *Poly) Add(x, y *Poly) *Poly {
	l1, l2 := new(Poly), new(Poly)
	if x.Len() >= y.Len() {
		l1, l2 = x, y
	} else {
		l1, l2 = y, x
	}
	p1 := new(Poly)
	p1.Set(l1)
	for i := 0; i < l2.Len(); i++ {
		(*p1)[i].Add(&(*l1)[i], &(*l2)[i])
	}
	p1.Normalize()

	*p = Poly(make(fr.Vector, len(*p1)))
	copy(*p, *p1)
	return p
}

// p.Sub(x,y): p=x-y
func (p *Poly) Sub(x, y *Poly) *Poly {
	p1 := new(Poly)
	if x.Len() < y.Len() {
		p1.SetLen(y.Len())
		for i := 0; i < x.Len(); i++ {
			(*p1)[i].Sub(&(*x)[i], &(*y)[i])
		}
		var zero fr.Element
		zero.SetUint64(0)
		for i := x.Len(); i < y.Len(); i++ {
			(*p1)[i].Sub(&zero, &(*y)[i])
		}
	} else {
		p1.Set(x)
		for i := 0; i < y.Len(); i++ {
			(*p1)[i].Sub(&(*x)[i], &(*y)[i])
		}
	}
	p1.Normalize()

	*p = Poly(make(fr.Vector, len(*p1)))
	copy(*p, *p1)
	return p
}

// p.Mul(x,y): p=x*y
// should be optimized by FFT
func (p *Poly) Mul(x, y *Poly) *Poly {
	if p == nil {
		p.SetZero()
	}
	if x.IsZero() || y.IsZero() {
		p.SetZero()
		return p
	}

	crossover := 1 << MulK0
	if x.Len()+y.Len() > crossover {
		return p.MulFFT(x, y)
	}

	p1 := new(Poly)
	p1.SetLen(x.Len() + y.Len() - 1)

	for i := 0; i < x.Len(); i++ {
		for j := 0; j < y.Len(); j++ {
			(*p1)[i+j] = *new(fr.Element).Add(&(*p1)[i+j], new(fr.Element).Mul(&(*x)[i], &(*y)[j]))
		}
	}
	p1.Normalize()

	*p = Poly(make(fr.Vector, len(*p1)))
	copy(*p, *p1)
	return p
}

func (p *Poly) MulFFT(x, y *Poly) *Poly {
	if p == nil {
		p.SetZero()
	}
	if x.IsZero() || y.IsZero() {
		p.SetZero()
		return p
	}

	n := len(*x)
	m := len(*y)

	N := ecc.NextPowerOfTwo(uint64(n + m - 1))
	x1, y1 := new(Poly), new(Poly)
	*x1 = Poly(make(fr.Vector, N))
	*y1 = Poly(make(fr.Vector, N))
	copy(*x1, *x)
	copy(*y1, *y)

	domain := fft.NewDomain(N)
	fft.BitReverse(*x1)
	fft.BitReverse(*y1)
	domain.FFT(*x1, fft.DIT, fft.WithNbTasks(FFT_TASKS))
	domain.FFT(*y1, fft.DIT, fft.WithNbTasks(FFT_TASKS))

	for i := uint64(0); i < N; i++ {
		(*x1)[i].Mul(&(*x1)[i], &(*y1)[i])
	}
	domain.FFTInverse(*x1, fft.DIF, fft.WithNbTasks(FFT_TASKS))
	fft.BitReverse(*x1)

	*p = Poly((*x1)[:n+m-1])
	return p
}

// x / y = p ... r
func PolyDivRem(x, y *Poly) (*Poly, *Poly, error) {
	x.Normalize()
	y.Normalize()

	p, r := new(Poly), new(Poly)

	if y.IsZero() {
		p.SetZero()
		r.SetZero()
		return p, r, fmt.Errorf("cannot division by zero")
	}
	if x.Len() < y.Len() {
		p.SetZero()
		r.Set(y)
		return p, r, nil
	}

	// results
	p.SetLen(x.Len() - y.Len() + 1) //  quotient
	r.Set(x)                        //  remainder

	for r.Len() >= y.Len() {
		var qLead fr.Element = *new(fr.Element).Mul(&(*r)[r.Len()-1], new(fr.Element).Inverse(&(*y)[y.Len()-1]))
		// fmt.Printf("qLead: %v\n", qLead.String())
		degDiff := r.Len() - y.Len()
		(*p)[degDiff] = qLead
		for i := y.Len() - 1; i >= 0; i-- {
			(*r)[i+degDiff] = *new(fr.Element).Sub(&(*r)[i+degDiff], new(fr.Element).Mul(&(*y)[i], &qLead))
		}
		r.Normalize()
	}
	return p, r, nil
}

func (p *Poly) IterBuild(r fr.Vector) *Poly {
	n := r.Len()
	if n == 0 {
		return p.SetOne()
	}

	nr := make(fr.Vector, r.Len())
	for i := range r {
		nr[i].Neg(&r[i])
	}

	p.SetLen(n + 1)
	for i := 0; i < n+1; i++ {
		(*p)[i].SetZero()
	}
	(*p)[0].Set(&nr[0])
	(*p)[1].SetOne()

	for k := 1; k <= n-1; k++ {
		(*p)[k+1] = (*p)[k]
		for i := k; i >= 1; i-- {
			(*p)[i].Mul(&(*p)[i], &nr[k]).Add(&(*p)[i], &(*p)[i-1])
		}
		(*p)[0].Mul(&(*p)[0], &nr[k])
	}
	return p
}

// P(x) = \sum (Coeffs[i] \cdot x^i)
func (p *Poly) BuildFromRoots(a fr.Vector) *Poly {

	n := a.Len()
	crossover := 1 << K0
	if n <= crossover {
		return p.IterBuild(a)
	}

	m := int(ecc.NextPowerOfTwo(uint64(n)))
	k := int(math.Log2(float64(m))) // n=3000 ->  m=4096, k=12

	// Initialize the polynomial with the roots, extended to length m+1
	b := make(fr.Vector, m+1)
	copy(b, a)
	for i := n; i < m; i++ {
		b[i].SetZero()
	}
	b[m].SetOne()

	// start := time.Now()
	for i := 0; i < m; i += crossover {
		g, h := make(fr.Vector, crossover), make(fr.Vector, crossover)
		for j := 0; j < crossover; j++ {
			g[j].Neg(&b[i+j])
		}
		// fmt.Printf("g: %v\n", printVector(g))
		if K0 > 0 {
			for j := 0; j < crossover; j += 2 {
				t1 := new(fr.Element).Mul(&g[j], &g[j+1])
				g[j+1].Add(&g[j], &g[j+1])
				g[j].Set(t1)
			}
		}
		// fmt.Printf("g: %v\n", printVector(g))
		for l := 1; l < K0; l++ {
			width := 1 << l
			ph, p1, p2 := new(Poly), new(Poly), new(Poly)
			for j := 0; j < crossover; j += 2 * width {
				// mul(&h[j], &g[j], &g[j + width], width);
				*p1 = make(Poly, width+1)
				*p2 = make(Poly, width+1)
				copy(*p1, g[j:j+width])
				copy(*p2, g[j+width:j+2*width])
				(*p1)[width].SetOne()
				(*p2)[width].SetOne()
				ph.Mul(p1, p2)
				// fmt.Printf("p1: %v\n", printVector(fr.Vector(*p1)))
				// fmt.Printf("p2: %v\n", printVector(fr.Vector(*p2)))
				// fmt.Printf("ph: %v\n", printVector(fr.Vector(*ph)))
				for i := 0; i < 2*width; i++ {
					h[j+i].Set(&(*ph)[i])
				}
				// fmt.Printf("h: %v\n", printVector(h))
				// fmt.Printf("for pause")
			}
			g, h = h, g
			// fmt.Printf("g: %v\n", printVector(g))
			// fmt.Printf("h: %v\n", printVector(h))
		}
		for j := 0; j < crossover; j++ {
			b[i+j].Set(&g[j])
		}
	}

	// elapsed := time.Since(start)
	// fmt.Printf("time: %v\n", elapsed)
	// start = time.Now()

	// fmt.Printf("b: %v\n", printVector(b))
	one := new(fr.Element).SetOne()
	//domain := fft.NewDomain(1 << (k + 1))
	for l := K0; l < k; l++ {
		width := 1 << l
		domain := fft.NewDomain(1 << (l + 1))
		var b1, b2 fr.Vector
		for i := 0; i < m; i += 2 * width {
			b1, b2 = make(fr.Vector, 2*width), make(fr.Vector, 2*width)
			copy(b1, b[i:i+width])
			b1[width].SetOne()
			copy(b2, b[i+width:i+2*width])
			b2[width].SetOne()
			// fmt.Printf("b1: %v\n", printVector(b1))
			// fmt.Printf("b2: %v\n", printVector(b2))
			fft.BitReverse(b1)
			fft.BitReverse(b2)
			domain.FFT(b1, fft.DIT, fft.WithNbTasks(FFT_TASKS))
			domain.FFT(b2, fft.DIT, fft.WithNbTasks(FFT_TASKS))
			for j := 0; j < 2*width; j++ {
				b1[j].Mul(&b1[j], &b2[j]) // b1=b1*b2
			}
			domain.FFTInverse(b1, fft.DIF, fft.WithNbTasks(FFT_TASKS))
			fft.BitReverse(b1)

			for j := 0; j < 2*width; j++ {
				b[i+j].Set(&b1[j])
			}
			b[i].Sub(&b[i], one)
			// fmt.Printf("b: %v\n", printVector(b))
		}
	}

	// elapsed = time.Since(start)
	// fmt.Printf("time: %v\n", elapsed)

	*p = Poly(make(fr.Vector, n+1))
	delta := m - n
	for i := range n + 1 {
		(*p)[i].Set(&b[i+delta])
	}

	return p
}

// func printVector(v fr.Vector) string {
// 	res := ""
// 	for i := 0; i < v.Len()-1; i++ {
// 		res += fmt.Sprintf("%v, ", v[i].String())
// 	}
// 	res += v[v.Len()-1].String()
// 	return res
// }
