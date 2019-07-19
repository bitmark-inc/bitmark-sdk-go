// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AssetTestSuite struct {
	BaseTestSuite
}

func (s *AssetTestSuite) TestRegisterExistingAsset() {
	params, _ := asset.NewRegistrationParams("another name", nil)
	params.SetFingerprintFromData([]byte("Fri May 10 14:01:41 CST 2019"))
	params.Sign(s.sender)
	_, err := asset.Register(params)
	s.Error(err)
}

func (s *AssetTestSuite) TestIssueBitmarksForNotExistingAsset() {
	notExistingAssetId := "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	params := bitmark.NewIssuanceParams(notExistingAssetId, 1)
	params.Sign(s.sender)
	_, err := bitmark.Issue(params)
	s.Error(err)
}

func TestAssetTestSuite(t *testing.T) {
	suite.Run(t, new(AssetTestSuite))
}
