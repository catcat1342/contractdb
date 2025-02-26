// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

contract ContractDBMulti {
    uint testgas; // used by TestGas functions
    uint256 internal constant PRIME_P = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
    uint256 internal constant INF = 115792089237316195423570985008687907853269984665640564039457584007913129639935; // used for modAdd

    uint256 constant PRIME_Q = 21888242871839275222246405745257275088696311157297823662689037894645226208583;

    struct G1Point {
        uint256 X;
        uint256 Y;
    }
    // Encoding of field elements is: X[0] * z + X[1]
    struct G2Point {
        uint256[2] X;
        uint256[2] Y;
    }

    struct QueryMulti {
        uint8 index; // 0: (name,value); 1: (name,bank); 2: (name,bank,value); 3: (name,bank,addr); 4: (name,bank,addr,value)
        uint8[] rtype; // 1 [vr,vr]; 2 (vl,vr]; 3 [vl,vr); 4 (vl,vr)
        uint64[] eval;
        uint64[] rval;
    }

    struct VerifySumParam {
        // SumProof sump;
        uint256 sum;
        uint256 a0;
        uint256 a1;
        uint256 a0inv; // for saving gas, input the inv
        G1Point fR;
        G1Point w1;
        G1Point w2;
        // DicItems dic;
        DicItemE[] itemE;
        DicItemR[] itemR;
        uint256[] itemInv;
        G2Point[] itemWit;
        // IProof iProof;
        G1Point[] iset; // G1Point
        G1Point[] igcd; // G1Point
        G2Point[] iwit; // G2Point
    }

    struct DicItemE {
        uint64 key;
        uint64 nxt;
        G1Point value;
    }

    struct DicItemR {
        uint64 key;
        uint64 nxt;
        G2Point value;
    }

    // struct DicItems {
    //     DicItemE[] itemE;
    //     DicItemR[] itemR;
    //     uint256[] itemInv;
    //     G2Point[] itemWit;
    // }

    // struct IProof {
    //     G1Point[] iset; // G1Point
    //     G1Point[] igcd; // G1Point
    //     G2Point[] iwit; // G2Point
    // }

    // struct SumProof {
    //     uint256 sum;s
    //     uint256 a0;
    //     uint256 a1;
    //     uint256 a0inv; // for saving gas, input the inv
    //     G1Point fR;
    //     G1Point w1;
    //     G1Point w2;
    // }

    G1Point internal G1 = G1Point(1, 2);
    G1Point internal sG1 =
        G1Point(
            659387509072190054499872531157327851974123307391705746989088605302742528024,
            20230846700664440841635347868950212739976037558458206556816116785749819631898
        );
    G2Point internal G2 =
        G2Point(
            [
                11559732032986387107991004021392285783925812861821192530917403151452391805634,
                10857046999023057135944570762232829481370756359578518086990519993285655852781
            ],
            [
                4082367875863433681332203403145435568316851327593401208105741076214120093531,
                8495653923123431417604973247489272438418190587263600148770280649306958101930
            ]
        );
    G2Point internal sG2 =
        G2Point(
            [
                13036583401943305728528725764547309024003329541956693774318040152348747598121,
                12276881524974618347963204355819156766680811359308712062143904989647769252018
            ],
            [
                2188726141281701348289276224237737751177036697763916293382809963867816883720,
                3596309672727352412249421695858998934225755266543379229855517037447563584325
            ]
        );

    G1Point[5] internal digestsMulti; // 0: (name,value); 1: (bank,name); 2: (bank,name,value); 3: (addr,bank,name); 4: (addr,bank,name,value)

    constructor() {
        digestsMulti[0].X = uint256(6159458686690713616529471692641016634900583717446801627184495732237963694026);
        digestsMulti[0].Y = uint256(12860754021025649890692309501072139929328081008835158233278232810146889880194);
        digestsMulti[1].X = uint256(18773105457264462803422938463021834716946221440815805151096662954412506849643);
        digestsMulti[1].Y = uint256(16010421961783507917303007679028708137532917472658384219313664733923555551147);
        digestsMulti[2].X = uint256(20246459698207796486935212416125072967016708693585516928690612225012059813321);
        digestsMulti[2].Y = uint256(7292323691925995207746639282184672086984205183190602449222138715208653943573);
        digestsMulti[3].X = uint256(329738446408575427612569506018095409541081660617558863716924695571200692040);
        digestsMulti[3].Y = uint256(20271886177555646240295359285434577499960046312814718633564395573614489259760);
        digestsMulti[4].X = uint256(20628696334436384818609201711793048828968667631681356746880570203883871994481);
        digestsMulti[4].Y = uint256(13805588535609973374545965874932086042134248668445161539799168790419013966574);
    }

    function negate(G1Point memory p) internal pure returns (G1Point memory) {
        // The prime q in the base field F_q for G1
        if (p.X == 0 && p.Y == 0) {
            return G1Point(0, 0);
        } else {
            return G1Point(p.X, PRIME_Q - (p.Y % PRIME_Q));
        }
    }

    /*
     * @return r the product of a point on G1 and a scalar, i.e.
     *         p == p.scalar_mul(1) and p.plus(p) == p.scalar_mul(2) for all
     *         points p.
     */
    function scalar_mul(G1Point memory p, uint256 s) internal view returns (G1Point memory r) {
        uint256[3] memory input;
        input[0] = p.X;
        input[1] = p.Y;
        input[2] = s;
        bool success;
        // solium-disable-next-line security/no-inline-assembly
        assembly {
            success := staticcall(sub(gas(), 2000), 7, input, 0x80, r, 0x60)
        }
    }

    function plus(G1Point memory p1, G1Point memory p2) internal view returns (G1Point memory r) {
        uint256[4] memory input;
        //uint256[] memory input = new uint256[](4);
        input[0] = p1.X;
        input[1] = p1.Y;
        input[2] = p2.X;
        input[3] = p2.Y;
        bool success;
        // solium-disable-next-line security/no-inline-assembly
        assembly {
            success := staticcall(sub(gas(), 2000), 6, input, 0xc0, r, 0x60)
        }
    }

    function pairingMulti(G1Point[] memory p1, G2Point[] memory p2) internal view returns (bool) {
        uint inputSize = p1.length * 6;
        uint256[] memory input = new uint256[](inputSize);

        for (uint256 i = 0; i < p1.length; i++) {
            uint256 j = i * 6;
            input[j + 0] = p1[i].X;
            input[j + 1] = p1[i].Y;
            input[j + 2] = p2[i].X[0];
            input[j + 3] = p2[i].X[1];
            input[j + 4] = p2[i].Y[0];
            input[j + 5] = p2[i].Y[1];
        }

        uint256[1] memory out;
        bool success;

        // solium-disable-next-line security/no-inline-assembly
        assembly {
            success := staticcall(sub(gas(), 2000), 8, add(input, 0x20), mul(inputSize, 0x20), out, 0x20)
            // Use "invalid" to make gas estimation work
        }
        return out[0] != 0;
    }

    // ensure that a<PRIME_P, b<PRIME_P, omit check for saving gas
    // function modAdd(uint256 a, uint256 b) internal pure returns (uint256 r) {
    //     assembly {
    //         r := addmod(a, b, PRIME_P)
    //     }
    // }
    // a*b % PRIME_P
    function modMul(uint256 a, uint256 b) internal pure returns (uint256 r) {
        assembly {
            r := mulmod(a, b, PRIME_P)
        }
    }
    function isInv(uint256 a, uint256 b) internal pure returns (bool) {
        uint256 r;
        assembly {
            r := mulmod(a, b, PRIME_P)
        }
        return (r == 1);
    }

    function checkItemEInv(DicItemE calldata item, uint256 inv) internal pure returns (bool) {
        uint256 itemUint = uint256(keccak256(abi.encodePacked(item.key, item.nxt, item.value.X, item.value.Y))) % PRIME_P;
        return modMul(itemUint, inv) == 1;
    }

    function checkItemInv(DicItemR calldata item, uint256 inv) internal pure returns (bool) {
        uint256 itemUint = uint256(
            keccak256(abi.encodePacked(item.key, item.nxt, item.value.X[1], item.value.X[0], item.value.Y[1], item.value.Y[0]))
        ) % PRIME_P;
        return modMul(itemUint, inv) == 1;
    }

    function VerifyQueryMulti(QueryMulti calldata q, VerifySumParam calldata param) public view returns (bool ver, uint flag) {
        uint offset;
        uint layer; // layer number upper the bottum layer
        bool res;
        G1Point[] memory p1;
        G2Point[] memory p2;

        // append all pairs in p1 and p2
        // sum requires 4 pairs, each itemE requires 2 pairs, each itemR requires 6 pairs
        if (q.index == 0) {
            // (name,value)
            p1 = new G1Point[](12);
            p2 = new G2Point[](12);
        } else if (q.index == 1) {
            // (bank,name)
            p1 = new G1Point[](8);
            p2 = new G2Point[](8);
        } else if (q.index == 2) {
            // (bank,name,value)
            p1 = new G1Point[](14);
            p2 = new G2Point[](14);
        } else if (q.index == 3) {
            // (addr,bank,name)
            p1 = new G1Point[](10);
            p2 = new G2Point[](10);
        } else if (q.index == 4) {
            // (addr,bank,name,value)
            p1 = new G1Point[](16);
            p2 = new G2Point[](16);
        } else {
            return (false, 4444);
        }

        // append sum pairings to p1 and p2
        if (param.fR.X == 1 && param.fR.Y == 2) return (param.sum == 0, 111);
        if (modMul(param.a0, param.a0inv) != 1) return (false, 222);
        if (modMul(param.a0inv, param.a1) != param.sum) return (false, 333);
        p1[0] = negate(param.w1);
        p1[1] = plus(param.fR, negate(scalar_mul(G1, param.a0)));
        p1[2] = negate(param.w2);
        p1[3] = plus(param.w1, negate(scalar_mul(G1, param.a1)));
        p2[0] = sG2;
        p2[1] = G2;
        p2[2] = sG2;
        p2[3] = G2;

        // check layer0 dic item and append it to (p1,p2)
        // 1. check queryCondition ?= dicItem.Key
        if (q.eval[0] != param.itemE[0].key) return (false, 1201);
        // 2. check inverse value
        if (checkItemEInv(param.itemE[0], param.itemInv[0]) == false) return (false, 14);
        // 3. append e(acc_item, sG1)==e(digest, wit) to (p1,p2)
        p1[4] = negate(plus(scalar_mul(G1, param.itemInv[0]), sG1));
        p1[5] = digestsMulti[q.index];
        p2[4] = param.itemWit[0];
        p2[5] = G2;

        // check layer1 in the middle if exists
        if (q.index == 2 || q.index == 3 || q.index == 4) {
            if (q.eval[1] != param.itemE[1].key) return (false, 1202);
            if (checkItemEInv(param.itemE[1], param.itemInv[1]) == false) return (false, 14);
            p1[6] = negate(plus(scalar_mul(G1, param.itemInv[1]), sG1));
            p1[7] = param.itemE[0].value;
            p2[6] = param.itemWit[1];
            p2[7] = G2;
        }

        // check layer2 in the middle if exists
        if (q.index == 4) {
            if (q.eval[2] != param.itemE[2].key) return (false, 1203);
            if (checkItemEInv(param.itemE[2], param.itemInv[2]) == false) return (false, 14);
            p1[8] = negate(plus(scalar_mul(G1, param.itemInv[2]), sG1));
            p1[9] = param.itemE[1].value;
            p2[8] = param.itemWit[2];
            p2[9] = G2;
        }

        // check bottom layer
        if (q.index == 0 || q.index == 1) {
            offset = 6;
            layer = 0;
        } else if (q.index == 2 || q.index == 3) {
            offset = 8;
            layer = 1;
        } else {
            offset = 10;
            layer = 2;
        }

        if (q.index == 0 || q.index == 2 || q.index == 4) {
            // range query at bottum
            // 1. check queryCondition ?= dicItem.Key
            if (q.rtype[0] == 1) {
                // [vl,vr]
                if ((param.itemR[0].key < q.rval[0] && q.rval[0] <= param.itemR[0].nxt) == false) return (false, 12201);
                if ((param.itemR[1].key <= q.rval[1] && q.rval[1] < param.itemR[1].nxt) == false) return (false, 1202);
            } else if (q.rtype[0] == 2) {
                // (vl,vr]
                if ((param.itemR[0].key <= q.rval[0] && q.rval[0] < param.itemR[0].nxt) == false) return (false, 12203);
                if ((param.itemR[1].key <= q.rval[1] && q.rval[1] < param.itemR[1].nxt) == false) return (false, 12204);
            } else if (q.rtype[0] == 3) {
                // [vl,vr)
                if ((param.itemR[0].key < q.rval[0] && q.rval[0] <= param.itemR[0].nxt) == false) return (false, 12205);
                if ((param.itemR[1].key < q.rval[1] && q.rval[1] <= param.itemR[1].nxt) == false) return (false, 12206);
            } else if (q.rtype[0] == 4) {
                // [vl,vr)
                if ((param.itemR[0].key <= q.rval[0] && q.rval[0] < param.itemR[0].nxt) == false) return (false, 12207);
                if ((param.itemR[1].key < q.rval[1] && q.rval[1] <= param.itemR[1].nxt) == false) return (false, 12208);
            } else {
                return (false, 12209);
            }

            // 2. check set difference, e(fR, item0.value)==e(G1, item1.value)
            p1[offset] = negate(param.fR); // negate the first element
            p1[offset + 1] = G1;
            p2[offset] = param.itemR[0].value;
            p2[offset + 1] = param.itemR[1].value;

            // 3. check inverse value
            if (checkItemInv(param.itemR[0], param.itemInv[layer + 1]) == false) return (false, 777);
            if (checkItemInv(param.itemR[1], param.itemInv[layer + 2]) == false) return (false, 778);

            // 4. pairing(negate(itemDigest), witness, dicDigest, G2) for each item
            // e(digest1, G2)==e(sG1+inv[item[1]]*G1, w[elen+2*i])
            // e(digest1, G2)==e(sG1+inv[item[2]]*G1, w[elen+2*i+1])
            // digest1 is the item value of layer0
            p1[offset + 2] = negate(param.itemE[layer].value);
            p1[offset + 3] = plus(scalar_mul(G1, param.itemInv[layer + 1]), sG1);
            p1[offset + 4] = negate(param.itemE[layer].value);
            p1[offset + 5] = plus(scalar_mul(G1, param.itemInv[layer + 2]), sG1);
            p2[offset + 2] = G2;
            p2[offset + 3] = param.itemWit[layer + 1];
            p2[offset + 4] = G2;
            p2[offset + 5] = param.itemWit[layer + 2];
        } else {
            // equivalent query at bottum
            // 1. check queryCondition ?= dicItem.Key
            if (q.eval[layer + 1] != param.itemE[layer + 1].key) return (false, 13201);
            // 2. check middle result digest mR
            if (param.fR.X != param.itemE[layer + 1].value.X || param.fR.Y != param.itemE[layer + 1].value.Y) return (false, 1302);
            // 3. check inverse value
            if (checkItemEInv(param.itemE[layer + 1], param.itemInv[layer + 1]) == false) return (false, 14);
            p1[offset] = negate(plus(scalar_mul(G1, param.itemInv[layer + 1]), sG1));
            p1[offset + 1] = param.itemE[layer].value;
            p2[offset] = param.itemWit[layer + 1];
            p2[offset + 1] = G2;
        }
        res = pairingMulti(p1, p2);
        if (res == false) {
            return (false, 5555);
        }
        return (true, 1555);
    }

    event TestGas2Result(bool success, uint result);

    function TestGas0Multi() public returns (bool, uint) {
        testgas = 0;
        return (true, 0);
    }

    function TestGas1Multi(QueryMulti calldata q, VerifySumParam calldata param) public returns (bool, uint) {
        testgas = param.sum;
        return (true, 0);
    }

    function TestGas2Multi(QueryMulti calldata q, VerifySumParam calldata param) public returns (bool ver, uint flag) {
        (ver, flag) = VerifyQueryMulti(q, param);
        emit TestGas2Result(ver, flag);
        if (ver) testgas = param.sum;
        return (ver, flag);
    }
}
