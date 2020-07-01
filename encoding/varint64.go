// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package encoding

const varint64MaximumBytes = 9

func ToVarint64(value uint64) []byte {
	result := make([]byte, 0, varint64MaximumBytes)
	if value < 0x80 {
		result = append(result, byte(value))
		return result
	}
	for i := 0; i < varint64MaximumBytes && value != 0; i++ {
		ext := uint64(0x80)
		if value < 0x80 {
			ext = 0x00
		}
		result = append(result, byte(value|ext))
		value >>= 7
	}
	return result
}
