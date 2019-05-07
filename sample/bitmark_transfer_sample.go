package main

import (
	"errors"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

func transferOneSignature(sender account.Account, bitmarkId string, receiverAccountNumber string) (string, error) {
	transferParams := bitmark.NewTransferParams(receiverAccountNumber)
	transferParams.FromBitmark(bitmarkId)
	transferParams.Sign(sender)
	txId, err := bitmark.Transfer(transferParams)

	return txId, err
}

func sendTransferOffer(sender account.Account, bitmarkId string, receiverAccountNumber string) error {
	offerParams := bitmark.NewOfferParams(receiverAccountNumber, nil)
	offerParams.FromBitmark(bitmarkId)
	offerParams.Sign(sender)

	err := bitmark.Offer(offerParams)

	return err
}

func respondToTransferOffer(receiver account.Account, bitmarkId string, confirmation bitmark.OfferResponseAction) error {
	bmk, _ := bitmark.Get(bitmarkId, false)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, confirmation)
	rp.Sign(receiver)

	err := bitmark.Respond(rp)
	return err
}

func cancelTransferOffer(sender account.Account, bitmarkId string) error {
	bmk, _ := bitmark.Get(bitmarkId, false)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, "cancel")
	rp.Sign(sender)

	err := bitmark.Respond(rp)
	return err
}
