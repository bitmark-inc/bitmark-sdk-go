package main

import (
	"errors"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

func transferOneSignature(sender account.Account, bitmarkID string, receiverAccountNumber string) (string, error) {
	transferParams, err := bitmark.NewTransferParams(receiverAccountNumber)
	if err != nil {
		return "", err
	}
	transferParams.FromBitmark(bitmarkID)
	transferParams.Sign(sender)
	txID, err := bitmark.Transfer(transferParams)

	return txID, err
}

func sendTransferOffer(sender account.Account, bitmarkID string, receiverAccountNumber string) error {
	offerParams, err := bitmark.NewOfferParams(receiverAccountNumber, nil)
	if err != nil {
		return err
	}
	offerParams.FromBitmark(bitmarkID)
	offerParams.Sign(sender)

	err = bitmark.Offer(offerParams)

	return err
}

func respondToTransferOffer(receiver account.Account, bitmarkID string, confirmation bitmark.OfferResponseAction) error {
	bmk, _ := bitmark.Get(bitmarkID)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, confirmation)
	rp.Sign(receiver)

	_, err := bitmark.Respond(rp)
	return err
}

func cancelTransferOffer(sender account.Account, bitmarkID string) error {
	bmk, _ := bitmark.Get(bitmarkID)

	if bmk != nil && bmk.Status != "offering" {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, "cancel")
	rp.Sign(sender)

	_, err := bitmark.Respond(rp)
	return err
}
