package accumulator

import (
	"fmt"
	"log"
	"math/big"
	"sort"
	"sync"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"golang.org/x/crypto/sha3"
)

var MIN = new(fr.Element).SetZero()
var INFint = new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
var INF = new(fr.Element).SetBigInt(INFint)

const SHORTBYTES = 8 // use uint64 for item key

type Item struct {
	Key   fr.Element
	Value interface{} // an accumulator Point or PointP
}

func (d *Item) ValueString() string {
	var res string
	switch v := d.Value.(type) {
	case bn.G1Affine:
		res = G1AffineToString(&v)
	case bn.G2Affine:
		res = G2AffineToString(&v)
	}
	return res
}

// interface for sort.Sort
type ItemList []Item

func (l ItemList) Len() int {
	return len(l)
}
func (l ItemList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l ItemList) Less(i, j int) bool {
	return l[i].Key.Cmp(&l[j].Key) == -1
}

type DicItem struct {
	Key   fr.Element  // current key
	Nxt   fr.Element  // the next key
	Value interface{} // an accumulator G1A or G2A
	//FI    bn.G1Affine // Hash(Item.Bytes()) * G1
	W bn.G2Affine // others * G2
}

func ToShort(a [32]byte) []byte { // 8-byte bytes, for key and nxt
	return a[32-SHORTBYTES:]
}

func ToLong(a [32]byte) []byte { // 32-byte bytes, for points
	return a[0:32]
}

func (d *DicItem) ToElement() *fr.Element {
	var dbytes []byte // key + nxt + value
	dbytes = append(dbytes, ToShort(d.Key.Bytes())...)
	dbytes = append(dbytes, ToShort(d.Nxt.Bytes())...)
	switch v := d.Value.(type) {
	case bn.G1Affine:
		dbytes = append(dbytes, ToLong(v.X.Bytes())...)
		dbytes = append(dbytes, ToLong(v.Y.Bytes())...)
	case bn.G2Affine:
		dbytes = append(dbytes, ToLong(v.X.A0.Bytes())...)
		dbytes = append(dbytes, ToLong(v.X.A1.Bytes())...)
		dbytes = append(dbytes, ToLong(v.Y.A0.Bytes())...)
		dbytes = append(dbytes, ToLong(v.Y.A1.Bytes())...)
	}

	h := sha3.NewLegacyKeccak256()
	h.Write(dbytes)
	return new(fr.Element).SetBytes(h.Sum(nil))
}

func (d *DicItem) ValueString() string {
	var res string
	switch v := d.Value.(type) {
	case bn.G1Affine:
		res = G1AffineToString(&v)
	case bn.G2Affine:
		res = G2AffineToString(&v)
	}
	return res
}

func CreateDic(items []Item, pk1 []bn.G1Affine, pk2 []bn.G2Affine) (dicItems []DicItem, dicDigest bn.G1Affine, err error) {
	n := len(items)

	if n == 0 {
		return []DicItem{}, pk1[0], nil
	}
	_, _, G1, G2 := bn.Generators()
	sort.Sort(ItemList(items))

	//fmt.Printf("items[0].Key: %v\n", items[0].Key.String())
	//fmt.Printf("items[n-1].Key: %v\n", items[n-1].Key.String())

	// items should not contain the MIN or INF item
	if items[0].Key.Equal(new(fr.Element).Set(MIN)) || items[n-1].Key.Equal(new(fr.Element).Set(INF)) {
		return dicItems, dicDigest, fmt.Errorf("items should not contain the MIN or INF key")
	}

	dicItems = make([]DicItem, n+1)
	switch items[0].Value.(type) {
	case bn.G1Affine:
		dicItems[0] = DicItem{Key: *MIN, Nxt: items[0].Key, Value: G1, W: *new(bn.G2Affine)}
	case bn.G2Affine:
		dicItems[0] = DicItem{Key: *MIN, Nxt: items[0].Key, Value: G2, W: *new(bn.G2Affine)}
	default:
		log.Panicf("item value should be G1Affine or G2Affine")
	}
	// Value=G1 denotes that there no data in [MIN, K0)
	for i := 0; i < n-1; i++ {
		dicItems[i+1] = DicItem{items[i].Key, items[i+1].Key, items[i].Value, *new(bn.G2Affine)}
	}
	dicItems[n] = DicItem{items[n-1].Key, *INF, items[n-1].Value, *new(bn.G2Affine)}

	//fmt.Printf("dicItems[0]: %v, %v\n", dicItems[0].Key.String(), dicItems[0].Nxt.String())
	//fmt.Printf("dicItems[n]: %v, %v\n", dicItems[n].Key.String(), dicItems[n].Nxt.String())
	dicDigest = CreateDicAuth(dicItems, pk1, pk2)

	return dicItems, dicDigest, nil
}

func CreateDicAuth(items []DicItem, pk1 []bn.G1Affine, pk2 []bn.G2Affine) bn.G1Affine {
	n := len(items)
	set := make(fr.Vector, n)
	for i := range items {
		set[i].Set((&items[i]).ToElement())
	}

	if n < 20 {
		for i := range items {
			Iset := make(fr.Vector, 1)
			Iset[0].Set((&items[i]).ToElement())
			Wset := Difference(set, Iset)
			w := ComputeAccG2(Wset, pk2)
			items[i].W.Set(&w)
		}
		dicDigest := ComputeAccG1(set, pk1)
		return dicDigest
	}

	chNum := 16
	chTask := n / chNum
	if n%chNum != 0 {
		chTask += 1
	}
	type SW struct {
		index int
		wit   bn.G2Affine
	}
	chResult := make(chan []SW, chNum)
	var wg sync.WaitGroup

	for i := range chNum {
		//log.Printf("channel %v started\n", i)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			iStart := i * chTask
			iEnd := iStart + chTask
			if iEnd > n {
				iEnd = n
			}
			var subW []SW
			for j := iStart; j < iEnd; j++ {
				Iset := make(fr.Vector, 1)
				Iset[0].Set(&set[j])
				Wset := Difference(set, Iset)
				subW = append(subW, SW{j, ComputeAccG2(Wset, pk2)})
			}
			chResult <- subW
			//log.Printf("channel %v finished\n", i)
		}(i)
	}

	go func() {
		wg.Wait()
		close(chResult)
	}()
	for subW := range chResult {
		for _, sw := range subW {
			i := sw.index
			w := sw.wit
			items[i].W.Set(&w)
		}
	}

	dicDigest := ComputeAccG1(set, pk1)
	return dicDigest
}

// fulfill D and W in dic items
func CreateDicAuthNoCurrent(items []DicItem, pk1 []bn.G1Affine, pk2 []bn.G2Affine) bn.G1Affine {
	set := make(fr.Vector, len(items))
	for i := range items {
		set[i].Set((&items[i]).ToElement())
	}
	// n := len(set)
	for i := range set {
		Iset := make(fr.Vector, 1)
		Iset[0].Set(&set[i])
		Wset := Difference(set, Iset)
		w := ComputeAccG2(Wset, pk2)
		items[i].W.Set(&w)
	}

	dicDigest := ComputeAccG1(set, pk1)
	return dicDigest
}

func VerifyDic(dicDigest bn.G1Affine, dicItem DicItem, pk1 []bn.G1Affine) bool {
	set := make(fr.Vector, 1)
	set[0].Set(dicItem.ToElement())
	d := ComputeAccG1(set, pk1)

	_, _, _, G2 := bn.Generators()
	p1 := []bn.G1Affine{d, dicDigest}
	p2 := []bn.G2Affine{dicItem.W, G2}
	p1[0].Neg(&p1[0])
	e1, err := bn.PairingCheck(p1, p2)
	if err != nil || !e1 {
		return false
	}
	return true
}
