package accumulator

import (
	"fmt"
	"math/big"
	"testing"

	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func TestGenKey(t *testing.T) {
	var s fr.Element
	sint, _ := new(big.Int).SetString("43247289572873481232976519249476583910", 10)
	s.SetBigInt(sint)

	n := 1500000
	//GenPubKeyToFile(&s, n)
	GenPubKey(&s, n)
}

func TestLoadKey(t *testing.T) {
	pk11, pk21 := LoadPubkey(50000)
	pk12, pk22 := LoadPubkeyFromFile(50000)

	for i := range 1000 {
		p11, p12 := pk11[i], pk12[i]
		p21, p22 := pk21[i], pk22[i]
		if !p11.Equal(&p12) || !p21.Equal(&p22) {
			fmt.Printf("%v false\n", i)
			return
		}

	}
	fmt.Printf("true\n")

}
