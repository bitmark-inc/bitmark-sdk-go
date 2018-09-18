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
	log "github.com/sirupsen/logrus"
)

type assetRegistrationTestCase struct {
	name     string
	metadata map[string]string
	hasError bool
}

var (
	senderSeed   = "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH"
	receiverSeed = "5XEECt4yuMK4xqBLr9ky5FBWpkAR6VHNZSz8fUzZDXPnN3D9MeivTSA"

	sender   *account.Account
	receiver *account.Account

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

	sender, _ = account.FromSeed(senderSeed)
	receiver, _ = account.FromSeed(receiverSeed)
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

// This test case wiil issue 4 bitmarks (2 by specifying quantity, 2 by specifying nonces).
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

	options := bitmark.QuantityOptions{
		Quantity: 2,
	}

	// issue bitmarks by specifying quantity
	ipq := bitmark.NewIssuanceParams(assetId, options)
	ipq.Sign(sender)
	bids, err := bitmark.Issue(ipq)
	if err != nil {
		t.Fatalf("issue for limited editions failed: %s", err)
	}
	if len(bids) != 2 {
		t.Fatalf("incorrect quantity of bitmarks are issued: %d", len(bids))
	}
	log.WithField("bitmark_ids", bids).Info("bitmarks are issued")
	bitmarkIds = append(bitmarkIds, bids...)

	// issue bitmarks by specifying nonces
	options.Nonces = []uint64{uint64(1), uint64(2)}
	ipn := bitmark.NewIssuanceParams(assetId, options) // test if nonces take precedence over quanity
	ipn.Sign(sender)
	bids, err = bitmark.Issue(ipn)
	if err != nil {
		t.Fatalf("issue on demand failed: %s", err)
	}
	if len(bids) != 2 {
		t.Fatalf("incorrect quantity of bitmarks are issued")
	}
	log.WithField("bitmark_ids", bids).Info("bitmarks are issued")
	bitmarkIds = append(bitmarkIds, bids...)

	log.Info("waiting for the issue txs to be confirmed")
	time.Sleep(3 * time.Minute)
	// asset, _ := asset.Get(assetId)
	// if asset.Status != "confirmed" {
	// 	continue
	// }
	// for _, bid := range bitmarkIds {
	// 	bmk, _ := bitmark.Get(bid, false)
	// 	if bmk.Status != "confirmed" {
	// 		continue
	// 	}
	// }
	// break

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

	rp := bitmark.NewResponseParams(bmk, "cancel")
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
	rp := bitmark.NewResponseParams(bmk, "reject")
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
	rp := bitmark.NewResponseParams(bmk, "accept")
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
