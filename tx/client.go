// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package tx

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

func Get(txID string) (*Tx, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/txs/%s?%s", txID, vals.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tx *Tx `json:"tx"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Tx, nil
}

func GetWithAsset(txID string) (*Tx, *asset.Asset, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")
	vals.Set("asset", "true")

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/txs/%s?%s", txID, vals.Encode()), nil)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Tx    *Tx          `json:"tx"`
		Asset *asset.Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, nil, err
	}

	return result.Tx, result.Asset, nil
}

func List(builder *QueryParamsBuilder) ([]*Tx, []*asset.Asset, error) {
	params, err := builder.Build()

	if err != nil {
		return nil, nil, err
	}

	client := sdk.GetAPIClient()
	req, err := client.NewRequest("GET", "/v3/txs?"+params, nil)

	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Txs    []*Tx          `json:"txs"`
		Assets []*asset.Asset `json:"assets"`
	}

	if err := client.Do(req, &result); err != nil {
		return nil, nil, err
	}

	return result.Txs, result.Assets, nil
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
	ub.params.Set("sent", "true")
	return ub
}

func (ub *QueryParamsBuilder) ReferencedBitmark(bitmarkID string) *QueryParamsBuilder {
	ub.params.Set("bitmark_id", bitmarkID)
	return ub
}

func (ub *QueryParamsBuilder) ReferencedBlockNumber(blockNumber int64) *QueryParamsBuilder {
	ub.params.Set("block_number", fmt.Sprintf("%d", blockNumber))
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

func (ub *QueryParamsBuilder) Pending(pending bool) *QueryParamsBuilder {
	ub.params.Set("pending", strconv.FormatBool(pending))
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
