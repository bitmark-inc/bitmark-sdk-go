// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package bitmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

type txItem struct {
	TxID string `json:"txID"`
}

func Issue(params *IssuanceParams) ([]string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return nil, err
	}

	req, err := client.NewRequest("POST", "/v3/issue", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Bitmarks []struct {
			ID string `json:"id"`
		} `json:"bitmarks"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	bitmarkIDs := make([]string, 0)
	for _, item := range result.Bitmarks {
		bitmarkIDs = append(bitmarkIDs, item.ID)
	}

	return bitmarkIDs, nil
}

func Transfer(params *TransferParams) (string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", err
	}

	req, err := client.NewRequest("POST", "/v3/transfer", body)
	if err != nil {
		return "", err
	}

	var result txItem
	if err := client.Do(req, &result); err != nil {
		return "", err
	}

	return result.TxID, nil
}

func Offer(params *OfferParams) error {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return err
	}

	req, err := client.NewRequest("POST", "/v3/transfer", body)
	if err != nil {
		return err
	}

	err = client.Do(req, nil)
	return err
}

func Respond(params *ResponseParams) (string, error) {
	if params.auth.Get("signature") == "" {
		return "", errors.New("response not signed")
	}

	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", err
	}

	req, err := client.NewRequest("PATCH", "/v3/transfer", body)
	if err != nil {
		return "", err
	}
	// TODO: set signature beautifully
	for k, v := range params.auth {
		req.Header.Add(k, v[0])
	}

	var result txItem

	if err := client.Do(req, &result); err != nil {
		return "", err
	}

	return result.TxID, nil
}

func CreateShares(params *ShareParams) (string, string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", "", err
	}

	req, err := client.NewRequest("POST", "/v3/shares", body)
	if err != nil {
		return "", "", err
	}

	var result struct {
		TxID    string `json:"tx_id"`
		ShareID string `json:"share_id"`
	}
	if err := client.Do(req, &result); err != nil {
		return "", "", err
	}

	return result.TxID, result.ShareID, nil
}

func GrantShare(params *ShareGrantingParams) (string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", err
	}

	req, err := client.NewRequest("POST", "/v3/share-offer", body)
	if err != nil {
		return "", err
	}

	var result struct {
		OfferID string `json:"offer_id"`
	}
	err = client.Do(req, &result)
	return result.OfferID, err
}

func ReplyShareOffer(params *GrantResponseParams) (string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", err
	}

	req, err := client.NewRequest("PATCH", "/v3/share-offer", body)
	if err != nil {
		return "", err
	}
	// TODO: set signaure beautifully
	for k, v := range params.auth {
		req.Header.Add(k, v[0])
	}

	var result txItem
	err = client.Do(req, &result)
	return result.TxID, err
}

func Get(bitmarkID string) (*Bitmark, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/bitmarks/%s?%s", bitmarkID, vals.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Bitmark *Bitmark `json:"bitmark"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Bitmark, nil
}

func GetWithAsset(bitmarkID string) (*Bitmark, *asset.Asset, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")
	vals.Set("asset", "true")

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/bitmarks/%s?%s", bitmarkID, vals.Encode()), nil)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Bitmark *Bitmark     `json:"bitmark"`
		Asset   *asset.Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, nil, err
	}

	return result.Bitmark, result.Asset, nil
}

func List(builder *QueryParamsBuilder) ([]*Bitmark, []*asset.Asset, error) {
	params, err := builder.Build()

	if err != nil {
		return nil, nil, err
	}

	client := sdk.GetAPIClient()
	req, err := client.NewRequest("GET", "/v3/bitmarks?"+params, nil)

	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Bitmarks []*Bitmark     `json:"bitmarks"`
		Assets   []*asset.Asset `json:"assets"`
	}

	if err := client.Do(req, &result); err != nil {
		return nil, nil, err
	}

	return result.Bitmarks, result.Assets, nil
}

func GetShareBalance(shareID, owner string) (*Share, error) {
	client := sdk.GetAPIClient()

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/shares?share_id=%s&owner=%s", shareID, owner), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Shares []*Share `json:"shares"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}
	if len(result.Shares) == 0 {
		return nil, nil
	}

	return result.Shares[0], nil
}

func ListShareOffers(from, to string) ([]*ShareOffer, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	if from != "" {
		vals.Set("from", from)
	}
	if to != "" {
		vals.Set("to", to)
	}

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/share-offer?%s", vals.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Offers []*ShareOffer `json:"offers"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Offers, nil
}

type QueryParamsBuilder struct {
	params url.Values
	err    error
}

func NewQueryParamsBuilder() *QueryParamsBuilder {
	return &QueryParamsBuilder{params: url.Values{}}
}

func (ub *QueryParamsBuilder) OwnedBy(owner string) *QueryParamsBuilder {
	ub.params.Set("owner", owner)
	return ub
}

func (ub *QueryParamsBuilder) OwnedByWithTransient(owner string) *QueryParamsBuilder {
	ub.params.Set("owner", owner)
	ub.params.Set("sent", strconv.FormatBool(true))
	return ub
}

func (ub *QueryParamsBuilder) IssuedBy(issuer string) *QueryParamsBuilder {
	ub.params.Set("issuer", issuer)
	return ub
}

func (ub *QueryParamsBuilder) Pending(pending bool) *QueryParamsBuilder {
	ub.params.Set("pending", strconv.FormatBool(pending))
	return ub
}

func (ub *QueryParamsBuilder) OfferFrom(sender string) *QueryParamsBuilder {
	ub.params.Set("offer_from", sender)
	return ub
}

func (ub *QueryParamsBuilder) OfferTo(receiver string) *QueryParamsBuilder {
	ub.params.Set("offer_to", receiver)
	return ub
}

func (ub *QueryParamsBuilder) BitmarkIDs(bitmarkIDs []string) *QueryParamsBuilder {
	for _, bitmarkID := range bitmarkIDs {
		ub.params.Add("bitmark_ids", bitmarkID)
	}
	return ub
}

func (ub *QueryParamsBuilder) ReferencedAsset(assetID string) *QueryParamsBuilder {
	ub.params.Set("asset_id", assetID)
	return ub
}

func (ub *QueryParamsBuilder) LoadAsset(load bool) *QueryParamsBuilder {
	ub.params.Set("asset", strconv.FormatBool(load))
	return ub
}

func (ub *QueryParamsBuilder) Limit(size int) *QueryParamsBuilder {
	if size > 100 {
		ub.err = errors.New("invalid size: max = 100")
	}
	ub.params.Set("limit", strconv.Itoa(size))
	return ub
}

func (ub *QueryParamsBuilder) At(at int) *QueryParamsBuilder {
	ub.params.Set("at", strconv.Itoa(at))
	return ub
}

func (ub *QueryParamsBuilder) To(direction utils.Direction) *QueryParamsBuilder {
	if direction != "" && (direction != utils.Later && direction != utils.Earlier) {
		ub.err = errors.New("it must be 'later' or 'earlier'")
	}

	ub.params.Set("to", string(direction))
	return ub
}

func (ub *QueryParamsBuilder) Build() (string, error) {
	if ub.err != nil {
		return "", ub.err
	}

	if ub.params.Get("pending") == "" {
		ub.params.Set("pending", "true")
	}

	return ub.params.Encode(), nil
}
