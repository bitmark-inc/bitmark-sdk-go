// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"errors"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

func transferOneSignature(sender account.Account, bitmarkId string, receiverAccountNumber string) (string, error) {
	transferParams, err := bitmark.NewTransferParams(receiverAccountNumber)
	if err != nil {
		return "", err
	}
	transferParams.FromBitmark(bitmarkId)
	transferParams.Sign(sender)
	txId, err := bitmark.Transfer(transferParams)

	return txId, err
}

func sendTransferOffer(sender account.Account, bitmarkId string, receiverAccountNumber string) error {
	offerParams, err := bitmark.NewOfferParams(receiverAccountNumber, nil)
	if err != nil {
		return err
	}
	offerParams.FromBitmark(bitmarkId)
	offerParams.Sign(sender)

	err = bitmark.Offer(offerParams)

	return err
}

func respondToTransferOffer(receiver account.Account, bitmarkId string, confirmation bitmark.OfferResponseAction) error {
	bmk, _ := bitmark.Get(bitmarkId)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, confirmation)
	rp.Sign(receiver)

	_, err := bitmark.Respond(rp)
	return err
}

func cancelTransferOffer(sender account.Account, bitmarkId string) error {
	bmk, _ := bitmark.Get(bitmarkId)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, "cancel")
	rp.Sign(sender)

	_, err := bitmark.Respond(rp)
	return err
}
