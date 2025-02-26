package contract

import (
	"context"
	acc "contractdb/accumulator"
	ads "contractdb/ads"
	"fmt"
	"log"
	"math/big"
	"path/filepath"

	bn "github.com/consensys/gnark-crypto/ecc/bn254"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/syndtr/goleveldb/leveldb"

	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const accountAddr = "fa2a7eeeac24d706f4b925430f9e8025ef5fd0c6c9fc5ca6e3e48f6dbb71ebed"
const contractAddr = "0x485902042071B0238F73CbA55681faFdC8103b74"

func TestContract(t *testing.T) {
	ind := 33
	res, q := genInput(ind)
	param := toParam(res)

	fmt.Printf("test gas cost for index %v\n", ind)

	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal(err)
	}
	contractAddress := common.HexToAddress(contractAddr)
	instance, err := NewContract(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// // debug function
	// result, err := instance.VerifyQuery(&bind.CallOpts{Context: context.Background()}, q, param)
	// if err != nil {
	// 	log.Fatalf("Failed to call VerifyQuery: %v", err)
	// }
	// fmt.Printf("DEBUG: result: %v, flag: %v\n", result.Ver, result.Flag)

	// test gas cost
	addr := accountAddr
	testGas(client, instance, addr, q, param)
}

func genInput(ind int) (ads.SumRes1d, ContractDBQuery) {
	pk1, pk2 := acc.LoadPubkey(10000)
	q := ads.Query{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	switch ind {
	case 1:
		q.Cond = []string{"value"}
		q.CondFlag = []int{1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}}
	case 2:
		q.Cond = []string{"value", "rate"}
		q.CondFlag = []int{1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}, []uint64{2000, 2500}}
	case 3:
		q.Cond = []string{"value", "rate", "grade"}
		q.CondFlag = []int{1, 1, 1}
		q.CondVal = []interface{}{[]uint64{1000, 2000}, []uint64{2000, 2500}, []uint64{3000, 3100}}
	case 10:
		q.Cond = []string{"name"}
		q.CondFlag = []int{0}
		q.CondVal = []interface{}{"Name3948"}
	case 11:
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}}
	case 12:
		q.Cond = []string{"name", "value", "rate"}
		q.CondFlag = []int{0, 1, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 13:
		q.Cond = []string{"name", "value", "rate", "grade"}
		q.CondFlag = []int{0, 1, 1, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	case 20:
		q.Cond = []string{"name", "bank"}
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Name3948", "Bank564"}
	case 21:
		q.Cond = []string{"name", "bank", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", []uint64{4000, 5000}}
	case 22:
		q.Cond = []string{"name", "bank", "value", "rate"}
		q.CondFlag = []int{0, 0, 1, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 23:
		q.Cond = []string{"name", "bank", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	case 30:
		q.Cond = []string{"name", "bank", "addr"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Name3948", "Bank564", "Addr221"}
	case 31:
		q.Cond = []string{"name", "bank", "addr", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", "Addr221", []uint64{4000, 5000}}
	case 32:
		q.Cond = []string{"name", "bank", "addr", "value", "rate"}
		q.CondFlag = []int{0, 0, 0, 1, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", "Addr221", []uint64{4000, 5000}, []uint64{2000, 3000}}
	case 33:
		q.Cond = []string{"name", "bank", "addr", "value", "rate", "grade"}
		q.CondFlag = []int{0, 0, 0, 1, 1, 1}
		q.CondVal = []interface{}{"Name3948", "Bank564", "Addr221", []uint64{4000, 5000}, []uint64{2000, 3000}, []uint64{3000, 4000}}
	default:
		panic("invalid input type")
	}

	res, err := ads.QueryDB(ads.DBINFO, q, pk1, pk2)
	if err != nil {
		log.Printf("query db failed: %v\n", err)
		return ads.SumRes1d{}, ContractDBQuery{}
	}
	verres, err := ads.Verify(ads.DBINFO, q, res, pk1, pk2)
	if err != nil {
		log.Printf("verify failed: %v\n", err)
		return ads.SumRes1d{}, ContractDBQuery{}
	}
	sum := res.(ads.SumRes1d).Sum
	fmt.Printf("result: %v, verres: %v\n", sum.String(), verres)
	return res.(ads.SumRes1d), toContractQuery(q)
}

func toContractQuery(q ads.Query) ContractDBQuery {
	ind := make(map[string]uint8)
	ind["name"] = 0
	ind["bank"] = 1
	ind["addr"] = 2
	ind["value"] = 3
	ind["rate"] = 4
	ind["grade"] = 5

	eind := []uint8{}
	rind := []uint8{}
	rtype := []uint8{} // indicating type (1,2,3,4) of rind
	eval := []uint64{}
	rval := []uint64{}

	for i, f := range q.CondFlag {
		switch f {
		case 0:
			eind = append(eind, ind[q.Cond[i]])
			eval = append(eval, new(fr.Element).SetBytes([]byte(q.CondVal[i].(string))).Uint64())
		case 1, 2, 3, 4:
			rind = append(rind, ind[q.Cond[i]])
			rtype = append(rtype, uint8(f))
			rval = append(rval, q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1])
		default:
			panic("invalid query")
		}
	}
	return ContractDBQuery{
		Eind:  eind,
		Rind:  rind,
		Rtype: rtype,
		Eval:  eval,
		Rval:  rval,
	}
}

func toG1Point(p bn.G1Affine) ContractDBG1Point {
	return ContractDBG1Point{
		X: p.X.BigInt(new(big.Int)),
		Y: p.Y.BigInt(new(big.Int)),
	}
}

func toG2Point(p bn.G2Affine) ContractDBG2Point {
	c0 := p.X.A0.BigInt(new(big.Int))
	c1 := p.X.A1.BigInt(new(big.Int))
	c2 := p.Y.A0.BigInt(new(big.Int))
	c3 := p.Y.A1.BigInt(new(big.Int))
	return ContractDBG2Point{
		X: [2]*big.Int{c1, c0}, // don't forget to exchange coeffs
		Y: [2]*big.Int{c3, c2},
	}
}

func toDicItemE(d acc.DicItem) ContractDBDicItemE {
	key := d.Key.Uint64()
	next := d.Nxt.Uint64()
	val := toG1Point(d.Value.(bn.G1Affine))
	return ContractDBDicItemE{
		Key:   key,
		Nxt:   next,
		Value: val,
	}
}

func toDicItemR(d acc.DicItem) ContractDBDicItemR {
	key := d.Key.Uint64()
	next := d.Nxt.Uint64()
	val := toG2Point(d.Value.(bn.G2Affine))
	return ContractDBDicItemR{
		Key:   key,
		Nxt:   next,
		Value: val,
	}
}

func toItemInv(item acc.DicItem) *big.Int {
	res := new(big.Int)
	itemInv := new(fr.Element).Inverse(item.ToElement())

	return itemInv.BigInt(res)
}

func toParam(res ads.SumRes1d) ContractDBVerifySumParam {
	sum := res.Sum.BigInt(new(big.Int))
	a0 := res.SumProof.A0.BigInt(new(big.Int))
	a1 := res.SumProof.A1.BigInt(new(big.Int))
	a0inv := new(fr.Element).Inverse(&res.SumProof.A0).BigInt(new(big.Int))
	w1, w2 := toG1Point(res.SumProof.W1), toG1Point(res.SumProof.W2)
	fr := toG1Point(res.FR)

	var itemE []ContractDBDicItemE
	var itemR []ContractDBDicItemR
	var itemInv []*big.Int
	var itemWit []ContractDBG2Point

	for _, it := range res.DicItems {
		switch it.Value.(type) {
		case bn.G1Affine:
			itemE = append(itemE, toDicItemE(it))
		case bn.G2Affine:
			itemR = append(itemR, toDicItemR(it))
		default:
			panic("invalid dic items")
		}
		itemInv = append(itemInv, toItemInv(it))
		itemWit = append(itemWit, toG2Point(it.W))
	}

	var iset, igcd []ContractDBG1Point
	var iwit []ContractDBG2Point
	if len(res.IProof.DigestGCD) == 0 {
		iset = append(iset, fr)
	} else {
		for _, i := range res.IProof.DigestSet {
			iset = append(iset, toG1Point(i))
		}
		for _, i := range res.IProof.DigestGCD {
			igcd = append(igcd, toG1Point(i))
		}
		for _, i := range res.IProof.Witness {
			iwit = append(iwit, toG2Point(i))
		}
	}

	param := ContractDBVerifySumParam{
		Sum:     sum,
		A0:      a0,
		A1:      a1,
		A0inv:   a0inv,
		FR:      fr,
		W1:      w1,
		W2:      w2,
		ItemE:   itemE,
		ItemR:   itemR,
		ItemInv: itemInv,
		ItemWit: itemWit,
		Iset:    iset,
		Igcd:    igcd,
		Iwit:    iwit,
	}
	return param
}

func testGas(client *ethclient.Client, instance *Contract, addr string, q ContractDBQuery, param ContractDBVerifySumParam) {
	// test gas
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}
	privateKey, err := crypto.HexToECDSA(addr)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	auth, _ := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Failed to get suggested gas price: %v", err)
	}

	// invoke TestGas0, TestGas1, TestGas2
	tx, err := instance.TestGas0(auth)
	if err != nil {
		log.Fatalf("Failed to call TestGas0: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction receipt: %v", err)
	}
	fmt.Printf("Transaction mined: %v, gas: %v, blobgas: %v\n", receipt.TxHash, receipt.GasUsed, receipt.BlobGasUsed)

	tx, err = instance.TestGas1(auth, q, param)
	if err != nil {
		log.Fatalf("Failed to call TestGas0: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	receipt, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction receipt: %v", err)
	}
	fmt.Printf("Transaction mined: %v, gas: %v, blobgas: %v\n", receipt.TxHash, receipt.GasUsed, receipt.BlobGasUsed)

	tx, err = instance.TestGas2(auth, q, param)
	if err != nil {
		log.Fatalf("Failed to call TestGas0: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	receipt, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction receipt: %v", err)
	}
	fmt.Printf("Transaction mined: %v, gas: %v, blobgas: %v\n", receipt.TxHash, receipt.GasUsed, receipt.BlobGasUsed)

	cf := instance.ContractFilterer
	opt := &bind.FilterOpts{Start: receipt.BlockNumber.Uint64(), End: nil, Context: nil}
	it, _ := cf.FilterTestGas2Result(opt)
	for it.Next() {
		event := it.Event
		fmt.Printf("event: success: %v, result: %v\n", event.Success, event.Result)
	}
}

func TestGetDigests(t *testing.T) {
	digests := []bn.G1Affine{}
	conds := []string{"name", "bank", "addr", "value", "rate", "grade"}
	ind := ads.IndexInfo{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	for i := range conds {
		ind.Cond = conds[i : i+1]
		authDB, _ := leveldb.OpenFile(filepath.Join(acc.BaseDir, "authdb", ind.AuthTable()), nil)
		inf := acc.INF.Marshal()
		val, _ := authDB.Get(inf, nil)
		authDB.Close()
		digests = append(digests, acc.StringToG1Affine(string(val)))
	}
	for i := range digests {
		fmt.Printf("digests[%v].X=uint256(%v);\n", i, digests[i].X.String())
		fmt.Printf("digests[%v].Y=uint256(%v);\n", i, digests[i].Y.String())
	}
}
