package miner

// func TestCalcNextLocalRequiredDifficultyCopy(t *testing.T) {

// 	// bigTotal := big.NewInt(0)
// 	// bigPledgeTotal := big.NewInt(0).SetUint64(0)
// 	// rate := uint64(0)
// 	// if bigTotal.Cmp(big.NewInt(0)) != 0 {
// 	// 	rate = bigPledgeTotal.Mul(bigPledgeTotal, big.NewInt(100)).Div(bigPledgeTotal, bigTotal).Uint64()
// 	// } else {
// 	// 	rate = 0
// 	// }

// 	// if rate < 0 || rate > 100 {
// 	// 	t.Error(fmt.Errorf("rate(%d) is not in the range of 0~100", rate))
// 	// }

// 	// basePledge := maxPledge * (100 - rate) / 100

// 	// localBits, err := calcNextLocalRequiredDifficultyCopy(536904082, basePledge, 0)
// 	// if err != nil {
// 	// 	t.Error(err)
// 	// }

// 	// if localBits != 536904082 {
// 	// 	t.Errorf("not equal:%d   local:%d", 536904082, localBits)
// 	// }

// 	a, err := calcNextLocalRequiredDifficultyCopy(big.NewInt(100), 10, 6)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	t.Error(a.String())
// }

// func TestBigToCommpact(t *testing.T) {
// 	num, ok := big.NewInt(0).SetString("-3138515231309674581056365894340303285240794530805638204815195683552231424", 10)
// 	if !ok {
// 		t.Error("set error")
// 	}

// 	t.Log(BigToCompact(num))

// 	nextNum := calcNextGlobalRequiredDifficultyCopy(1635451777, 1635451615, 520165556)
// 	t.Log(nextNum)

// 	nextBig := CompactToBig(nextNum)
// 	t.Log(nextBig)

// 	t.Log(big.NewInt(0).SetBytes(nextBig.Bytes()))

// 	t.Log(BigToCompact(nextBig))
// }

// // 3138493663352340860544530160728233669536255621090100172357210800664032052
// // 3138515231309674581056365894340303285240794530805638204815195683552231424
