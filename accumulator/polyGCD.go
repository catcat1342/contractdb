package accumulator

import (
	_ "github.com/go-sql-driver/mysql"
)

type Matrix [2][2]Poly

func NewMatrix() Matrix {
	p11, p12, p21, p22 := new(Poly), new(Poly), new(Poly), new(Poly)
	p11.SetOne()
	p12.SetZero()
	p21.SetZero()
	p22.SetOne()

	elts := [2][2]Poly{{*p11, *p12}, {*p21, *p22}}
	return Matrix(elts)
}

func (m Matrix) SetUnit() {
	m[0][0].SetOne()
	m[0][1].SetZero()
	m[1][0].SetZero()
	m[1][1].SetOne()
}

func (m Matrix) PrintString() string {
	var res string
	res = m[0][0].PrintString()
	res += "  "
	res += m[0][1].PrintString()
	res += "\n"
	res += m[1][0].PrintString()
	res += "  "
	res += m[1][1].PrintString()
	return res
}

// XGCD: d=gcd(a,b), a·s+b·t=d

func XGCD(a *Poly, b *Poly) (d *Poly, s *Poly, t *Poly, err error) {
	if a.IsZero() && b.IsZero() {
		d.SetZero()
		s.SetOne()
		t.SetZero()
		err = nil
		return
	}
	u, v := new(Poly), new(Poly)
	u.Set(a)
	v.Set(b)

	q := new(Poly)
	flag := 0
	if u.Len() == v.Len() {
		q, u, _ = PolyDivRem(u, v) // u / v = q ... r
		u, v = v, u
		flag = 1
	} else if u.Len() < v.Len() {
		u, v = v, u
		flag = 2
	}

	// Matrix M
	M, u := IterHalfGCD(u, v, u.Len())

	d = u
	if flag == 0 {
		s = &M[0][0]
		t = &M[0][1]
	} else if flag == 1 {
		s = &M[0][1]
		t = new(Poly).Sub(&M[0][0], new(Poly).Mul(q, &M[0][1]))
	} else {
		s = &M[0][1]
		t = &M[0][0]
	}

	// normalize
	w := new(Poly)
	w.SetLen(1)
	(*w)[0].Inverse(&(*d)[d.Len()-1])
	d.Mul(d, w)
	s.Mul(s, w)
	t.Mul(t, w)

	// check the result
	// fmt.Printf("d: %v\ns:%v\nt:%v\n", d.PrintString(), s.PrintString(), t.PrintString())
	// t1, t2 := NewPoly(), NewPoly()
	// t1.Mul(a, s)
	// t2.Mul(b, t)
	// fmt.Printf("check beizu: %v\n", NewPoly().Add(t1, t2).Equal(d))

	return
}

func IterHalfGCD(u *Poly, v *Poly, dRed int) (M Matrix, u1 *Poly) {

	M = NewMatrix()
	goal := u.Len() - 1 - dRed

	//fmt.Printf("u: %v, v: %v, goal: %v\n", u.PrintString(), v.PrintString(), goal)

	if v.IsZero() || v.Len()-1 <= goal {
		return M, u
	}

	q := new(Poly)
	for v.Len()-1 > goal {
		q, u, _ = PolyDivRem(u, v)
		u, v = v, u

		t := new(Poly)
		t.Sub(&M[0][0], new(Poly).Mul(q, &M[1][0]))
		M[0][0].Set(&M[1][0])
		M[1][0].Set(t)

		t.Sub(&M[0][1], new(Poly).Mul(q, &M[1][1]))
		M[0][1].Set(&M[1][1])
		M[1][1].Set(t)
	}
	return M, u
}
