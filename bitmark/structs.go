// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package bitmark

import (
	"encoding/json"
	"time"
)

type Bitmark struct {
	ID          string         `json:"id"`
	AssetID     string         `json:"asset_id"`
	LatestTxID  string         `json:"head_id"`
	Issuer      string         `json:"issuer"`
	Owner       string         `json:"owner"`
	Status      string         `json:"status"`
	Offer       *TransferOffer `json:"offer"`
	BlockNumber int            `json:"block_number"`
	Edition     int            `json:"edition"`
	Offset      int            `json:"offset"`
	CreatedAt   time.Time      `json:"created_at"`
	ConfirmedAt time.Time      `json:"confirmed_at"`
}

type TransferOffer struct {
	ID        string                        `json:"id"`
	From      string                        `json:"from"`
	To        string                        `json:"to"`
	Record    *CountersignedTransferRequest `json:"record"`
	ExtraInfo map[string]string             `json:"extra_info"`
	CreatedAt time.Time                     `json:"created_at"`
}

type Share struct {
	ID        string `json:"share_id"`
	Owner     string `json:"owner"`
	Balance   uint64 `json:"balance"`
	Available uint64 `json:"available"`
}

type ShareOffer struct {
	ID        string          `json:"id"`
	ShareID   string          `json:"share_id"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Record    GrantRequest    `json:"record"`
	ExtraInfo json.RawMessage `json:"extra_info"`
	CreatedAt time.Time       `json:"created_at"`
}
