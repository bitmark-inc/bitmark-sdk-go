// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

func issueBitmarks(issuer account.Account, assetID string, quantity int) ([]string, error) {
	issuanceParams, _ := bitmark.NewIssuanceParams(assetID, quantity)
	issuanceParams.Sign(issuer)

	bitmarkIDs, err := bitmark.Issue(issuanceParams)

	return bitmarkIDs, err
}
