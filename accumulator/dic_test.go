package accumulator

import (
	"fmt"
	"testing"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	"golang.org/x/exp/rand"
)

func TestDicBytes(t *testing.T) {
	pk1, pk2 := LoadPubkey(20)

	item := new(DicItem)
	item.Key.SetUint64(uint64(5959))
	item.Nxt.SetUint64(uint64(794556))
	item.Value = *new(bn.G1Affine).Set(&pk1[0])
	item.W.Set(&pk2[0])

	elem := item.ToElement()
	fmt.Printf("elem: %v\n", elem.String())
}

func TestDicAuth(t *testing.T) {
	pk1, pk2 := LoadPubkey(20)

	items := make([]Item, 1000)

	for i := range items {
		randKey := rand.Uint64()
		items[i].Key.SetUint64(randKey)
		items[i].Value = *new(bn.G1Affine).Set(&pk1[0])
	}

	dicItems, dicDigest, _ := CreateDic(items, pk1, pk2)
	fmt.Printf("dic digest: %v\n", dicDigest.String())
	fmt.Printf("first 10 items: \n")
	for i := range 10 {
		fmt.Printf("item[%v]: key=%v, nxt=%v, value=%v", i, dicItems[i].Key.String(), dicItems[i].Nxt.String(), dicItems[i].ValueString())
	}
}
