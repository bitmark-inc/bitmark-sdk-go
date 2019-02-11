package test

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	log "github.com/sirupsen/logrus"
)

type assetRegistrationTestCase struct {
	name     string
	metadata map[string]string
	hasError bool
}

var (
	sender   account.Account
	receiver account.Account

	bitmarkIds = make([]string, 0)

	assetRegistrationTestCases = []assetRegistrationTestCase{
		assetRegistrationTestCase{"RUN TestRegisterAsset", map[string]string{"k1": "v1"}, false},            // brand new asset
		assetRegistrationTestCase{"RUN TestRegisterAsset (1)", map[string]string{"k1": "v1"}, true},         // name mismatch the pending asset
		assetRegistrationTestCase{"RUN TestRegisterAsset", map[string]string{"k1": "v1", "k2": "v2"}, true}, // metadata mismatch the pending asset
	}
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

// This test case will try to register the same asset three times
// [1] brand new asset
// [2] same fingerprint but different name
// [3] same fingerprint but different metadata
func TestRegisterAsset(t *testing.T) {
	// generate new asset content
	content := time.Now().String()

	for i, tc := range assetRegistrationTestCases {
		p, _ := asset.NewRegistrationParams(tc.name, tc.metadata)
		p.SetFingerprint([]byte(content))
		p.Sign(sender)

		if _, err := asset.Register(p); (err == nil) == tc.hasError {
			t.Fatalf("test case (%d) failed", i)
		}
	}
}

// This test case wiil issue 4 bitmarks:
// bitmark #1 will be directly transferred to the receiver
// bitmark #2 will be offered and then canceled
// bitmark #3 will be offered and then rejected
// bitmark #4 will be offered and then accepted
func TestOwnershipChange(t *testing.T) {
	// generate new asset content
	content := time.Now().String()

	// register asset
	rp, _ := asset.NewRegistrationParams("RUN TestOwnershipChange", nil)
	rp.SetFingerprint([]byte(content))
	rp.Sign(sender)
	assetId, err := asset.Register(rp)
	if err != nil {
		t.Fatal(err)
	}
	log.WithField("asset_id", assetId).Info("asset is registered")

	// issue bitmarks
	ip := bitmark.NewIssuanceParams(assetId, 4)
	ip.Sign(sender)
	bitmarkIds, err := bitmark.Issue(ip)
	if err != nil {
		t.Fatalf("issue failed: %s", err)
	}
	if len(bitmarkIds) != 4 {
		t.Fatalf("incorrect quantity of bitmarks are issued: %d", len(bitmarkIds))
	}
	log.WithField("bitmark_ids", bitmarkIds).Info("bitmarks are issued")

	for _, bid := range bitmarkIds {
		bmk, err := bitmark.Get(bid, false)
		if err != nil {
			t.Fatalf("failed to query bitmark: %s", err)
		}

		if bmk.Status != "issuing" {
			t.Fatalf("bitmark status should be issuing: %s", bid)
		}
	}

	log.Info("waiting for the issue tx to be confirmed")
	time.Sleep(5 * time.Minute)

	// direct transfer the first bitmark
	if err := directTransfer(bitmarkIds[0]); err != nil {
		t.Fatalf("failed to directly transfer bitmark %s: %s", bitmarkIds[0], err)
	}

	// create transfer offers for the rest bitmarks
	for i, bid := range bitmarkIds[1:] {
		op := bitmark.NewOfferParams(receiver.AccountNumber(), nil)
		op.FromBitmark(bid)
		op.Sign(sender)
		if err := bitmark.Offer(op); err != nil {
			t.Fatalf("failed to create offer for bitmark %d: %s", i, err)
		}
	}

	// respond to offers
	if err := cancelOffer(bitmarkIds[1]); err != nil {
		t.Fatalf("failed to cancel offer for bitmark %s: %s", bitmarkIds[1], err)
	}
	if err := rejectOffer(bitmarkIds[2]); err != nil {
		t.Fatalf("failed to reject offer for bitmark %s: %s", bitmarkIds[2], err)
	}
	if err := acceptOffer(bitmarkIds[3]); err != nil {
		t.Fatalf("failed to accept offer for bitmark %s: %s", bitmarkIds[2], err)
	}
}

func TestIssueBitmarksForNonExsistingAsset(t *testing.T) {
	nonExsistingAssetId := "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	ip := bitmark.NewIssuanceParams(nonExsistingAssetId, 100)
	ip.Sign(sender)
	_, err := bitmark.Issue(ip)
	if err == nil {
		t.Fatalf("issue should have been rejected")
	}
}
func TestIssueMoreBitmarks(t *testing.T) {
	exsistingAssetId := "f738f1a797a4b97e9f43d26764d22242a0507b180b8bbb370df39a6219d1b1b9d52b4bca6335ddf51966c1d62d388c9dba4b633a126265e66ec168a74f980d92"
	ip := bitmark.NewIssuanceParams(exsistingAssetId, 5)
	ip.Sign(sender)
	bitmarkIds, err := bitmark.Issue(ip)
	if err != nil {
		t.Fatalf("issue failed: %s", err)
	}
	log.WithField("bitmark_ids", bitmarkIds).Info("bitmarks are issued")

	time.Sleep(5 * time.Minute)
	for _, bid := range bitmarkIds {
		bmk, err := bitmark.Get(bid, false)
		if bmk.Status != "settled" || err != nil {
			t.Fatalf("bitmark %s not settled: %s", bid, err)
		}
	}
}

func TestCreateTransferOfferImmediately(t *testing.T) {
	// generate new asset content
	content := time.Now().String()

	// register asset
	rp, _ := asset.NewRegistrationParams("test", nil)
	rp.SetFingerprint([]byte(content))
	rp.Sign(sender)
	assetId, _ := asset.Register(rp)

	// sender issues a new bitmark
	ip := bitmark.NewIssuanceParams(assetId, 1)
	ip.Sign(sender)
	bitmarkIds, _ := bitmark.Issue(ip)

	// sender can create the offer right after creating a new bitmark without waiting for confirmations
	offerParams := bitmark.NewOfferParams(receiver.AccountNumber(), nil)
	offerParams.FromBitmark(bitmarkIds[0])
	offerParams.Sign(sender)
	bitmark.Offer(offerParams)

	// receiver can accept the offer after the issue is confirmed
	// bitmark status: `issuing` -> `transferring`
	log.Info("waiting for the issue tx to be confirmed")
	time.Sleep(5 * time.Minute)

	bmk, _ := bitmark.Get(bitmarkIds[0], false) // bitmark status: offering
	respParams := bitmark.NewTransferResponseParams(bmk, "accept")
	respParams.Sign(receiver)
	bitmark.Respond(respParams)
	bmk, _ = bitmark.Get(bitmarkIds[0], false)
	t.Logf("bitmark status %s", bmk.Status) // bitmark status: `transferring`
}

func TestCreateAndGrantShares(t *testing.T) {
	assetId, err := registerAsset()
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

func registerAsset() (string, error) {
	// generate new asset content
	content := time.Now().String()

	// register asset
	rp, _ := asset.NewRegistrationParams("RUN TestOwnershipChange", nil)
	rp.SetFingerprint([]byte(content))
	rp.Sign(sender)
	return asset.Register(rp)
}

func issueBitmarks(assetId string, quantity int) ([]string, error) {
	ip := bitmark.NewIssuanceParams(assetId, quantity)
	ip.Sign(sender)
	return bitmark.Issue(ip)
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
	if err != nil {
		return err
	}

	bmk, _ := bitmark.Get(bid, false)
	if !validBitmark(bmk, receiver.AccountNumber(), "transferring", true) {
		return fmt.Errorf("bitmark is not transferred: %+v", bmk)
	}
	return nil
}

func cancelOffer(bid string) error {
	bmk, _ := bitmark.Get(bid, false)
	if !validBitmark(bmk, sender.AccountNumber(), "offering", false) {
		return errors.New("bitmark is not offering")
	}

	rp := bitmark.NewTransferResponseParams(bmk, "cancel")
	rp.Sign(sender)
	if err := bitmark.Respond(rp); err != nil {
		return err
	}

	bmk, _ = bitmark.Get(bmk.Id, false)
	if !validBitmark(bmk, sender.AccountNumber(), "settled", true) {
		return errors.New("bitmark is not canceled")
	}

	return nil
}

func rejectOffer(bid string) error {
	bmk, _ := bitmark.Get(bid, false)
	if !validBitmark(bmk, sender.AccountNumber(), "offering", false) {
		return errors.New("bitmark is not offering")
	}

	// receiver rejects the offer
	rp := bitmark.NewTransferResponseParams(bmk, "reject")
	rp.Sign(receiver)

	if err := bitmark.Respond(rp); err != nil {
		return err
	}

	bmk, _ = bitmark.Get(bid, false)
	if !validBitmark(bmk, sender.AccountNumber(), "settled", true) {
		return errors.New("bitmark is not rejected")
	}
	return nil
}

func acceptOffer(bid string) error {
	bmk, _ := bitmark.Get(bid, false)
	if !validBitmark(bmk, sender.AccountNumber(), "offering", false) {
		return errors.New("bitmark is not offering")
	}

	// receiver wants to accept the offer
	rp := bitmark.NewTransferResponseParams(bmk, "accept")
	rp.Sign(receiver)
	if err := bitmark.Respond(rp); err != nil {
		return err
	}

	bmk, _ = bitmark.Get(bid, false)
	if !validBitmark(bmk, receiver.AccountNumber(), "transferring", true) {
		return errors.New("bitmark is not transferred")
	}
	return nil
}

func validBitmark(bmk *bitmark.Bitmark, owner, status string, emptyOffer bool) bool {
	return bmk.Owner == owner && bmk.Status == status && (bmk.Offer == nil) == emptyOffer
}
