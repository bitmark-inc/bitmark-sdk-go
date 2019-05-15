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
	_, err := registerAsset("another name", []byte("Fri May 10 14:01:41 CST 2019"))
	assert.Error(t, err)
}

// This test case wiil issue 4 bitmarks and check test the ownership changes:
// bitmark #1 will be directly transferred to the receiver
// bitmark #2 will be offered and then canceled
// bitmark #3 will be offered and then rejected
// bitmark #4 will be offered and then accepted
func TestOwnershipChange(t *testing.T) {
	// register asset
	assetId, err := registerAsset("", []byte(time.Now().String()))
	if err != nil {
		assert.NoError(t, err)
	}
	log.WithField("asset_id", assetId).Info("asset is registered")

	// issue bitmarks
	bitmarkIds, err := issueBitmarks(assetId, 4)
	if err != nil {
		assert.NoError(t, err)
	}
	if len(bitmarkIds) != 4 {
		assert.Equal(t, 4, len(bitmarkIds), "more or less bitmarks are issued")
	}
	for _, bid := range bitmarkIds {
		log.WithField("bitmark_id", bid).Info("bitmark is issued")
		vefiryBitmark(t, bid, sender.AccountNumber(), "issuing")
	}

	log.Info("waiting for the issues to be confirmed")
	time.Sleep(5 * time.Minute)

	// direct transfer the first bitmark
	assert.NoError(t, directTransfer(bitmarkIds[0]))
	vefiryBitmark(t, bitmarkIds[0], receiver.AccountNumber(), "transferring")

	// create and reply to transfer offers for the rest bitmarks
	for i, bid := range bitmarkIds[1:] {
		assert.NoError(t, createOffer(bid))
		bmk := vefiryBitmark(t, bid, sender.AccountNumber(), "offering")
		switch i {
		case 0: // cancel offer for bitmarkIds[1]
			assert.NoError(t, cancelOffer(bmk))
			vefiryBitmark(t, bitmarkIds[1], sender.AccountNumber(), "settled")
		case 1: // reject offer for bitmarkIds[2]
			assert.NoError(t, rejectOffer(bmk))
			vefiryBitmark(t, bitmarkIds[2], sender.AccountNumber(), "settled")
		case 2: // accept offer for bitmarkIds[3]
			assert.NoError(t, acceptOffer(bmk))
			vefiryBitmark(t, bitmarkIds[3], receiver.AccountNumber(), "transferring")
		}
	}
}

func TestIssueBitmarksForNonExsistingAsset(t *testing.T) {
	nonExsistingAssetId := "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	_, err := issueBitmarks(nonExsistingAssetId, 100)
	assert.Error(t, err)
}

func TestIssueMoreBitmarks(t *testing.T) {
	exsistingAssetId := "f738f1a797a4b97e9f43d26764d22242a0507b180b8bbb370df39a6219d1b1b9d52b4bca6335ddf51966c1d62d388c9dba4b633a126265e66ec168a74f980d92"
	bitmarkIds, err := issueBitmarks(exsistingAssetId, 1)
	assert.NoError(t, err)
	log.WithField("bitmark_id", bitmarkIds[0]).Info("bitmark is issued")

	time.Sleep(5 * time.Minute)

	for _, bid := range bitmarkIds {
		vefiryBitmark(t, bid, sender.AccountNumber(), "settled")
	}
}

func TestCreateTransferOfferImmediately(t *testing.T) {
	// register asset
	assetId, err := registerAsset("", []byte(time.Now().String()))
	assert.NoError(t, err)

	// sender issues a new bitmark
	bitmarkIds, err := issueBitmarks(assetId, 1)
	assert.NoError(t, err)

	// sender can create the offer right after creating a new bitmark without waiting for confirmations
	assert.NoError(t, createOffer(bitmarkIds[0]))

	// receiver can accept the offer after the issue is confirmed
	// bitmark status: `issuing` -> `transferring`
	log.Info("waiting for the issue tx to be confirmed")
	time.Sleep(5 * time.Minute)

	bmk := vefiryBitmark(t, bitmarkIds[0], sender.AccountNumber(), "offering")
	assert.NoError(t, acceptOffer(bmk))
	vefiryBitmark(t, bitmarkIds[0], receiver.AccountNumber(), "transferring")
}

func TestCreateAndGrantShares(t *testing.T) {
	assetId, err := registerAsset("", []byte(time.Now().String()))
	if err != nil {
		t.Fatalf("failed to register a new asset: %s", err)
	}
	log.WithField("asset_id", assetId).Info("asset is registered")

	bitmarkIds, err := issueBitmarks(assetId, 1)
	if err != nil {
		t.Fatalf("failed to issue a bitmark: %s", err)
	}
	log.WithField("bitmark_ids", bitmarkIds).Info("bitmarks are issued")

	log.Info("waiting for the issue tx to be confirmed...")
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

func registerAsset(name string, content []byte) (string, error) {
	params, _ := asset.NewRegistrationParams(name, nil)
	params.SetFingerprint(content)
	params.Sign(sender)
	return asset.Register(params)
}

func issueBitmarks(assetId string, quantity int) ([]string, error) {
	params := bitmark.NewIssuanceParams(assetId, quantity)
	params.Sign(sender)
	return bitmark.Issue(params)
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

func directTransfer(bid string) error {
	params := bitmark.NewTransferParams(receiver.AccountNumber())
	params.FromBitmark(bid)
	params.Sign(sender)
	_, err := bitmark.Transfer(params)
	return err
}

func createOffer(bid string) error {
	params := bitmark.NewOfferParams(receiver.AccountNumber(), nil)
	params.FromBitmark(bid)
	params.Sign(sender)
	return bitmark.Offer(params)
}

func cancelOffer(bmk *bitmark.Bitmark) error {
	params := bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(sender)
	return bitmark.Respond(params)
}

func rejectOffer(bmk *bitmark.Bitmark) error {
	params := bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(receiver)
	return bitmark.Respond(params)
}

func acceptOffer(bmk *bitmark.Bitmark) error {
	params := bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(receiver)
	return bitmark.Respond(params)
}

func vefiryBitmark(t *testing.T, bitmarkId, owner, status string) *bitmark.Bitmark {
	time.Sleep(5 * time.Second)
	bmk, err := bitmark.Get(bitmarkId, false)
	assert.NoError(t, err)
	assert.Equal(t, owner, bmk.Owner)
	assert.Equal(t, status, bmk.Status)
	return bmk
}
