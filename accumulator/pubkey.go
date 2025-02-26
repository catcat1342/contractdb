package accumulator

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/syndtr/goleveldb/leveldb"
)

const BaseDir = "/home/ubuntu/contractdb"
const pubfilesize = 10000

func G1AffineToString(g *bn.G1Affine) string {
	return g.X.String() + " " + g.Y.String()
}

func G2AffineToString(g *bn.G2Affine) string {
	return g.X.A0.String() + " " + g.X.A1.String() + " " + g.Y.A0.String() + " " + g.Y.A1.String()
}

func StringsToG1Affine(sx, sy string) bn.G1Affine {
	var g1 bn.G1Affine
	g1.X.SetString(sx)
	g1.Y.SetString(sy)
	return g1
}

func StringsToG2Affine(s1, s2, s3, s4 string) bn.G2Affine {
	var g2 bn.G2Affine
	g2.X.A0.SetString(s1)
	g2.X.A1.SetString(s2)
	g2.Y.A0.SetString(s3)
	g2.Y.A1.SetString(s4)

	return g2
}

func StringToG1Affine(s string) bn.G1Affine {
	ss := strings.Split(s, " ")
	var g1 bn.G1Affine
	g1.X.SetString(ss[0])
	g1.Y.SetString(ss[1])
	return g1
}

func StringToG2Affine(s string) bn.G2Affine {
	ss := strings.Split(s, " ")
	var g2 bn.G2Affine
	g2.X.A0.SetString(ss[0])
	g2.X.A1.SetString(ss[1])
	g2.Y.A0.SetString(ss[2])
	g2.Y.A1.SetString(ss[3])

	return g2
}

func GenPubKey(s *fr.Element, MAX_LEN int) error {
	if MAX_LEN == 0 {
		panic("GenKey cannot accept n=0")
	}

	// log.Println("GEN PUB KEYS ...")
	// var si fr.Vector = make(fr.Vector, MAX_LEN)
	// si[0].SetInt64(1)
	// for i := 1; i < MAX_LEN; i++ {
	// 	si[i].Mul(&si[i-1], s)
	// }
	// _, _, G1, G2 := bn.Generators()
	// pk1 := bn.BatchScalarMultiplicationG1(&G1, si)
	// pk2 := bn.BatchScalarMultiplicationG2(&G2, si)
	// log.Printf("\n...GEN PUB KEYS FINISHED")

	pk1, pk2 := LoadPubkeyFromFile(MAX_LEN)

	log.Printf("write pubkey to levelDB\n")

	pkDB := filepath.Join(BaseDir, "authdb", "pubkey") // store pubkey to authdb
	_, err := os.Stat(pkDB)
	if err == nil || os.IsExist(err) {
		os.RemoveAll(pkDB)
	}
	db, err := leveldb.OpenFile(pkDB, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return err
	}
	defer db.Close()

	batch := new(leveldb.Batch)
	batchSize := 10000
	batchCount := 0
	for i := range pk1 {
		key := UintToBytes(uint64(i))
		v := G1AffineToString(&pk1[i])
		v += "|"
		v += G2AffineToString(&pk2[i])

		batch.Put(key, []byte(v))
		batchCount++

		if batchCount >= batchSize || i == len(pk1)-1 {
			err = db.Write(batch, nil)
			if err != nil {
				log.Fatalf("Failed to write batch: %v", err)
				return err
			}
			batch.Reset()
			batchCount = 0
		}
	}

	return nil
}

func UintToBytes(num uint64) []byte {
	buf := make([]byte, 8) // uint64 占用 8 个字节
	binary.BigEndian.PutUint64(buf, num)
	return buf
}

func LoadPubkey(n int) ([]bn.G1Affine, []bn.G2Affine) {
	var pk1 []bn.G1Affine
	var pk2 []bn.G2Affine

	pkDB := filepath.Join(BaseDir, "authdb", "pubkey") // store pubkey to authdb

	db, err := leveldb.OpenFile(pkDB, nil)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return nil, nil
	}
	defer db.Close()

	count := 0
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		if count == n {
			break
		}
		valBytes := iter.Value()
		valString := string(valBytes)
		pks := strings.Split(valString, "|")
		pk1 = append(pk1, StringToG1Affine(pks[0]))
		pk2 = append(pk2, StringToG2Affine(pks[1]))
		count += 1
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		log.Fatalf("Iterator error: %v", err)
	}

	return pk1, pk2
}

func GenPubKeyToFile(s *fr.Element, MAX_LEN int) {
	if MAX_LEN == 0 {
		panic("GenKey cannot accept n=0")
	}

	log.Println("GEN PUB KEYS ...")

	var si fr.Vector = make(fr.Vector, MAX_LEN)
	si[0].SetInt64(1)
	for i := 1; i < MAX_LEN; i++ {
		si[i].Mul(&si[i-1], s)
	}

	_, _, G1, G2 := bn.Generators()

	pk1 := bn.BatchScalarMultiplicationG1(&G1, si)
	pk2 := bn.BatchScalarMultiplicationG2(&G2, si)
	log.Printf("\n...GEN PUB KEYS FINISHED")

	// load pk1,pk2 to pubkey files
	filenum := math.Ceil(float64(MAX_LEN) / float64(pubfilesize))
	// baseDir := os.Getenv("PROJECT_ROOT")
	//baseDir := "/home/ubuntu/contractdb"

	for i := range int(filenum) {
		filename := filepath.Join(BaseDir, "pubkey", fmt.Sprintf("pubkey.tbl.%v", i))
		file, err := os.Create(filename)
		if err != nil {
			log.Panicf("Error opening file %s: %v\n", filename, err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		for j := 0; j < pubfilesize; j++ {
			k := i*pubfilesize + j
			if k >= MAX_LEN {
				break
			}
			line := fmt.Sprintf("%v|%v|%v|%v|%v|%v\n", pk1[k].X.String(), pk1[k].Y.String(), pk2[k].X.A0.String(), pk2[k].X.A1.String(), pk2[k].Y.A0.String(), pk2[k].Y.A1.String())
			_, err := writer.WriteString(line)
			if err != nil {
				log.Panicf("Error writing file %s: %v", filename, err)
			}
		}
		err = writer.Flush()
		if err != nil {
			log.Panicf("Error writer flush: %v", err)
		}
	}
}

// if n>len(pk1), return all pubkeys
func LoadPubkeyFromFile(n int) ([]bn.G1Affine, []bn.G2Affine) {
	var pk1 []bn.G1Affine
	var pk2 []bn.G2Affine

	//baseDir := "/home/ubuntu/contractdb"

	i := 0
	count := 0
	for {
		filename := filepath.Join(BaseDir, "pubkey", fmt.Sprintf("pubkey.tbl.%v", i))
		file, err := os.Open(filename)
		if err != nil {
			log.Panicf("Error opening file %s: %v\n", filename, err)
			return pk1, pk2

		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			ks := strings.Split(line, "|")
			pk1 = append(pk1, StringsToG1Affine(ks[0], ks[1]))
			pk2 = append(pk2, StringsToG2Affine(ks[2], ks[3], ks[4], ks[5]))
			count += 1
			if count >= n {
				break
			}
		}
		if count >= n {
			pk1 = pk1[:n]
			pk2 = pk2[:n]
			break
		}
		i += 1
	}
	return pk1, pk2
}
