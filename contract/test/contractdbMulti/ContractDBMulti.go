// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractMulti

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractDBMultiDicItemE is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiDicItemE struct {
	Key   uint64
	Nxt   uint64
	Value ContractDBMultiG1Point
}

// ContractDBMultiDicItemR is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiDicItemR struct {
	Key   uint64
	Nxt   uint64
	Value ContractDBMultiG2Point
}

// ContractDBMultiG1Point is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiG1Point struct {
	X *big.Int
	Y *big.Int
}

// ContractDBMultiG2Point is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiG2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

// ContractDBMultiQueryMulti is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiQueryMulti struct {
	Index uint8
	Rtype []uint8
	Eval  []uint64
	Rval  []uint64
}

// ContractDBMultiVerifySumParam is an auto generated low-level Go binding around an user-defined struct.
type ContractDBMultiVerifySumParam struct {
	Sum     *big.Int
	A0      *big.Int
	A1      *big.Int
	A0inv   *big.Int
	FR      ContractDBMultiG1Point
	W1      ContractDBMultiG1Point
	W2      ContractDBMultiG1Point
	ItemE   []ContractDBMultiDicItemE
	ItemR   []ContractDBMultiDicItemR
	ItemInv []*big.Int
	ItemWit []ContractDBMultiG2Point
	Iset    []ContractDBMultiG1Point
	Igcd    []ContractDBMultiG1Point
	Iwit    []ContractDBMultiG2Point
}

// ContractMetaData contains all meta data concerning the Contract contract.
var ContractMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"result\",\"type\":\"uint256\"}],\"name\":\"TestGas2Result\",\"type\":\"event\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"},{\"internalType\":\"uint8[]\",\"name\":\"rtype\",\"type\":\"uint8[]\"},{\"internalType\":\"uint64[]\",\"name\":\"eval\",\"type\":\"uint64[]\"},{\"internalType\":\"uint64[]\",\"name\":\"rval\",\"type\":\"uint64[]\"}],\"internalType\":\"structContractDBMulti.QueryMulti\",\"name\":\"q\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0inv\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"fR\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w1\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w2\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemE[]\",\"name\":\"itemE\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemR[]\",\"name\":\"itemR\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"itemInv\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"itemWit\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"iset\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"igcd\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"iwit\",\"type\":\"tuple[]\"}],\"internalType\":\"structContractDBMulti.VerifySumParam\",\"name\":\"param\",\"type\":\"tuple\"}],\"name\":\"VerifyQueryMulti\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"ver\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"flag\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"TestGas0Multi\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"},{\"internalType\":\"uint8[]\",\"name\":\"rtype\",\"type\":\"uint8[]\"},{\"internalType\":\"uint64[]\",\"name\":\"eval\",\"type\":\"uint64[]\"},{\"internalType\":\"uint64[]\",\"name\":\"rval\",\"type\":\"uint64[]\"}],\"internalType\":\"structContractDBMulti.QueryMulti\",\"name\":\"q\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0inv\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"fR\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w1\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w2\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemE[]\",\"name\":\"itemE\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemR[]\",\"name\":\"itemR\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"itemInv\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"itemWit\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"iset\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"igcd\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"iwit\",\"type\":\"tuple[]\"}],\"internalType\":\"structContractDBMulti.VerifySumParam\",\"name\":\"param\",\"type\":\"tuple\"}],\"name\":\"TestGas1Multi\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"},{\"internalType\":\"uint8[]\",\"name\":\"rtype\",\"type\":\"uint8[]\"},{\"internalType\":\"uint64[]\",\"name\":\"eval\",\"type\":\"uint64[]\"},{\"internalType\":\"uint64[]\",\"name\":\"rval\",\"type\":\"uint64[]\"}],\"internalType\":\"structContractDBMulti.QueryMulti\",\"name\":\"q\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"sum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a1\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"a0inv\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"fR\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w1\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"w2\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemE[]\",\"name\":\"itemE\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint64\",\"name\":\"key\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"nxt\",\"type\":\"uint64\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point\",\"name\":\"value\",\"type\":\"tuple\"}],\"internalType\":\"structContractDBMulti.DicItemR[]\",\"name\":\"itemR\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"itemInv\",\"type\":\"uint256[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"itemWit\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"iset\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"X\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"Y\",\"type\":\"uint256\"}],\"internalType\":\"structContractDBMulti.G1Point[]\",\"name\":\"igcd\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"uint256[2]\",\"name\":\"X\",\"type\":\"uint256[2]\"},{\"internalType\":\"uint256[2]\",\"name\":\"Y\",\"type\":\"uint256[2]\"}],\"internalType\":\"structContractDBMulti.G2Point[]\",\"name\":\"iwit\",\"type\":\"tuple[]\"}],\"internalType\":\"structContractDBMulti.VerifySumParam\",\"name\":\"param\",\"type\":\"tuple\"}],\"name\":\"TestGas2Multi\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"ver\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"flag\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ContractABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractMetaData.ABI instead.
var ContractABI = ContractMetaData.ABI

// Contract is an auto generated Go binding around an Ethereum contract.
type Contract struct {
	ContractCaller     // Read-only binding to the contract
	ContractTransactor // Write-only binding to the contract
	ContractFilterer   // Log filterer for contract events
}

// ContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractSession struct {
	Contract     *Contract         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractCallerSession struct {
	Contract *ContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractTransactorSession struct {
	Contract     *ContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractRaw struct {
	Contract *Contract // Generic contract binding to access the raw methods on
}

// ContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractCallerRaw struct {
	Contract *ContractCaller // Generic read-only contract binding to access the raw methods on
}

// ContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractTransactorRaw struct {
	Contract *ContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContract creates a new instance of Contract, bound to a specific deployed contract.
func NewContract(address common.Address, backend bind.ContractBackend) (*Contract, error) {
	contract, err := bindContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contract{ContractCaller: ContractCaller{contract: contract}, ContractTransactor: ContractTransactor{contract: contract}, ContractFilterer: ContractFilterer{contract: contract}}, nil
}

// NewContractCaller creates a new read-only instance of Contract, bound to a specific deployed contract.
func NewContractCaller(address common.Address, caller bind.ContractCaller) (*ContractCaller, error) {
	contract, err := bindContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractCaller{contract: contract}, nil
}

// NewContractTransactor creates a new write-only instance of Contract, bound to a specific deployed contract.
func NewContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractTransactor, error) {
	contract, err := bindContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractTransactor{contract: contract}, nil
}

// NewContractFilterer creates a new log filterer instance of Contract, bound to a specific deployed contract.
func NewContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractFilterer, error) {
	contract, err := bindContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractFilterer{contract: contract}, nil
}

// bindContract binds a generic wrapper to an already deployed contract.
func bindContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.ContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.ContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contract *ContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contract *ContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contract *ContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contract.Contract.contract.Transact(opts, method, params...)
}

// VerifyQueryMulti is a free data retrieval call binding the contract method 0x63b9b528.
//
// Solidity: function VerifyQueryMulti((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) view returns(bool ver, uint256 flag)
func (_Contract *ContractCaller) VerifyQueryMulti(opts *bind.CallOpts, q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (struct {
	Ver  bool
	Flag *big.Int
}, error) {
	var out []interface{}
	err := _Contract.contract.Call(opts, &out, "VerifyQueryMulti", q, param)

	outstruct := new(struct {
		Ver  bool
		Flag *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Ver = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Flag = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// VerifyQueryMulti is a free data retrieval call binding the contract method 0x63b9b528.
//
// Solidity: function VerifyQueryMulti((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) view returns(bool ver, uint256 flag)
func (_Contract *ContractSession) VerifyQueryMulti(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (struct {
	Ver  bool
	Flag *big.Int
}, error) {
	return _Contract.Contract.VerifyQueryMulti(&_Contract.CallOpts, q, param)
}

// VerifyQueryMulti is a free data retrieval call binding the contract method 0x63b9b528.
//
// Solidity: function VerifyQueryMulti((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) view returns(bool ver, uint256 flag)
func (_Contract *ContractCallerSession) VerifyQueryMulti(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (struct {
	Ver  bool
	Flag *big.Int
}, error) {
	return _Contract.Contract.VerifyQueryMulti(&_Contract.CallOpts, q, param)
}

// TestGas0Multi is a paid mutator transaction binding the contract method 0x0c3ad2d8.
//
// Solidity: function TestGas0Multi() returns(bool, uint256)
func (_Contract *ContractTransactor) TestGas0Multi(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "TestGas0Multi")
}

// TestGas0Multi is a paid mutator transaction binding the contract method 0x0c3ad2d8.
//
// Solidity: function TestGas0Multi() returns(bool, uint256)
func (_Contract *ContractSession) TestGas0Multi() (*types.Transaction, error) {
	return _Contract.Contract.TestGas0Multi(&_Contract.TransactOpts)
}

// TestGas0Multi is a paid mutator transaction binding the contract method 0x0c3ad2d8.
//
// Solidity: function TestGas0Multi() returns(bool, uint256)
func (_Contract *ContractTransactorSession) TestGas0Multi() (*types.Transaction, error) {
	return _Contract.Contract.TestGas0Multi(&_Contract.TransactOpts)
}

// TestGas1Multi is a paid mutator transaction binding the contract method 0xc397d19d.
//
// Solidity: function TestGas1Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool, uint256)
func (_Contract *ContractTransactor) TestGas1Multi(opts *bind.TransactOpts, q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "TestGas1Multi", q, param)
}

// TestGas1Multi is a paid mutator transaction binding the contract method 0xc397d19d.
//
// Solidity: function TestGas1Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool, uint256)
func (_Contract *ContractSession) TestGas1Multi(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.Contract.TestGas1Multi(&_Contract.TransactOpts, q, param)
}

// TestGas1Multi is a paid mutator transaction binding the contract method 0xc397d19d.
//
// Solidity: function TestGas1Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool, uint256)
func (_Contract *ContractTransactorSession) TestGas1Multi(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.Contract.TestGas1Multi(&_Contract.TransactOpts, q, param)
}

// TestGas2Multi is a paid mutator transaction binding the contract method 0xb43b0b0c.
//
// Solidity: function TestGas2Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool ver, uint256 flag)
func (_Contract *ContractTransactor) TestGas2Multi(opts *bind.TransactOpts, q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.contract.Transact(opts, "TestGas2Multi", q, param)
}

// TestGas2Multi is a paid mutator transaction binding the contract method 0xb43b0b0c.
//
// Solidity: function TestGas2Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool ver, uint256 flag)
func (_Contract *ContractSession) TestGas2Multi(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.Contract.TestGas2Multi(&_Contract.TransactOpts, q, param)
}

// TestGas2Multi is a paid mutator transaction binding the contract method 0xb43b0b0c.
//
// Solidity: function TestGas2Multi((uint8,uint8[],uint64[],uint64[]) q, (uint256,uint256,uint256,uint256,(uint256,uint256),(uint256,uint256),(uint256,uint256),(uint64,uint64,(uint256,uint256))[],(uint64,uint64,(uint256[2],uint256[2]))[],uint256[],(uint256[2],uint256[2])[],(uint256,uint256)[],(uint256,uint256)[],(uint256[2],uint256[2])[]) param) returns(bool ver, uint256 flag)
func (_Contract *ContractTransactorSession) TestGas2Multi(q ContractDBMultiQueryMulti, param ContractDBMultiVerifySumParam) (*types.Transaction, error) {
	return _Contract.Contract.TestGas2Multi(&_Contract.TransactOpts, q, param)
}

// ContractTestGas2ResultIterator is returned from FilterTestGas2Result and is used to iterate over the raw logs and unpacked data for TestGas2Result events raised by the Contract contract.
type ContractTestGas2ResultIterator struct {
	Event *ContractTestGas2Result // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractTestGas2ResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractTestGas2Result)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractTestGas2Result)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractTestGas2ResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractTestGas2ResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractTestGas2Result represents a TestGas2Result event raised by the Contract contract.
type ContractTestGas2Result struct {
	Success bool
	Result  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTestGas2Result is a free log retrieval operation binding the contract event 0x1662b5f33b72b4a2ea7aa48e44c56a648f7ab0056638a88df3c4101479d02bad.
//
// Solidity: event TestGas2Result(bool success, uint256 result)
func (_Contract *ContractFilterer) FilterTestGas2Result(opts *bind.FilterOpts) (*ContractTestGas2ResultIterator, error) {

	logs, sub, err := _Contract.contract.FilterLogs(opts, "TestGas2Result")
	if err != nil {
		return nil, err
	}
	return &ContractTestGas2ResultIterator{contract: _Contract.contract, event: "TestGas2Result", logs: logs, sub: sub}, nil
}

// WatchTestGas2Result is a free log subscription operation binding the contract event 0x1662b5f33b72b4a2ea7aa48e44c56a648f7ab0056638a88df3c4101479d02bad.
//
// Solidity: event TestGas2Result(bool success, uint256 result)
func (_Contract *ContractFilterer) WatchTestGas2Result(opts *bind.WatchOpts, sink chan<- *ContractTestGas2Result) (event.Subscription, error) {

	logs, sub, err := _Contract.contract.WatchLogs(opts, "TestGas2Result")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractTestGas2Result)
				if err := _Contract.contract.UnpackLog(event, "TestGas2Result", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTestGas2Result is a log parse operation binding the contract event 0x1662b5f33b72b4a2ea7aa48e44c56a648f7ab0056638a88df3c4101479d02bad.
//
// Solidity: event TestGas2Result(bool success, uint256 result)
func (_Contract *ContractFilterer) ParseTestGas2Result(log types.Log) (*ContractTestGas2Result, error) {
	event := new(ContractTestGas2Result)
	if err := _Contract.contract.UnpackLog(event, "TestGas2Result", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
