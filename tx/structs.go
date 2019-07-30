// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package tx

type Tx struct {
	ID           string `json:"id"`
	BitmarkID    string `json:"bitmark_id"`
	AssetID      string `json:"asset_id"`
	Owner        string `json:"owner"`
	Status       string `json:"status"`
	BlockNumber  int    `json:"block_number"`
	Sequence     int    `json:"offset"`
	PreviousID   string `json:"previous_id"`
	Confirmation uint64 `json:"confirmation"`

	ShareInfo map[string]interface{} `json:"share_info,omitempty"`
}
