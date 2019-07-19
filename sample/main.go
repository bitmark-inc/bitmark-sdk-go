// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	"net/http"
)

const (
	API_TOKEN    = "YOUR_API_TOKEN"
	NETWORK_MODE = sdk.Testnet // sdk.Testnet or sdk.Livenet
)

func main() {
	//////////////////////////////////////
	// 0. SDK INITIALIZATION
	//////////////////////////////////////
	sdk.Init(&sdk.Config{
		Network:    NETWORK_MODE,
		APIToken:   API_TOKEN,
		HTTPClient: http.DefaultClient,
	})

	//////////////////////////////////////
	// 1. ACCOUNT
	//////////////////////////////////////

	// 1.1 Create totally new acc
	acc, _ := createNewAccount()
	fmt.Println("Account Number:", acc.AccountNumber())

	/*
	   1.2 Generate Recovery Phrase(twelve words)
	    Recovery Phrase(twelve words) is your secret key. You should keep it in safe place
	    (ex: write it down to paper, shouldn't store it on your machine) and don't reveal it to anyone else.
	*/

	recoveryPhrase, _ := getRecoveryPhraseFromAccount(acc)
	fmt.Println("Recovery Phrase:", recoveryPhrase)

	// 1.3 Create acc using Recovery Phrase
	acc, _ = getAccountFromRecoveryPhrase(recoveryPhrase)
	fmt.Println("Account Number:", acc.AccountNumber())

	//////////////////////////////////////
	// 2. REGISTER ASSET & ISSUE BITMARKS
	//////////////////////////////////////

	/*
	   2.1 Register asset
	   You need to register asset to Bitmark block-chain before you can issue bitmarks for it
	*/
	assetName := "YOUR_ASSET_NAME" // Asset length must be less than or equal 64 characters
	assetFilePath := "YOUR_ASSET_FILE_PATH"
	metadata := map[string]string{
		"k1": "v1",
		"k2": "v2",
	} // Metadata length must be less than or equal 2048 characters

	assetID, err := registerAsset(acc, assetName, assetFilePath, metadata)

	if err != nil {
		fmt.Println("Can not register asset!")
		panic(err)
	}

	fmt.Println("Asset ID:", assetID)

	/*
	   2.2 Issue bitmarks for asset
	   You need provide asset ID to issue bitmarks for asset
	*/

	assetID = "YOUR_ASSET_ID"
	quantity := 100 // Number of bitmarks you want to issue, quantity must be less than or equal 100.

	bitmarkIDs, err := issueBitmarks(acc, assetID, quantity)

	if err != nil {
		fmt.Println("Can not issue bitmarks!")
		panic(err)
	}

	fmt.Println("Bitmark IDs:", bitmarkIDs)

	//////////////////////////////////////
	// 3. QUERY
	//////////////////////////////////////

	// 3.1 Query bitmark/bitmarks

	/*
	   3.1.1 Query bitmarks
	   Ex: Query bitmarks which you are owner
	*/
	bitmarkQueryBuilder := bitmark.NewQueryParamsBuilder().OwnedBy(acc.AccountNumber())

	bitmarks, err := queryBitmarks(bitmarkQueryBuilder)
	if err != nil {
		fmt.Println("Can not query bitmarks!")
		panic(err)
	}

	fmt.Println("Bitmarks Length:", len(bitmarks))

	// 3.1.2 Query bitmark
	bitmarkID := "BITMARK_ID"
	bitmark, err := queryBitmarkByID(bitmarkID)

	if err != nil {
		fmt.Println("Can not query bitmark!")
		panic(err)
	}

	fmt.Println("Bitmark:", bitmark)

	// 3.2 Query transaction/transactions

	/*
	   3.2.1 Query transactions
	   Ex: Query transactions which you are owner
	*/
	txQueryBuilder := tx.NewQueryParamsBuilder().OwnedBy(acc.AccountNumber())

	txs, err := queryTransactions(txQueryBuilder)
	if err != nil {
		fmt.Println("Can not query transactions!")
		panic(err)
	}

	fmt.Println("Transactions Length:", len(txs))

	// 3.2.2 Query Transaction
	txID := "TRANSACTION_ID"
	tx, err := queryTransactionByID(txID)
	if err != nil {
		fmt.Println("Can not query transaction!")
		panic(err)
	}

	fmt.Println("Transaction:", tx)

	// 3.3 Query assets/asset

	/**
	 * 3.3.1 Query assets
	 * Ex: Query assets which you registered
	 */
	assetQueryBuilder := asset.NewQueryParamsBuilder().
		RegisteredBy(acc.AccountNumber())

	assets, err := queryAssets(assetQueryBuilder)
	if err != nil {
		fmt.Println("Can not query assets!")
		panic(err)
	}

	fmt.Println("Bitmarks Length:", len(assets))

	// 3.3.2 Query asset
	assetID = "ASSET_ID"
	asset, err := queryAssetByID(assetID)

	if err != nil {
		fmt.Println("Can not query asset!")
		panic(err)
	}

	fmt.Println("Asset:", asset)

	//////////////////////////////////////
	// 4. TRANSFER BITMARKS
	//////////////////////////////////////

	/*
	   4.1 Transfer bitmark using 1 signature
	   You can transfer your bitmark to another account without their acceptance.
	   Note: Your bitmark must be confirmed on Bitmark block-chain(status=settled) before you can transfer it. You can query bitmark by bitmark ID to check it's status.
	*/
	transferBitmarkID := "YOUR_BITMARK_ID"
	receiverAccountNumber := "ACCOUNT_NUMBER_YOU_WANT_TO_TRANSFER_BITMARK_TO"
	txIDResponse, err := transferOneSignature(acc, transferBitmarkID, receiverAccountNumber)

	if err != nil {
		fmt.Println("Can not transfer bitmark!")
		panic(err)
	}

	fmt.Println("txIDResponse:", txIDResponse)

	/*
	   4.2 Transfer bitmark using 2 signatures
	   When you transfer your bitmark to another account(receiver) using 2 signatures transfer, the receiver is able to accept or reject your transfer.
	   The flow is:
	   a. You(sender): Send a transfer offer to receiver
	   b. Receiver: Accept/Reject your transfer offer
	   Notes:
	   1. Your bitmark must be confirmed on Bitmark block-chain(status=settled) before you can transfer it. You can query bitmark by bitmark ID to check it's status.
	   2. You can cancel your transfer offer if the receiver doesn't accept/reject it yet.
	*/

	// YOUR CODE: Send transfer offer to receiver
	offerBitmarkID := "YOUR_BITMARK_ID"
	receiverAccountNumber = "ACCOUNT_NUMBER_YOU_WANT_TO_TRANSFER_BITMARK_TO"
	err = sendTransferOffer(acc, offerBitmarkID, receiverAccountNumber)

	if err != nil {
		fmt.Println("Can not send transfer offer")
		panic(err)
	}

	// 4.2.1 Receiver respond(accept/reject) your transfer offer
	// RECEIVER's CODE
	bitmarkID = "WILL_RECEIVE_BITMARK_ID"
	receiverAccount, _ := getAccountFromRecoveryPhrase("RECEIVER_RECOVERY_PHRASE")
	err = respondToTransferOffer(receiverAccount, bitmarkID, "accept")

	if err != nil {
		fmt.Println("Can not response to transfer offer")
		panic(err)
	}

	// 4.2.2 You cancel your own transfer offer
	// YOUR CODE
	bitmarkID = "YOUR_BITMARK_ID_SENT"
	err = cancelTransferOffer(acc, bitmarkID)
	if err != nil {
		fmt.Println("Can not send transfer offer")
		panic(err)
	}
}
