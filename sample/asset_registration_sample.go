// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"io/ioutil"
)

func registerAsset(registrant account.Account, assetName string, assetFilePath string, metadata map[string]string) (string, error) {
	assetFileContent, err := ioutil.ReadFile(assetFilePath)

	if err != nil {
		fmt.Println("Can not read file!")
		return "", err
	}

	registrationParams, _ := asset.NewRegistrationParams(assetName, metadata)
	registrationParams.SetFingerprintFromData([]byte(assetFileContent))
	registrationParams.Sign(registrant)

	assetID, err := asset.Register(registrationParams)

	return assetID, err
}
