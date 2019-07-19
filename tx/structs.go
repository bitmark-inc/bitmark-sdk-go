// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package tx

import (
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
)

type Tx struct {
	Id           string `json:"id"`
	BitmarkId    string `json:"bitmark_id"`
	AssetId      string `json:"asset_id"`
	Asset        *asset.Asset
	Owner        string `json:"owner"`
	Status       string `json:"status"`
	BlockNumber  int    `json:"block_number"`
	Sequence     int    `json:"offset"`
	PreviousId   string `json:"previous_id"`
	Confirmation uint64 `json:"confirmation"`

	ShareInfo map[string]interface{} `json:"share_info,omitempty"`
}
