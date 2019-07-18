package main

import (
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

func issueBitmarks(issuer account.Account, assetID string, quantity int) ([]string, error) {
	issuanceParams := bitmark.NewIssuanceParams(assetID, quantity)
	issuanceParams.Sign(issuer)

	bitmarkIDs, err := bitmark.Issue(issuanceParams)

	return bitmarkIDs, err
}
