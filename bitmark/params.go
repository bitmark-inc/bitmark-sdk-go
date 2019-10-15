// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package bitmark

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

type OfferResponseAction string

const Accept OfferResponseAction = "accept"
const Reject OfferResponseAction = "reject"
const Cancel OfferResponseAction = "cancel"

var nonceIndex uint64

type QuantityOptions struct {
	Nonces   []uint64
	Quantity int
}

type IssuanceParams struct {
	Issuances []*IssueRequest `json:"issues"`
}

type IssueRequest struct {
	AssetID   string `json:"asset_id" pack:"hex64"`
	Owner     string `json:"owner" pack:"account"`
	Nonce     uint64 `json:"nonce" pack:"uint64"`
	Signature string `json:"signature"`
}

func NewIssuanceParams(assetID string, quantity int) (*IssuanceParams, error) {
	if quantity < 1 {
		return nil, errors.New("quantity must be greater than or equal to 1")
	}

	ip := &IssuanceParams{
		Issuances: make([]*IssueRequest, 0),
	}

	builder := NewQueryParamsBuilder().ReferencedAsset(assetID)
	bitmarks, _, _ := List(builder)
	if len(bitmarks) == 0 {
		issuance := &IssueRequest{
			AssetID: assetID,
			Nonce:   0,
		}
		ip.Issuances = append(ip.Issuances, issuance)

		quantity--
	}

	for i := 0; i < quantity; i++ {
		atomic.AddUint64(&nonceIndex, 1)
		nonce := uint64(time.Now().UTC().Unix())*1000 + nonceIndex%1000
		issuance := &IssueRequest{
			AssetID: assetID,
			Nonce:   nonce,
		}
		ip.Issuances = append(ip.Issuances, issuance)
	}

	return ip, nil
}

// Sign all issunaces in a batch
func (p *IssuanceParams) Sign(issuer account.Account) error {
	for _, issuance := range p.Issuances {
		issuance.Owner = issuer.AccountNumber()
		message, err := utils.Pack(issuance)
		if err != nil {
			return err
		}
		issuance.Signature = hex.EncodeToString(issuer.Sign(message))
	}

	return nil
}

type TransferParams struct {
	Transfer *TransferRequest `json:"transfer"`
}

type TransferRequest struct {
	Link                    string   `json:"link" pack:"hex32"`
	Escrow                  *payment `json:"-" pack:"payment"` // optional escrow payment address
	Owner                   string   `json:"owner" pack:"account"`
	Signature               string   `json:"signature"`
	requireCountersignature bool
}

type payment struct {
	Currency string `json:"currency"`
	Address  string `json:"address"`
	Amount   uint64 `json:"amount,string"`
}

func NewTransferParams(receiver string) (*TransferParams, error) {
	if err := account.ValidateAccountNumber(receiver); err != nil {
		return nil, err
	}

	return &TransferParams{
		Transfer: &TransferRequest{
			Owner:                   receiver,
			requireCountersignature: false,
		},
	}, nil
}

// FromBitmark sets link asynchronously
func (t *TransferParams) FromBitmark(bitmarkID string) error {
	bitmark, err := Get(bitmarkID)
	if err != nil {
		return err
	}

	t.Transfer.Link = bitmark.LatestTxID
	return nil
}

// FromLatestTx sets link synchronously
func (t *TransferParams) FromLatestTx(txID string) {
	t.Transfer.Link = txID
}

func (t *TransferParams) Sign(sender account.Account) error {
	message, err := utils.Pack(t.Transfer)
	if err != nil {
		return err
	}
	t.Transfer.Signature = hex.EncodeToString(sender.Sign(message))
	return nil
}

// Copy of bitmark share structure
type ShareRequest struct {
	Link      string `json:"link" pack:"hex32"`
	Quantity  uint64 `json:"quantity" pack:"uint64"`
	Signature string `json:"signature"`
}

// ShareParams is the parameter for creating shares via core api
type ShareParams struct {
	Share *ShareRequest `json:"share"`
}

// NewShareParams returns ShareParams
func NewShareParams(quantity uint64) *ShareParams {
	return &ShareParams{
		Share: &ShareRequest{
			Quantity: quantity,
		},
	}
}

// FromBitmark will set the latest transaction for a target bitmark
func (s *ShareParams) FromBitmark(bitmarkID string) error {
	bitmark, err := Get(bitmarkID)
	if err != nil {
		return err
	}
	s.Share.Link = bitmark.LatestTxID
	return nil
}

// Sign will generate the signature for a share request
func (s *ShareParams) Sign(creator account.Account) error {
	message, err := utils.Pack(s.Share)
	if err != nil {
		return err
	}
	s.Share.Signature = hex.EncodeToString(creator.Sign(message))
	return nil
}

// Copy of bitmark share granting structure
type GrantRequest struct {
	ShareID     string `json:"shareID" pack:"hex32"`
	Quantity    uint64 `json:"quantity" pack:"uint64"`
	Owner       string `json:"owner" pack:"account"`
	Recipient   string `json:"recipient" pack:"account"`
	BeforeBlock uint64 `json:"beforeBlock" pack:"uint64"`
	Signature   string `json:"signature"`
}

// ShareGrantingParams is the parameter for granting shares to other accounts via core api
type ShareGrantingParams struct {
	Grant     *GrantRequest          `json:"record"`
	ExtraInfo map[string]interface{} `json:"extra_info"`
}

// NewShareGrantingParams returns ShareGrantingParams
func NewShareGrantingParams(shareID string, receiver string, quantity uint64, extraInfo map[string]interface{}) *ShareGrantingParams {
	return &ShareGrantingParams{
		Grant: &GrantRequest{
			ShareID:   shareID,
			Recipient: receiver,
			Quantity:  quantity,
		},
		ExtraInfo: extraInfo,
	}
}

// BeforeBlock will assign a block number which is the deadline of this request
func (s *ShareGrantingParams) BeforeBlock(blockNumber uint64) {
	s.Grant.BeforeBlock = blockNumber
}

// Sign will generate the signature for a granting request
func (s *ShareGrantingParams) Sign(granter account.Account) error {
	s.Grant.Owner = granter.AccountNumber()
	message, err := utils.Pack(s.Grant)
	if err != nil {
		return err
	}

	s.Grant.Signature = hex.EncodeToString(granter.Sign(message))
	return nil
}

// Copy of bitmark share granting structure with counter signature
type CountersignedGrantRequest struct {
	ShareID          string `json:"shareID" pack:"hex32"`
	Quantity         uint64 `json:"quantity" pack:"uint64"`
	Owner            string `json:"owner" pack:"account"`
	Recipient        string `json:"recipient" pack:"account"`
	BeforeBlock      uint64 `json:"beforeBlock" pack:"uint64"`
	Signature        string `json:"signature" pack:"hex64"`
	Countersignature string `json:"countersignature"`
}

// GrantResponseParams is the parameter for respond a share granting request
type GrantResponseParams struct {
	ID               string              `json:"id"`
	Action           OfferResponseAction `json:"action"`
	Countersignature string              `json:"countersignature"`
	auth             http.Header
	record           *CountersignedGrantRequest
}

// Sign will generate the signature for a granting responding request
func (g *GrantResponseParams) Sign(receiver account.Account) error {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		"updateOffer",
		g.ID,
		receiver.AccountNumber(),
		ts,
	}
	msg := strings.Join(parts, "|")
	sig := hex.EncodeToString(receiver.Sign([]byte(msg)))

	g.auth.Add("requester", receiver.AccountNumber())
	g.auth.Add("timestamp", ts)
	g.auth.Add("signature", sig)

	message, err := utils.Pack(g.record)
	if err != nil {
		return err
	}

	g.Countersignature = hex.EncodeToString(receiver.Sign(message))
	return nil
}

// NewGrantResponseParams returns GrantResponseParams
func NewGrantResponseParams(id string, grant *GrantRequest, action OfferResponseAction) *GrantResponseParams {
	return &GrantResponseParams{
		ID:     id,
		Action: action,
		auth:   make(http.Header),
		record: &CountersignedGrantRequest{
			ShareID:     grant.ShareID,
			Quantity:    grant.Quantity,
			Owner:       grant.Owner,
			Recipient:   grant.Recipient,
			BeforeBlock: grant.BeforeBlock,
			Signature:   grant.Signature,
		},
	}
}

// Copy of bitmark share swap structure
type SwapRequest struct {
	ShareIDOne  string `json:"shareIDOne" pack:"hex32"`   // share = issue id
	QuantityOne uint64 `json:"quantityOne" pack:"uint64"` // shares to transfer > 0
	OwnerOne    string `json:"ownerOne" pack:"account"`   // base58
	ShareIDTwo  string `json:"shareIDTwo" pack:"hex32"`   // share = issue id
	QuantityTwo uint64 `json:"quantityTwo" pack:"uint64"` // shares to transfer > 0
	OwnerTwo    string `json:"ownerTwo" pack:"account"`   // base58
	BeforeBlock uint64 `json:"beforeBlock" pack:"uint64"` // expires when chain height > before block
	Signature   string `json:"signature"`                 // hex
}

// ShareSwapParams is the parameter for swaping shares between two accounts via core api
type ShareSwapParams struct {
	Swap *SwapRequest `json:"swap"`
}

// FromShare will assign the first share for swaping
func (p *ShareSwapParams) FromShare(shareID, owner string, quantity uint64) *ShareSwapParams {
	p.Swap.ShareIDOne = shareID
	p.Swap.OwnerOne = owner
	p.Swap.QuantityOne = quantity
	return p
}

// ToShare will assign the second share for swaping
func (p *ShareSwapParams) ToShare(shareID, owner string, quantity uint64) *ShareSwapParams {
	p.Swap.ShareIDTwo = shareID
	p.Swap.OwnerTwo = owner
	p.Swap.QuantityTwo = quantity
	return p
}

// Sign will generate the signature for a swaping request
func (p *ShareSwapParams) Sign(requester account.Account) error {
	message, err := utils.Pack(p.Swap)
	if err != nil {
		return err
	}
	p.Swap.Signature = hex.EncodeToString(requester.Sign(message))
	return nil
}

// NewShareSwapParams returns ShareSwapParams
func NewShareSwapParams(beforeBlock uint64) *ShareSwapParams {
	return &ShareSwapParams{
		Swap: &SwapRequest{
			BeforeBlock: beforeBlock,
		},
	}
}

// Copy of bitmark share swap structure with counter signature
type CountersignedSwapRequest struct {
	ShareIDOne       string `json:"shareIDOne" pack:"hex32"`   // share = issue id
	QuantityOne      uint64 `json:"quantityOne" pack:"uint64"` // shares to transfer > 0
	OwnerOne         string `json:"ownerOne" pack:"account"`   // base58
	ShareIDTwo       string `json:"shareIDTwo" pack:"hex32"`   // share = issue id
	QuantityTwo      uint64 `json:"quantityTwo" pack:"uint64"` // shares to transfer > 0
	OwnerTwo         string `json:"ownerTwo" pack:"account"`   // base58
	BeforeBlock      uint64 `json:"beforeBlock" pack:"uint64"` // expires when chain height > before block
	Signature        string `json:"signature" pack:"hex64"`
	Countersignature string `json:"countersignature"`
}

// ShareSwapParams is the parameter for responding swaping shares between two accounts via core api
type SwapResponseParams struct {
	ID               string              `json:"id"`
	Action           OfferResponseAction `json:"action"`
	Countersignature string              `json:"countersignature"`
	auth             http.Header
	record           *CountersignedSwapRequest
}

// Sign will generate the signature for a swaping responding request
func (s *SwapResponseParams) Sign(receiver account.Account) error {
	message, err := utils.Pack(s.record)
	if err != nil {
		return err
	}

	s.Countersignature = hex.EncodeToString(receiver.Sign(message))
	return nil
}

// NewSwapResponseParams returns SwapResponseParams
func NewSwapResponseParams(swap *SwapRequest, action OfferResponseAction) *SwapResponseParams {
	return &SwapResponseParams{
		Action: action,
		auth:   make(http.Header),
		record: &CountersignedSwapRequest{
			ShareIDOne:  swap.ShareIDOne,
			QuantityOne: swap.QuantityOne,
			OwnerOne:    swap.OwnerOne,
			ShareIDTwo:  swap.ShareIDTwo,
			QuantityTwo: swap.QuantityTwo,
			OwnerTwo:    swap.OwnerTwo,
			BeforeBlock: swap.BeforeBlock,
			Signature:   swap.Signature,
		},
	}
}

type OfferParams struct {
	Offer struct {
		Transfer  *TransferRequest       `json:"record"`
		ExtraInfo map[string]interface{} `json:"extra_info"`
	} `json:"offer"`
}

func NewOfferParams(receiver string, info map[string]interface{}) (*OfferParams, error) {
	if err := account.ValidateAccountNumber(receiver); err != nil {
		return nil, err
	}

	return &OfferParams{
		Offer: struct {
			Transfer  *TransferRequest       `json:"record"`
			ExtraInfo map[string]interface{} `json:"extra_info"`
		}{
			Transfer: &TransferRequest{
				Owner:                   receiver,
				requireCountersignature: true,
			},
			ExtraInfo: info,
		},
	}, nil
}

// FromBitmark sets link asynchronously
func (o *OfferParams) FromBitmark(bitmarkID string) error {
	bitmark, err := Get(bitmarkID)
	if err != nil {
		return err
	}

	o.Offer.Transfer.Link = bitmark.LatestTxID
	return nil
}

// FromLatestTx sets link synchronously
func (o *OfferParams) FromLatestTx(txID string) {
	o.Offer.Transfer.Link = txID
}

func (o *OfferParams) Sign(sender account.Account) error {
	message, err := utils.Pack(o.Offer.Transfer)
	if err != nil {
		return err
	}
	o.Offer.Transfer.Signature = hex.EncodeToString(sender.Sign(message))
	return nil
}

type CountersignedTransferRequest struct {
	Link             string   `json:"link" pack:"hex32"`
	Escrow           *payment `json:"-" pack:"payment"` // optional escrow payment address
	Owner            string   `json:"owner" pack:"account"`
	Signature        string   `json:"signature" pack:"hex64"`
	Countersignature string   `json:"countersignature"`
}

type ResponseParams struct {
	ID               string              `json:"id"`
	Action           OfferResponseAction `json:"action"`
	Countersignature string              `json:"countersignature"`
	auth             http.Header
	record           *CountersignedTransferRequest
}

func NewTransferResponseParams(bitmark *Bitmark, action OfferResponseAction) *ResponseParams {
	return &ResponseParams{
		ID:     bitmark.Offer.ID,
		Action: action,
		auth:   make(http.Header),
		record: bitmark.Offer.Record,
	}
}

func (r *ResponseParams) Sign(acct account.Account) error {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		"updateOffer",
		r.ID,
		acct.AccountNumber(),
		ts,
	}
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.Sign([]byte(message)))

	r.auth.Add("requester", acct.AccountNumber())
	r.auth.Add("timestamp", ts)
	r.auth.Add("signature", sig)

	if r.Action == Accept {
		message, err := utils.Pack(r.record)
		if err != nil {
			return err
		}
		r.Countersignature = hex.EncodeToString(acct.Sign(message))
	}
	return nil
}
