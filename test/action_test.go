// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
)

// func TestCreateAndGrantShares(t *testing.T) {
// 	assetID := mustRegisterAsset(t, "", []byte(time.Now().String()))
// 	bitmarkIDs := mustIssueBitmarks(t, assetID, 1)

// 	log.Info("waiting for the issue to be confirmed...")
// 	for {
// 		if txsAreReady(bitmarkIDs) {
// 			break
// 		}
// 		time.Sleep(30 * time.Second)
// 	}

// 	params := bitmark.NewShareParams(10)
// 	params.FromBitmark(bitmarkIDs[0])
// 	params.Sign(sender)
// 	txID, shareID, err := bitmark.CreateShares(params)
// 	if err != nil {
// 		t.Fatalf("failed to create shares: %s", err)
// 	}
// 	log.WithField("share_id", shareID).WithField("tx_id", txID).Info("shares are created")

// 	log.Info("waiting for the share tx to be confirmed...")
// 	for {
// 		if txsAreReady([]string{txID}) {
// 			break
// 		}
// 		time.Sleep(30 * time.Second)
// 	}

// 	share, err := bitmark.GetShareBalance(shareID, sender.AccountNumber())
// 	if err != nil {
// 		t.Fatalf("failed to query shares: %s", err)
// 	}
// 	if share.Balance != 10 || share.Available != 10 {
// 		t.Fatalf("incorrect balance of sender")
// 	}

// 	grantParams := bitmark.NewShareGrantingParams(shareID, receiver.AccountNumber(), 5, nil)
// 	// TODO: how to decide before block
// 	grantParams.BeforeBlock(14817)
// 	grantParams.Sign(sender)
// 	if _, err := bitmark.GrantShare(grantParams); err != nil {
// 		t.Fatalf("failed to grant shares: %s", err)
// 	}

// 	offers, err := bitmark.ListShareOffers(sender.AccountNumber(), receiver.AccountNumber())
// 	if err != nil {
// 		t.Fatalf("failed to query share offers: %s", err)
// 	}
// 	replyParams := bitmark.NewGrantResponseParams(offers[0].ID, &offers[0].Record, bitmark.Accept)
// 	replyParams.Sign(receiver)

// 	txID, err = bitmark.ReplyShareOffer(replyParams)
// 	if err != nil {
// 		t.Fatalf("failed to reply share offer: %s", err)
// 	}
// 	log.WithField("tx_id", txID).Info("shares are granted")

// 	senderShare, _ := bitmark.GetShareBalance(shareID, sender.AccountNumber())
// 	if senderShare.Balance != 10 || senderShare.Available != 5 {
// 		t.Fatalf("incorrect balance of sender")
// 	}

// 	receiverShare, _ := bitmark.GetShareBalance(shareID, receiver.AccountNumber())
// 	if receiverShare.Balance != 0 || receiverShare.Available != 0 {
// 		t.Fatalf("incorrect balance of receiver")
// 	}

// 	log.Info("waiting for the grant tx to be confirmed...")
// 	for {
// 		if txsAreReady([]string{txID}) {
// 			break
// 		}
// 		time.Sleep(30 * time.Second)
// 	}

// 	senderShare, _ = bitmark.GetShareBalance(shareID, sender.AccountNumber())
// 	if senderShare.Balance != 5 || senderShare.Available != 5 {
// 		t.Fatalf("incorrect balance of sender")
// 	}

// 	receiverShare, _ = bitmark.GetShareBalance(shareID, receiver.AccountNumber())
// 	if receiverShare.Balance != 5 || receiverShare.Available != 5 {
// 		t.Fatalf("incorrect balance of receiver")
// 	}
// }

func txsAreReady(txIDs []string) bool {
	for _, txID := range txIDs {
		tx, _ := tx.Get(txID)
		if tx != nil && tx.Status != "confirmed" {
			return false
		}
	}
	return true
}
