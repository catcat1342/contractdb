# ContractDB: Enabling Secure and Efficient Database Accessing in Smart Contracts

## I. Introduction

This is a prototype implementation of ContractDB, which is used to check the functionalities of the ContractDB design and evaluate its performance. We did not perform extensive engineering optimization or vulnerability checks on our implementation; therefore, it cannot be used for practical applications. However, by referring to this document, one can easily understand the workflow of ContractDB and find its approximate running costs for both the server side and the contract side.

### Summary of ContractDB

In brief, ContractDB aims to solve two problems faced by smart contracts:

**Problem 1: Data storage is expensive.**
In Ethereum, storing 32-byte data in storage mode costs 20,000 gas. This makes it expensive, and often infeasible, to deploy applications with large datasets on Ethereum. In particular, many applications are expected to migrate to the blockchain rather than building brand-new alternatives. This implies the need to handle massive legacy data. Clearly, storing such data directly in smart contracts is impractical.

**Problem 2: Data query is not well-supported.**
Ethereum only supports key-value queries (using the map data structure). However, practical applications often require complex conditional queries, such as multi-dimensional equivalence and range queries. Furthermore, the demand for querying indicates that the data actually needed is far less than the full dataset. This further demonstrates that storing all data in contracts is inappropriate.

**Our solution:**

**Step 1: Verifiable Database (VDB) & Smart Contract.** We combine VDB with smart contracts, allowing us to store data in an external database and provide the required data for each invocation along with the invocation input. The smart contract can verify the provided data based on VDB technology.

**Step 2: VDB with Constant Proof Size.** Currently, existing VDB schemes require O(log N) proof size, where N is the number of data items (or the line number of the database table).

## II. Implementation framework

ContractDB consists of two parts:

- A verifiable database implemented in Go.
- A smart contract implemented in Solidity.

The verifiable database (VDB) supports verifiable queries and updates on a MySQL database. Our design of the VDB relies on two technologies: accumulator-based verifiable set operations and accumulator-based verifiable dictionaries, both of which utilize bilinear pairing (BP) provided by gnark-crypto/bn254. Thus, the implementation of the VDB consists of the following packages:

- **accumulator**: Implements accumulator-based verifiable set operations and dictionaries.
- **ads**: Implements authenticated data structures (ADS) on a MySQL database, supporting ADS construction, queries with proof, and query result verification. [To be done: verifiable update]

In addition, the **dataset** package is used to generate datasets, and the **performance** package provides a set of evaluation functions. The **contract** folder implements smart contracts that verify the query results and provides interfaces to invoke the contracts.

## III. Evaluate the VDB

The main function provides the following functionalities related to the VDB. Before running these functions, ensure that the following settings are correctly configured:

- Run the MySQL server and modify the `dbinfo` defined in `contractdb/ads/auth.go`:
  `const DBINFO = "ubuntu:ubuntu@tcp(localhost:3306)/contractdb"`
- Modify the `basedir` defined in `contractdb/accumulator/pubkey.go`:
  `const BaseDir = "/home/ubuntu/contractdb"`

Then, run the following functions one by one:

1. **InitPubkey()**: Generates 10,000,000 public keys and stores them in `authdb/pubkey`.

2. **GenTestDataAll()**: Generates all datasets (N=2^15, 2^16, ..., 2^20) used for performance evaluation.

3. **CreateIndexTest()**: Evaluates index creation performance for each N. This function may take several hours to create all indexes for N ranging from 2^15 to 2^20.

4. **RuntimeQueryWithIntersection()**: Evaluates query performance on N=2^20 with different query conditions.

5. **CreateMultiIndexTest()**: Evaluates multi-condition index creation performance for each N. TThis function may take several hours to create all indexes for N ranging from 2^15 to 2^20.

6. **RuntimeQueryWithMultiCond()**: Evaluates query runtime with multi-condition indexes on N=2^20 with various query conditions.

## IV. Evaluate the contract

The `contractdb/contract` directory implements the contract to verify query results. The required steps are as follows:

### 1. Run Ganache Testnet

Use an independent terminal to start an Ethereum testnet based on Ganache CLI:
`ganache-cli -i 1337`. Once the testnet is started, it outputs a list of accounts. Record one of them and copy it to the following files:

- `contractdb/contract/test/contractdb/contract_test.go`
- `contractdb/contract/test/contractdbMulti/contract_multi_test.go`

For example:

```go
// contractdb/contract/test/contractdb/contract_test.go
const accountAddr = "fa2a7eeeac24d706f4b925430f9e8025ef5fd0c6c9fc5ca6e3e48f6dbb71ebed"

// contractdb/contract/test/contractdbMulti/contract_multi_test.go
const accountAddr = "fa2a7eeeac24d706f4b925430f9e8025ef5fd0c6c9fc5ca6e3e48f6dbb71ebed"
```

### 2. deploy contracts

Navigate to the contract directory:
cd contractdb/contract, then run `truffle migrate --reset` to compile and deploy the contracts.

The deployment will output three contract addresses. Record the addresses of ContractDB and ContractDBMulti in the following files:

contractdb/contract/test/contractdb/contract_test.go
contractdb/contract/test/contractdbMulti/contract_multi_test.go

For example:

```go
// contractdb/contract/test/contractdb/contract_test.go
const contractAddr = "0x485902042071B0238F73CbA55681faFdC8103b74"

// contractdb/contract/test/contractdbMulti/contract_multi_test.go
const contractAddr = "0x8a92d55d9EE30087d6b6d4BbDF8880695BA2BD9b"
```

### 3. Invoke Contracts and Evaluate Gas Cost

Navigate to the directory:
`cd test/contractdb`, then run `go test -v -run ^TestContract$`.

Alternatively, you can navigate to:
`cd test/contractdbMulti` and run `go test -v -run ^TestContractMulti$`.

The `TestContract` function sends three invocations that separately call `TestGas0`, `TestGas1`, and `TestGas2` in the contract.

- `TestGas0` evaluates the basic gas cost of modifying a storage slot in the contract.
- `TestGas1` evaluates the gas cost of sending the query results and proofs.
- `TestGas2` evaluates the total gas cost of the verification.
