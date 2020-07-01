// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"golang.org/x/text/language"
	"strings"
)

func createNewAccount() (account.Account, error) {
	acc, err := account.New()
	return acc, err
}

func getAccountFromRecoveryPhrase(recoveryPhrase string) (account.Account, error) {
	acc, err := account.FromRecoveryPhrase(strings.Split(recoveryPhrase, " "), language.AmericanEnglish)
	return acc, err
}

func getRecoveryPhraseFromAccount(acc account.Account) (string, error) {
	recoveryPhrase, err := acc.RecoveryPhrase(language.AmericanEnglish)

	if err != nil {
		return "", err
	} else {
		return strings.Join(recoveryPhrase, " "), nil
	}
}
