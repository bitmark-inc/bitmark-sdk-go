// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package utils

import (
	"encoding/hex"
	"fmt"
	"reflect"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/encoding"
)

const (
	tagName = "pack"

	tagRegister              = uint64(2)
	tagIssue                 = uint64(3)
	tagDirectTransfer        = uint64(4)
	tagCountersignedTransfer = uint64(5)
	tagShare                 = uint64(8)
	tagShareGrant            = uint64(9)
	tagShareSwap             = uint64(10)
)

type Direction string

const (
	Earlier = Direction("earlier")
	Later   = Direction("later")
)

func Pack(params interface{}) ([]byte, error) {
	var buffer []byte

	v := reflect.ValueOf(params).Elem()
	switch v.Type().String() {
	case "asset.RegistrationParams":
		buffer = encoding.ToVarint64(tagRegister)
	case "bitmark.IssueRequest":
		buffer = encoding.ToVarint64(tagIssue)
	case "bitmark.TransferRequest":
		requireCountersigned := v.Field(v.NumField() - 1).Bool()
		buffer = encoding.ToVarint64(tagDirectTransfer)
		if requireCountersigned {
			buffer = encoding.ToVarint64(tagCountersignedTransfer)
		}
	case "bitmark.CountersignedTransferRequest":
		buffer = encoding.ToVarint64(tagCountersignedTransfer)
	case "bitmark.ShareRequest":
		buffer = encoding.ToVarint64(tagShare)
	case "bitmark.GrantRequest", "bitmark.CountersignedGrantRequest":
		buffer = encoding.ToVarint64(tagShareGrant)
	case "bitmark.SwapRequest", "bitmark.CountersignedSwapRequest":
		buffer = encoding.ToVarint64(tagShareSwap)
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get(tagName)
		switch tag {
		case "utf8":
			value := v.Field(i).String()
			buffer = appendString(buffer, value)
		case "hex64":
			value := v.Field(i).String()
			bytes, err := hex.DecodeString(value)
			if err != nil || len(bytes) != 64 {
				return nil, fmt.Errorf("invalid %s", v.Type().Field(i).Name)
			}
			buffer = appendBytes(buffer, bytes)
		case "hex32":
			value := v.Field(i).String()
			bytes, err := hex.DecodeString(value)
			if err != nil || len(bytes) != 32 {
				return nil, fmt.Errorf("invalid %s", v.Type().Field(i).Name)
			}
			buffer = appendBytes(buffer, bytes)
		case "account":
			value := v.Field(i).String()
			bytes := encoding.FromBase58(value)
			if len(bytes) != account.Base58AccountNumberLength {
				return nil, fmt.Errorf("invalid %s", v.Type().Field(i).Name)
			}
			buffer = appendBytes(buffer, bytes[:len(bytes)-account.ChecksumLength])
		case "payment": // TODO: support escrow
			buffer = append(buffer, 0)
		case "uint64":
			value := v.Field(i).Uint()
			buffer = appendUint64(buffer, value)
		}
	}

	return buffer, nil
}

func appendString(buffer []byte, s string) []byte {
	l := encoding.ToVarint64(uint64(len(s)))
	buffer = append(buffer, l...)
	return append(buffer, s...)
}

func appendBytes(buffer []byte, data []byte) []byte {
	l := encoding.ToVarint64(uint64(len(data)))
	buffer = append(buffer, l...)
	buffer = append(buffer, data...)
	return buffer
}

func appendUint64(buffer []byte, value uint64) []byte {
	valueBytes := encoding.ToVarint64(value)
	buffer = append(buffer, valueBytes...)
	return buffer
}
