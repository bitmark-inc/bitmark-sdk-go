// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2020 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
)

type OwnershipTestSuite struct {
	BaseTestSuite

	bitmarkCount int
	bitmarkIDs   []string
}

func NewOwnershipTestSuite(bitmarkCount int) *OwnershipTestSuite {
	s := new(OwnershipTestSuite)
	s.bitmarkCount = bitmarkCount
	return s
}

func (s *OwnershipTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	var err error
	s.sender, err = account.FromSeed(os.Getenv("SENDER_SEED"))
	if err != nil {
		s.Fail(err.Error())
	}
	s.receiver, err = account.FromSeed(os.Getenv("RECEIVER_SEED"))
	if err != nil {
		s.Fail(err.Error())
	}

	assetID := s.mustRegisterAsset("", []byte(time.Now().String()))
	s.bitmarkIDs = s.mustIssueBitmarks(assetID, s.bitmarkCount)

loop:
	for {
		if s.txsAreReady(s.bitmarkIDs) {
			break loop
		}
		time.Sleep(15 * time.Second)
	}
}

func (s *OwnershipTestSuite) TestDirectTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustDirectTransfer(bitmarkID)
	s.verifyBitmark(bitmarkID, s.receiver.AccountNumber(), "transferring")
}

func (s *OwnershipTestSuite) TestCreateAndCancelCountersignedTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustCreateOffer(bitmarkID)

	bmk := s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// cancelled not by sender
	params := bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(s.receiver)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2014] message: not transfer offer sender reason: not authorized requester")

	bmk = s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// cancelled by sender
	params = bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(s.sender)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "settled")
}

func (s *OwnershipTestSuite) TestCreateAndRejectCountersignedTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustCreateOffer(bitmarkID)

	bmk := s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// rejected not by receiver
	params := bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(s.sender)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2015] message: not transfer offer receiver reason: not authorized requester")

	bmk = s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// rejected by receiver
	params = bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(s.receiver)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "settled")
}

func (s *OwnershipTestSuite) TestCreateAndAcceptCountersignedTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustCreateOffer(bitmarkID)

	bmk := s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// accepted not by receiver
	params := bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(s.sender)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2015] message: not transfer offer receiver reason: invalid transfer offer request because of error: only the recipient can accept a transfer offer")

	bmk = s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering")

	// accepted by receiver
	params = bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(s.receiver)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkID, s.receiver.AccountNumber(), "transferring")
}

func (s *OwnershipTestSuite) TestRegisterDuplicateAsset() {
	params, _ := asset.NewRegistrationParams("another name", nil)
	params.SetFingerprintFromData([]byte("Fri May 10 14:01:41 CST 2019"))
	params.Sign(s.sender)
	_, err := asset.Register(params)
	s.Error(err)
}

func (s *OwnershipTestSuite) TestIssueForNotExistingAsset() {
	notExistingAssetID := "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"
	params, _ := bitmark.NewIssuanceParams(notExistingAssetID, 1)
	params.Sign(s.sender)
	_, err := bitmark.Issue(params)
	s.Error(err)
}

func TestOwnershipTestSuite(t *testing.T) {
	suite.Run(t, NewOwnershipTestSuite(4))
}
