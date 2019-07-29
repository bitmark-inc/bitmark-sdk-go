// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package account

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"golang.org/x/text/language"
)

type valid struct {
	seed          string
	phrases       []string
	accountNumber string
	network       sdk.Network
	version       Version
}

var (
	testnetAccount = valid{
		"9J87CAsHdFdoEu6N1unZk3sqhVBkVL8Z8",
		[]string{
			"name gaze apart lamp lift zone believe steak session laptop crowd hill",
			"箱 阻 起 歸 徹 矮 問 栽 瓜 鼓 支 樂",
		},
		"eMCcmw1SKoohNUf3LeioTFKaYNYfp2bzFYpjm3EddwxBSWYVCb",
		sdk.Testnet,
		V2,
	}

	livenetAccount = valid{
		"9J87GaPq7FR9Uacdi3FUoWpP6LbEpo1Ax",
		[]string{
			"surprise mesh walk inject height join sound minor margin over jewel venue",
			"薯 托 劍 景 擔 額 牢 痛 亦 軟 凱 誼",
		},
		"aiKFA9dKkNHPys3nSZrLTPusoocPqXSFp5EexsgQ1hbYUrJVne",
		sdk.Livenet,
		V2,
	}

	testnetDeprecatedAccount = valid{
		"5XEECt18HGBGNET1PpxLhy5CsCLG9jnmM6Q8QGF4U2yGb1DABXZsVeD",
		[]string{
			"accident syrup inquiry you clutch liquid fame upset joke glow best school repeat birth library combine access camera organ trial crazy jeans lizard science",
		},
		"ec6yMcJATX6gjNwvqp8rbc4jNEasoUgbfBBGGyV5NvoJ54NXva",
		sdk.Testnet,
		V1,
	}

	livenetDeprecatedAccount = valid{
		"5XEECqWqA47qWg86DR5HJ29HhbVqwigHUAhgiBMqFSBycbiwnbY639s",
		[]string{
			"ability panel leave spike mixture token voice certain today market grief crater cruise smart camera palm wheat rib swamp labor bid rifle piano glass",
		},
		"bDnC8nCaupb1AQtNjBoLVrGmobdALpBewkyYRG7kk2euMG93Bf",
		sdk.Livenet,
		V1,
	}

	langCheckSequence = []language.Tag{language.AmericanEnglish, language.TraditionalChinese}
)

func check(t *testing.T, a Account, data valid) {
	if a.Seed() != data.seed {
		t.Fatalf("invalid seed: expected = %s, actual = %s", testnetAccount.seed, a.Seed())
	}

	for i, p := range data.phrases {
		phrase, _ := a.RecoveryPhrase(langCheckSequence[i])
		if strings.Join(phrase, " ") != data.phrases[i] {
			t.Fatalf("invalid recovery phrase: expected = %s, actual = %s", p, phrase)
		}
	}

	if a.AccountNumber() != data.accountNumber {
		t.Fatalf("invalid account number: expected = %s, actual = %s", data.accountNumber, a.AccountNumber())
	}

	if a.Network() != data.network {
		t.Fatalf("invalid network: expected = %s, actual = %s", data.network, a.Network())
	}

	if a.Version() != data.version {
		t.Fatalf("invalid version: expected = %s, actual = %s", data.version, a.Version())
	}
}

func TestTestnetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	acctFromSeed, err := FromSeed(testnetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, testnetAccount)

	for i, lang := range langCheckSequence {
		phrase := strings.Split(testnetAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, testnetAccount)
	}
}

func TestLivenetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	acctFromSeed, err := FromSeed(livenetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, livenetAccount)

	for i, lang := range langCheckSequence {
		phrase := strings.Split(livenetAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, livenetAccount)
	}
}

func TestRejectAccountFromWrongNetwork(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	if _, err := FromSeed(testnetAccount.seed); err == nil {
		t.Fatal("seed from testnet not rejected")
	}

	for i, lang := range langCheckSequence {
		phrase := strings.Split(testnetAccount.phrases[i], " ")
		if _, err := FromRecoveryPhrase(phrase, lang); err == nil {
			t.Fatal("seed from testnet not rejected")
		}
	}

	if _, err := FromSeed(testnetDeprecatedAccount.seed); err == nil {
		t.Fatal("seed from testnet not rejected")
	}

	for i, lang := range langCheckSequence {
		if i >= len(testnetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(testnetDeprecatedAccount.phrases[i], " ")
		if _, err := FromRecoveryPhrase(phrase, lang); err == nil {
			t.Fatal("seed from testnet not rejected")
		}
	}

	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	if _, err := FromSeed(livenetAccount.seed); err == nil {
		t.Fatal("seed from livenet not rejected")
	}

	for i, lang := range langCheckSequence {
		phrase := strings.Split(livenetAccount.phrases[i], " ")
		if _, err := FromRecoveryPhrase(phrase, lang); err == nil {
			t.Fatal("seed from livenet not rejected")
		}
	}

	if _, err := FromSeed(livenetDeprecatedAccount.seed); err == nil {
		t.Fatal("seed from livenet not rejected")
	}

	for i, lang := range langCheckSequence {
		if i >= len(testnetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(livenetDeprecatedAccount.phrases[i], " ")
		if _, err := FromRecoveryPhrase(phrase, lang); err == nil {
			t.Fatal("seed from livenet not rejected")
		}
	}
}

func TestTestnetDeprecatedTestnetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	acctFromSeed, err := FromSeed(testnetDeprecatedAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, testnetDeprecatedAccount)

	for i, lang := range langCheckSequence {
		if i >= len(testnetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(testnetDeprecatedAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, testnetDeprecatedAccount)
	}
}

func TestTestnetDeprecatedLivenetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	acctFromSeed, err := FromSeed(livenetDeprecatedAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, livenetDeprecatedAccount)

	for i, lang := range langCheckSequence {
		if i >= len(livenetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(livenetDeprecatedAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, livenetDeprecatedAccount)
	}
}

func TestValidateAccountNumber(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	err := ValidateAccountNumber(testnetAccount.accountNumber)
	assert.NoError(t, err)

	err = ValidateAccountNumber(testnetDeprecatedAccount.accountNumber)
	assert.NoError(t, err)

	err = ValidateAccountNumber(livenetAccount.accountNumber)
	assert.EqualError(t, err, ErrWrongNetwork.Error())

	err = ValidateAccountNumber(livenetDeprecatedAccount.accountNumber)
	assert.EqualError(t, err, ErrWrongNetwork.Error())

	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	err = ValidateAccountNumber(testnetAccount.accountNumber)
	assert.EqualError(t, err, ErrWrongNetwork.Error())

	err = ValidateAccountNumber(testnetDeprecatedAccount.accountNumber)
	assert.EqualError(t, err, ErrWrongNetwork.Error())

	err = ValidateAccountNumber(livenetAccount.accountNumber)
	assert.NoError(t, err)

	err = ValidateAccountNumber(livenetDeprecatedAccount.accountNumber)
	assert.NoError(t, err)
}

func TestRecoverV1Account(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})
	acctFromPhrase, err := FromRecoveryPhrase(
		strings.Split("為 廠 磨 燕 華 已 忍 罵 稍 桌 搜 事 伴 爐 調 拜 輝 荒 巡 只 僚 空 之 填", " "),
		language.TraditionalChinese,
	)
	assert.NoError(t, err)
	assert.Equal(t, "fBHRe9f7g3vQgpyq8NGar3QVMfCSPNfDeKPYF5Maef6gCYKsP4", acctFromPhrase.AccountNumber())
}
