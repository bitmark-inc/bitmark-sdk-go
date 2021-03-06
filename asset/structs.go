// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package asset

import "time"

type Asset struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Metadata    map[string]string `json:"metadata"`
	Fingerprint string            `json:"fingerprint"`
	Registrant  string            `json:"registrant"`
	Status      string            `json:"status"`
	BlockNumber int               `json:"block_number"`
	Offset      int               `json:"offset"`
	CreatedAt   time.Time         `json:"created_at"`
}
