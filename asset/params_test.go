// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package asset

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

var (
	seed = "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH"

	registrant account.Account
)

func init() {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	registrant, _ = account.FromSeed(seed)
}

func TestNewRegistrantionParams(t *testing.T) {
	params, err := NewRegistrationParams(
		"name",
		map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "",
			"":   "v4",
		})
	assert.NoError(t, err)
	assert.Equal(t, params.Metadata, "k1\u0000v1\u0000k2\u0000v2")

	_, err = NewRegistrationParams(strings.Repeat("X", 65), nil)
	assert.EqualError(t, err, ErrInvalidNameLength.Error())

	_, err = NewRegistrationParams("", map[string]string{strings.Repeat("X", 1024): strings.Repeat("Y", 1025)})
	assert.EqualError(t, err, ErrInvalidMetadataLength.Error())
}

func TestSetFingerprintFromData(t *testing.T) {
	params, err := NewRegistrationParams("", nil)
	assert.NoError(t, err)

	err = params.SetFingerprintFromData([]byte("hello world"))
	assert.NoError(t, err)
	assert.Equal(t, params.Fingerprint, "01840006653e9ac9e95117a15c915caab81662918e925de9e004f774ff82d7079a40d4d27b1b372657c61d46d470304c88c788b3a4527ad074d1dccbee5dbaa99a")

	err = params.SetFingerprintFromData(nil)
	assert.Error(t, err, ErrEmptyContent.Error())
}

func TestSetFingerprintFromReader(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))

	params, err := NewRegistrationParams("", nil)
	assert.NoError(t, err)

	err = params.SetFingerprintFromReader(r)
	assert.NoError(t, err)
	assert.Equal(t, params.Fingerprint, "01840006653e9ac9e95117a15c915caab81662918e925de9e004f774ff82d7079a40d4d27b1b372657c61d46d470304c88c788b3a4527ad074d1dccbee5dbaa99a")

	err = params.SetFingerprintFromData(nil)
	assert.Error(t, err, ErrEmptyContent.Error())
}

func TestSetFingerprint(t *testing.T) {
	params, err := NewRegistrationParams("", nil)
	assert.NoError(t, err)

	err = params.SetFingerprint("hello world")
	assert.NoError(t, err)
	assert.Equal(t, params.Fingerprint, "00hello world")

	err = params.SetFingerprint("")
	assert.Error(t, err, ErrEmptyContent.Error())
}

func TestSign(t *testing.T) {
	params, err := NewRegistrationParams(
		"name",
		map[string]string{
			"k1": "v1",
			"k2": "v2",
			"k3": "",
			"":   "v4",
		})
	assert.NoError(t, err)
	assert.NoError(t, params.SetFingerprintFromData([]byte("hello world")))
	assert.NoError(t, params.Sign(registrant))
	assert.Equal(t, params, &RegistrationParams{
		Name:        "name",
		Metadata:    "k1\u0000v1\u0000k2\u0000v2",
		Fingerprint: "01840006653e9ac9e95117a15c915caab81662918e925de9e004f774ff82d7079a40d4d27b1b372657c61d46d470304c88c788b3a4527ad074d1dccbee5dbaa99a",
		Registrant:  "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
		Signature:   "dc9ad2f4948d5f5defaf9043098cd2f3c245b092f0d0c2fc9744fab1835cfb1ad533ee0ff2a72d1cdd7a69f8ba6e95013fc517d5d4a16ca1b0036b1f3055270c",
	})

	assert.Error(t, params.Sign(nil), ErrNullRegistrant)
}
