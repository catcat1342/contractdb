package contractMulti

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
const contractAddr = "0x8a92d55d9EE30087d6b6d4BbDF8880695BA2BD9b"

func TestContractMulti(t *testing.T) {
	ind := 4
	res, q := genInputMulti(ind)
	param := toParamMulti(res)

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
	// result, err := instance.VerifyQueryMulti(&bind.CallOpts{Context: context.Background()}, q, param)
	// if err != nil {
	// 	log.Fatalf("Failed to call VerifyQueryMulti: %v", err)
	// }
	// fmt.Printf("DEBUG: result: %v, flag: %v\n", result.Ver, result.Flag)

	// test gas cost
	addr := accountAddr
	testGasMulti(client, instance, addr, q, param)
}

func genInputMulti(ind int) (ads.SumRes1d, ContractDBMultiQueryMulti) {
	MAXN := 1 << 12
	pk1, pk2 := acc.LoadPubkey(MAXN)
	q := ads.Query{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	switch ind {
	case 0:
		q.Cond = []string{"name", "value"}
		q.CondFlag = []int{0, 1}
		q.CondVal = []interface{}{"Name3948", []uint64{4000, 5000}}
	case 1:
		q.Cond = []string{"bank", "name"}
		q.CondFlag = []int{0, 0}
		q.CondVal = []interface{}{"Bank564", "Name3948"}
	case 2:
		q.Cond = []string{"bank", "name", "value"}
		q.CondFlag = []int{0, 0, 1}
		q.CondVal = []interface{}{"Bank564", "Name3948", []uint64{4000, 5000}}
	case 3:
		q.Cond = []string{"addr", "bank", "name"}
		q.CondFlag = []int{0, 0, 0}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948"}
	case 4:
		q.Cond = []string{"addr", "bank", "name", "value"}
		q.CondFlag = []int{0, 0, 0, 1}
		q.CondVal = []interface{}{"Addr221", "Bank564", "Name3948", []uint64{4000, 5000}}
	default:
		panic("invalid input type for multi_query")
	}

	res, err := ads.QueryDBMulti(ads.DBINFO, q, pk1, pk2)
	if err != nil {
		log.Printf("query db failed: %v\n", err)
		return ads.SumRes1d{}, ContractDBMultiQueryMulti{}
	}
	verres, err := ads.VerifyMulti(ads.DBINFO, q, res, pk1, pk2)
	if err != nil {
		log.Printf("verify failed: %v\n", err)
		return ads.SumRes1d{}, ContractDBMultiQueryMulti{}
	}
	sum := res.(ads.SumRes1d).Sum
	fmt.Printf("result: %v, verres: %v\n", sum.String(), verres)
	return res.(ads.SumRes1d), toContractQueryMulti(q, ind)
}

func toContractQueryMulti(q ads.Query, ind int) ContractDBMultiQueryMulti {
	// ind: 11, 20, 21, 30, 21

	rtype := []uint8{} // indicating type (1,2,3,4) of rind
	eval := []uint64{}
	rval := []uint64{}

	for i, f := range q.CondFlag {
		switch f {
		case 0:
			eval = append(eval, new(fr.Element).SetBytes([]byte(q.CondVal[i].(string))).Uint64())
		case 1, 2, 3, 4:
			rtype = append(rtype, uint8(f))
			rval = append(rval, q.CondVal[i].([]uint64)[0], q.CondVal[i].([]uint64)[1])
		default:
			panic("invalid query")
		}
	}
	return ContractDBMultiQueryMulti{
		Index: uint8(ind),
		Rtype: rtype,
		Eval:  eval,
		Rval:  rval,
	}
}

func toParamMulti(res ads.SumRes1d) ContractDBMultiVerifySumParam {

	sum := res.Sum.BigInt(new(big.Int))
	a0 := res.SumProof.A0.BigInt(new(big.Int))
	a1 := res.SumProof.A1.BigInt(new(big.Int))
	a0inv := new(fr.Element).Inverse(&res.SumProof.A0).BigInt(new(big.Int))
	w1, w2 := toG1Point(res.SumProof.W1), toG1Point(res.SumProof.W2)
	fr := toG1Point(res.FR)

	var itemE []ContractDBMultiDicItemE
	var itemR []ContractDBMultiDicItemR
	var itemInv []*big.Int
	var itemWit []ContractDBMultiG2Point

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

	// multi query param has no intersection proof
	param := ContractDBMultiVerifySumParam{
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
	}
	return param
}

func testGasMulti(client *ethclient.Client, instance *Contract, addr string, q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) {
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
	tx, err := instance.TestGas0Multi(auth)
	if err != nil {
		log.Fatalf("Failed to call TestGas0: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction receipt: %v", err)
	}
	fmt.Printf("Transaction mined: %v, gas: %v, blobgas: %v\n", receipt.TxHash, receipt.GasUsed, receipt.BlobGasUsed)

	tx, err = instance.TestGas1Multi(auth, q, param)
	if err != nil {
		log.Fatalf("Failed to call TestGas0: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	receipt, err = bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Failed to wait for transaction receipt: %v", err)
	}
	fmt.Printf("Transaction mined: %v, gas: %v, blobgas: %v\n", receipt.TxHash, receipt.GasUsed, receipt.BlobGasUsed)

	tx, err = instance.TestGas2Multi(auth, q, param)
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

func TestGetDigestsMulti(t *testing.T) {
	digests := []bn.G1Affine{}
	conds := [][]string{
		{"name", "value"},
		{"bank", "name"},
		{"bank", "name", "value"},
		{"addr", "bank", "name"},
		{"addr", "bank", "name", "value"},
	}
	ind := ads.IndexInfo{
		Table:    "test10",
		Dest:     "result", // add keyid at first, use fixed sequence
		DestType: "sum",
	}
	for i := range conds {
		ind.Cond = conds[i]
		authDB, _ := leveldb.OpenFile(filepath.Join(acc.BaseDir, "authdb", ind.AuthTable()), nil)
		inf := acc.INF.Marshal()
		val, _ := authDB.Get(inf, nil)
		authDB.Close()
		digests = append(digests, acc.StringToG1Affine(string(val)))
	}
	for i := range digests {
		fmt.Printf("digestsMulti[%v].X=uint256(%v);\n", i, digests[i].X.String())
		fmt.Printf("digestsMulti[%v].Y=uint256(%v);\n", i, digests[i].Y.String())
	}
}

func toG1Point(p bn.G1Affine) ContractDBMultiG1Point {
	return ContractDBMultiG1Point{
		X: p.X.BigInt(new(big.Int)),
		Y: p.Y.BigInt(new(big.Int)),
	}
}

func toG2Point(p bn.G2Affine) ContractDBMultiG2Point {
	c0 := p.X.A0.BigInt(new(big.Int))
	c1 := p.X.A1.BigInt(new(big.Int))
	c2 := p.Y.A0.BigInt(new(big.Int))
	c3 := p.Y.A1.BigInt(new(big.Int))
	return ContractDBMultiG2Point{
		X: [2]*big.Int{c1, c0}, // don't forget to exchange coeffs
		Y: [2]*big.Int{c3, c2},
	}
}

func toDicItemE(d acc.DicItem) ContractDBMultiDicItemE {
	key := d.Key.Uint64()
	next := d.Nxt.Uint64()
	val := toG1Point(d.Value.(bn.G1Affine))
	return ContractDBMultiDicItemE{
		Key:   key,
		Nxt:   next,
		Value: val,
	}
}

func toDicItemR(d acc.DicItem) ContractDBMultiDicItemR {
	key := d.Key.Uint64()
	next := d.Nxt.Uint64()
	val := toG2Point(d.Value.(bn.G2Affine))
	return ContractDBMultiDicItemR{
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
