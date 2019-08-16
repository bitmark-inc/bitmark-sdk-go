// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package tx

type Tx struct {
	ID            string                 `json:"id"`
	Owner         string                 `json:"owner"`
	PreviousID    string                 `json:"previous_id"`
	PreviousOwner string                 `json:"previous_owner"`
	BitmarkID     string                 `json:"bitmark_id"`
	AssetID       string                 `json:"asset_id"`
	Countersign   bool                   `json:"countersign"`
	Status        string                 `json:"status"`
	BlockNumber   int                    `json:"block_number"`
	Confirmation  uint64                 `json:"confirmation"`
	ShareInfo     map[string]interface{} `json:"share_info,omitempty"`
	Sequence      int                    `json:"offset"`
}
