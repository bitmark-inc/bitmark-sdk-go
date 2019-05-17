package test

import (
	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	sender   account.Account
	receiver account.Account
)

func init() {
	network := os.Getenv("SDK_TEST_NETWORK")
	token := os.Getenv("SDK_TEST_API_TOKEN")
	cfg := &sdk.Config{
		HTTPClient: http.DefaultClient,
		Network:    sdk.Network(network),
		APIToken:   token,
	}
	sdk.Init(cfg)

	sender, _ = account.FromSeed(os.Getenv("SENDER_SEED"))
	receiver, _ = account.FromSeed(os.Getenv("RECEIVER_SEED"))
}

func TestRegisterExistingAsset(t *testing.T) {
	params, _ := asset.NewRegistrationParams("another name", nil)
	params.SetFingerprint([]byte("Fri May 10 14:01:41 CST 2019"))
	params.Sign(sender)
	_, err := asset.Register(params)
	assert.Error(t, err)
}

// This test case wiil issue 4 bitmarks and check test the ownership changes:
// bitmark #1 will be directly transferred to the receiver
// bitmark #2 will be offered and then canceled
// bitmark #3 will be offered and then rejected
// bitmark #4 will be offered and then accepted
func TestOwnershipChange(t *testing.T) {
	assetId := mustRegisterAsset(t, "", []byte(time.Now().String()))
	log.WithField("asset_id", assetId).Info("asset is registered")

	bitmarkIds := mustIssueBitmarks(t, assetId, 4)
	if len(bitmarkIds) != 4 {
		if !assert.Equal(t, 4, len(bitmarkIds), "more or less bitmarks are issued") {
			t.Fatal()
		}
	}
	for _, bid := range bitmarkIds {
		log.WithField("bitmark_id", bid).Info("bitmark is issued")
		verifyBitmark(t, bid, sender.AccountNumber(), "issuing")
	}

	log.Info("waiting for the issues to be confirmed")
	for {
		if txsAreReady(bitmarkIds) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	// direct transfer the first bitmark
	mustDirectTransfer(t, bitmarkIds[0])
	verifyBitmark(t, bitmarkIds[0], receiver.AccountNumber(), "transferring")

	// create and reply to transfer offers for the rest bitmarks
	for i, bid := range bitmarkIds[1:] {
		mustCreateOffer(t, bid)
		bmk := verifyBitmark(t, bid, sender.AccountNumber(), "offering")
		switch i {
		case 0: // cancel offer for bitmarkIds[1]
			mustCancelOffer(t, bmk)
			verifyBitmark(t, bitmarkIds[1], sender.AccountNumber(), "settled")
		case 1: // reject offer for bitmarkIds[2]
			mustRejectOffer(t, bmk)
			verifyBitmark(t, bitmarkIds[2], sender.AccountNumber(), "settled")
		case 2: // accept offer for bitmarkIds[3]
			mustAcceptOffer(t, bmk)
			verifyBitmark(t, bitmarkIds[3], receiver.AccountNumber(), "transferring")
		}
	}
}

func TestIssueBitmarksForNonExsistingAsset(t *testing.T) {
	nonExsistingAssetId := "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	params := bitmark.NewIssuanceParams(nonExsistingAssetId, 1)
	params.Sign(sender)
	_, err := bitmark.Issue(params)
	assert.Error(t, err)
}

func TestIssueMoreBitmarks(t *testing.T) {
	exsistingAssetId := "f738f1a797a4b97e9f43d26764d22242a0507b180b8bbb370df39a6219d1b1b9d52b4bca6335ddf51966c1d62d388c9dba4b633a126265e66ec168a74f980d92"
	bitmarkIds := mustIssueBitmarks(t, exsistingAssetId, 1)

	log.Info("waiting for the issue to be confirmed...")
	for {
		if txsAreReady(bitmarkIds) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	for _, bid := range bitmarkIds {
		verifyBitmark(t, bid, sender.AccountNumber(), "settled")
	}
}

func TestCreateTransferOfferImmediately(t *testing.T) {
	assetId := mustRegisterAsset(t, "", []byte(time.Now().String()))
	bitmarkIds := mustIssueBitmarks(t, assetId, 1)

	// sender can create the offer right after creating a new bitmark without waiting for confirmations
	mustCreateOffer(t, bitmarkIds[0])
	verifyBitmark(t, bitmarkIds[0], sender.AccountNumber(), "issuing")

	log.Info("waiting for the issue to be confirmed")
	for {
		if txsAreReady(bitmarkIds) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	bmk := verifyBitmark(t, bitmarkIds[0], sender.AccountNumber(), "offering")
	mustAcceptOffer(t, bmk)
	verifyBitmark(t, bitmarkIds[0], receiver.AccountNumber(), "transferring")
}

func TestCreateAndGrantShares(t *testing.T) {
	assetId := mustRegisterAsset(t, "", []byte(time.Now().String()))
	bitmarkIds := mustIssueBitmarks(t, assetId, 1)

	log.Info("waiting for the issue to be confirmed...")
	for {
		if txsAreReady(bitmarkIds) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	params := bitmark.NewShareParams(10)
	params.FromBitmark(bitmarkIds[0])
	params.Sign(sender)
	txId, shareId, err := bitmark.CreateShares(params)
	if err != nil {
		t.Fatalf("failed to create shares: %s", err)
	}
	log.WithField("share_id", shareId).WithField("tx_id", txId).Info("shares are created")

	log.Info("waiting for the share tx to be confirmed...")
	for {
		if txsAreReady([]string{txId}) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	share, err := bitmark.GetShareBalance(shareId, sender.AccountNumber())
	if err != nil {
		t.Fatalf("failed to query shares: %s", err)
	}
	if share.Balance != 10 || share.Available != 10 {
		t.Fatalf("incorrect balance of sender")
	}

	grantParams := bitmark.NewShareGrantingParams(shareId, receiver.AccountNumber(), 5, nil)
	// TODO: how to decide before block
	grantParams.BeforeBlock(14817)
	grantParams.Sign(sender)
	if _, err := bitmark.GrantShare(grantParams); err != nil {
		t.Fatalf("failed to grant shares: %s", err)
	}

	offers, err := bitmark.ListShareOffers(sender.AccountNumber(), receiver.AccountNumber())
	if err != nil {
		t.Fatalf("failed to query share offers: %s", err)
	}
	replyParams := bitmark.NewGrantResponseParams(offers[0].Id, &offers[0].Record, bitmark.Accept)
	replyParams.Sign(receiver)

	txId, err = bitmark.ReplyShareOffer(replyParams)
	if err != nil {
		t.Fatalf("failed to reply share offer: %s", err)
	}
	log.WithField("tx_id", txId).Info("shares are granted")

	senderShare, _ := bitmark.GetShareBalance(shareId, sender.AccountNumber())
	if senderShare.Balance != 10 || senderShare.Available != 5 {
		t.Fatalf("incorrect balance of sender")
	}

	receiverShare, _ := bitmark.GetShareBalance(shareId, receiver.AccountNumber())
	if receiverShare.Balance != 0 || receiverShare.Available != 0 {
		t.Fatalf("incorrect balance of receiver")
	}

	log.Info("waiting for the grant tx to be confirmed...")
	for {
		if txsAreReady([]string{txId}) {
			break
		}
		time.Sleep(30 * time.Second)
	}

	senderShare, _ = bitmark.GetShareBalance(shareId, sender.AccountNumber())
	if senderShare.Balance != 5 || senderShare.Available != 5 {
		t.Fatalf("incorrect balance of sender")
	}

	receiverShare, _ = bitmark.GetShareBalance(shareId, receiver.AccountNumber())
	if receiverShare.Balance != 5 || receiverShare.Available != 5 {
		t.Fatalf("incorrect balance of receiver")
	}
}

func txsAreReady(txIds []string) bool {
	for _, txId := range txIds {
		tx, _ := tx.Get(txId, false)
		if tx != nil && tx.Status != "confirmed" {
			return false
		}
	}
	return true
}

func mustRegisterAsset(t *testing.T, name string, content []byte) string {
	params, _ := asset.NewRegistrationParams(name, nil)
	params.SetFingerprint(content)
	params.Sign(sender)
	assetId, err := asset.Register(params)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	log.WithField("asset_id", assetId).Info("asset is registered")

	return assetId
}

func mustIssueBitmarks(t *testing.T, assetId string, quantity int) []string {
	params := bitmark.NewIssuanceParams(assetId, quantity)
	params.Sign(sender)
	bitmarkIds, err := bitmark.Issue(params)
	if !assert.NoError(t, err) {
		t.Fatal()
	}

	for i, bitmarkId := range bitmarkIds {
		log.WithField("bitmark_id", bitmarkId).Infof("bitmark #%d is issued", i+1)
	}

	return bitmarkIds
}

func mustDirectTransfer(t *testing.T, bid string) {
	params := bitmark.NewTransferParams(receiver.AccountNumber())
	params.FromBitmark(bid)
	params.Sign(sender)
	_, err := bitmark.Transfer(params)
	if !assert.NoError(t, err) {
		t.Fatal()
	}
}

func mustCreateOffer(t *testing.T, bid string) {
	params := bitmark.NewOfferParams(receiver.AccountNumber(), nil)
	params.FromBitmark(bid)
	params.Sign(sender)
	if !assert.NoError(t, bitmark.Offer(params)) {
		t.Fatal()
	}
}

func mustCancelOffer(t *testing.T, bmk *bitmark.Bitmark) {
	params := bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(sender)
	if !assert.NoError(t, bitmark.Respond(params)) {
		t.Fatal()
	}
}

func mustRejectOffer(t *testing.T, bmk *bitmark.Bitmark) {
	params := bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(receiver)
	if !assert.NoError(t, bitmark.Respond(params)) {
		t.Fatal()
	}
}

func mustAcceptOffer(t *testing.T, bmk *bitmark.Bitmark) {
	params := bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(receiver)
	if !assert.NoError(t, bitmark.Respond(params)) {
		t.Fatal()
	}
}

func verifyBitmark(t *testing.T, bitmarkId, owner, status string) *bitmark.Bitmark {
	time.Sleep(5 * time.Second)
	bmk, err := bitmark.Get(bitmarkId, false)
	if !assert.NoError(t, err) || !assert.Equal(t, owner, bmk.Owner) || !assert.Equal(t, status, bmk.Status) {
		t.Fatal()
	}
	return bmk
}
