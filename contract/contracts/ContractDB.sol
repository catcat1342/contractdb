// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

contract ContractDB {
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

    struct Query {
        uint8[] eind;
        uint8[] rind;
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

    G1Point[6] internal digests; // name, bank, addr, value, rate, grade

    constructor() {
        digests[0].X = uint256(15951140197727074445119937693316174405529173087009701195705266480137303890556);
        digests[0].Y = uint256(9837468871853556325877590303753686632148266880063898020900862481265786922430);
        digests[1].X = uint256(4819624951180689431967439292553656652629575862474188011606754124232171981659);
        digests[1].Y = uint256(5943780203371573012511430230106106588971705926006719111767701022196344949651);
        digests[2].X = uint256(13977370901538128853417103227842899434850560937959359965642645195362494313162);
        digests[2].Y = uint256(8085313929396725633079740549220233411653596848243609049352314420457912816548);
        digests[3].X = uint256(1462198174415527659761713420558039402765030878382463160973829374879927981847);
        digests[3].Y = uint256(17071411402522037927990888758715215193262000639844498092564040248856348711842);
        digests[4].X = uint256(8319923958942147694896185760150349500838874672589881301200086050208375576866);
        digests[4].Y = uint256(13270915921715506622922410738049291351513360430794272418277738539960231500741);
        digests[5].X = uint256(12325337399494413194197845873321303161897533434305611641143141004189516794724);
        digests[5].Y = uint256(107937045436142327598318305234089692083842933277196542161046095092569264941);
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
    // do not use modInv, since it is expensive than uploading inv in calldata
    // function modInv(uint256 a) internal view returns (uint256) {
    //     bytes memory precompileData = abi.encode(32, 32, 32, a, PRIME_P - 2, PRIME_P);
    //     (bool ok, bytes memory data) = address(5).staticcall(precompileData);
    //     require(ok, 'expMod failed');
    //     return abi.decode(data, (uint256));
    // }

    // function verifySum(SumProof calldata sump) internal view returns (bool) {
    //     if (sump.fR.X == 1 && sump.fR.Y == 2) return (sump.sum == 0);
    //     if (modMul(sump.a0, sump.a0inv) != 1) return false;
    //     if (modMul(sump.a0inv, sump.a1) != sump.sum) return false;
    //     // e(w1,sG2) == e(fR-a0*G1, G2)
    //     // e(w2,sG2) == e(w1-a1*G1, G2)
    //     // p1: -w1, fR-a0G1, -w2, w1-a1G1
    //     // p2: sG2, G2, sG2, G2
    //     // G1Point memory tmp1 = negate(scalar_mul(G1, sump.a0));
    //     // G1Point memory tmp2 = negate(scalar_mul(G1, sump.a1));

    //     G1Point[] memory p1 = new G1Point[](4);
    //     G2Point[] memory p2 = new G2Point[](4);
    //     p1[0] = negate(sump.w1);
    //     p1[1] = plus(sump.fR, negate(scalar_mul(G1, sump.a0)));
    //     p1[2] = negate(sump.w2);
    //     p1[3] = plus(sump.w1, negate(scalar_mul(G1, sump.a1)));
    //     p2[0] = sG2;
    //     p2[1] = G2;
    //     p2[2] = sG2;
    //     p2[3] = G2;

    //     bool res = pairingMulti(p1, p2);

    //     return res;
    // }

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

    function VerifyQuery(Query calldata q, VerifySumParam calldata param) public view returns (bool ver, uint flag) {
        // if (verifySum(sump) == false) return (false, 13);

        uint i;
        uint offset;

        // verify dic items
        // each eq query requires one pairings
        // each range query requires three pairings
        offset = 4 + q.eind.length * 2 + q.rind.length * 6 + param.iset.length * 3 + 1;
        G1Point[] memory p1 = new G1Point[](offset);
        G2Point[] memory p2 = new G2Point[](offset);

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

        // G1Point[] memory p1 = new G1Point[](2);
        for (i = 0; i < q.eind.length; i++) {
            // 1. check queryCondition ?= dicItem.Key
            if (q.eval[i] != param.itemE[i].key) return (false, 1201);
            // 2. check middle result digest mR
            if (param.iset[i].X != param.itemE[i].value.X || param.iset[i].Y != param.itemE[i].value.Y) return (false, 1302);
            // 3. check inverse value
            if (checkItemEInv(param.itemE[i], param.itemInv[i]) == false) return (false, 14);

            offset = 4 + 2 * i;
            p1[offset] = negate(plus(scalar_mul(G1, param.itemInv[i]), sG1));
            p1[offset + 1] = digests[q.eind[i]];
            p2[offset] = param.itemWit[i];
            p2[offset + 1] = G2;
        }

        // range query
        for (i = 0; i < q.rind.length; i++) {
            // item1 = dic.itemR[2*i], item2 = dic.itemR[2*i+1]

            // 1. check queryCondition ?= dicItem.Key
            if (q.rtype[i] == 1) {
                // [vl,vr]
                if ((param.itemR[2 * i].key < q.rval[2 * i] && q.rval[2 * i] <= param.itemR[2 * i].nxt) == false) return (false, 1201);
                if ((param.itemR[2 * i + 1].key <= q.rval[2 * i + 1] && q.rval[2 * i + 1] < param.itemR[2 * i + 1].nxt) == false)
                    return (false, 1202);
            } else if (q.rtype[i] == 2) {
                // (vl,vr]
                if ((param.itemR[2 * i].key <= q.rval[2 * i] && q.rval[2 * i] < param.itemR[2 * i].nxt) == false) return (false, 1203);
                if ((param.itemR[2 * i + 1].key <= q.rval[2 * i + 1] && q.rval[2 * i + 1] < param.itemR[2 * i + 1].nxt) == false)
                    return (false, 1204);
            } else if (q.rtype[i] == 3) {
                // [vl,vr)
                if ((param.itemR[2 * i].key < q.rval[2 * i] && q.rval[2 * i] <= param.itemR[2 * i].nxt) == false) return (false, 1205);
                if ((param.itemR[2 * i + 1].key < q.rval[2 * i + 1] && q.rval[2 * i + 1] <= param.itemR[2 * i + 1].nxt) == false)
                    return (false, 1206);
            } else if (q.rtype[i] == 4) {
                // [vl,vr)
                if ((param.itemR[2 * i].key <= q.rval[2 * i] && q.rval[2 * i] < param.itemR[2 * i].nxt) == false) return (false, 1207);
                if ((param.itemR[2 * i + 1].key < q.rval[2 * i + 1] && q.rval[2 * i + 1] <= param.itemR[2 * i + 1].nxt) == false)
                    return (false, 1208);
            } else {
                return (false, 1209);
            }

            // 2. check middle result digest mR
            // e(iProof.iset[i+q.eind.length], dic.itemR[2*i].value)==e(G1,dic.itemR[2*i+1].value)
            offset = 4 + 2 * q.eind.length + 2 * i;
            p1[offset] = negate(param.iset[i + q.eind.length]);
            p1[offset + 1] = G1;
            p2[offset] = param.itemR[2 * i].value;
            p2[offset + 1] = param.itemR[2 * i + 1].value;

            // 3. check inverse value
            if (checkItemInv(param.itemR[2 * i], param.itemInv[q.eind.length + 2 * i]) == false) return (false, 777);
            if (checkItemInv(param.itemR[2 * i + 1], param.itemInv[q.eind.length + 2 * i + 1]) == false) return (false, 778);

            // 4. pairing(negate(itemDigest), witness, dicDigest, G2) for each item
            // e(digest[i], G2)==e(sG1+inv[elen+2*i]*G1, w[elen+2*i])
            // e(digest[i], G2)==e(sG1+inv[elen+2*i+1]*G1, w[elen+2*i+1])
            offset = 4 + 2 * q.eind.length + 2 * i + 2;
            p1[offset] = negate(digests[q.rind[i]]);
            p1[offset + 1] = plus(scalar_mul(G1, param.itemInv[q.eind.length + 2 * i]), sG1);
            p1[offset + 2] = negate(digests[q.rind[i]]);
            p1[offset + 3] = plus(scalar_mul(G1, param.itemInv[q.eind.length + 2 * i + 1]), sG1);
            p2[offset] = G2;
            p2[offset + 1] = param.itemWit[q.eind.length + 2 * i];
            p2[offset + 2] = G2;
            p2[offset + 3] = param.itemWit[q.eind.length + 2 * i + 1];
        }

        // 5. verify intersection
        // e(fR, iwit[i]) == e(iset[i], G2)
        // e(G1,G2) == e(igcd[0], iwit[0]) + e(igcd[1], iwit[1]) + ...
        // if len(iwit)==0, no need to verify intersection
        if (param.iset.length > 1) {
            offset = 4 + q.eind.length * 2 + q.rind.length * 6;
            for (i = 0; i < param.iset.length; i++) {
                p1[offset + 2 * i] = negate(param.fR);
                p1[offset + 2 * i + 1] = param.iset[i];
                p2[offset + 2 * i] = param.iwit[i];
                p2[offset + 2 * i + 1] = G2;
            }
            offset = offset + 2 * param.iset.length;
            p1[offset] = negate(G1);
            p2[offset] = G2;
            offset += 1;
            for (i = 0; i < param.iset.length; i++) {
                p1[offset + i] = param.igcd[i];
                p2[offset + i] = param.iwit[i];
            }
        }

        bool res = pairingMulti(p1, p2);
        if (res == false) {
            return (false, 55);
        }
        return (true, 15);
    }

    event TestGas2Result(bool success, uint result);

    function TestGas0() public returns (bool, uint) {
        testgas = 0;
        return (true, 0);
    }

    function TestGas1(Query calldata q, VerifySumParam calldata param) public returns (bool, uint) {
        testgas = param.sum;
        return (true, 0);
    }

    function TestGas2(Query calldata q, VerifySumParam calldata param) public returns (bool ver, uint flag) {
        (ver, flag) = VerifyQuery(q, param);
        emit TestGas2Result(ver, flag);
        if (ver) testgas = param.sum;
        return (ver, flag);
    }
}
