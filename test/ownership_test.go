// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type OwnershipTestSuite struct {
	BaseTestSuite
}

func NewOwnershipTestSuite(bitmarkCount int) *OwnershipTestSuite {
	s := new(OwnershipTestSuite)
	s.bitmarkCount = bitmarkCount
	return s
}

func (s *OwnershipTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	for {
		if txsAreReady(s.bitmarkIds) {
			break
		}
		time.Sleep(15 * time.Second)
	}
}

func (s *OwnershipTestSuite) TestDirectTransfer() {
	bitmarkId := s.bitmarkIds[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkId)

	s.mustDirectTransfer(bitmarkId)
	s.verifyBitmark(bitmarkId, s.receiver.AccountNumber(), "transferring", 10*time.Second)
}

func (s *OwnershipTestSuite) TestCreateAndCancelCountersignedTransfer() {
	bitmarkId := s.bitmarkIds[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkId)

	s.mustCreateOffer(bitmarkId)

	bmk := s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// cancelled not by sender
	params := bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(s.receiver)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2014] message: not transfer offer sender reason: not authorized requester")

	bmk = s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// cancelled by sender
	params = bitmark.NewTransferResponseParams(bmk, "cancel")
	params.Sign(s.sender)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "settled", 0)
}

func (s *OwnershipTestSuite) TestCreateAndRejectCountersignedTransfer() {
	bitmarkId := s.bitmarkIds[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkId)

	s.mustCreateOffer(bitmarkId)

	bmk := s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// rejected not by receiver
	params := bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(s.sender)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2015] message: not transfer offer receiver reason: not authorized requester")

	bmk = s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// rejected by receiver
	params = bitmark.NewTransferResponseParams(bmk, "reject")
	params.Sign(s.receiver)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "settled", 0)
}

func (s *OwnershipTestSuite) TestCreateAndAcceptCountersignedTransfer() {
	bitmarkId := s.bitmarkIds[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkId)

	s.mustCreateOffer(bitmarkId)

	bmk := s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// accepted not by receiver
	params := bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(s.sender)
	_, err := bitmark.Respond(params)
	s.EqualError(err, "[2015] message: not transfer offer receiver reason: invalid transfer offer request because of error: only the receiptant can accept a transfer offer")

	bmk = s.verifyBitmark(bitmarkId, s.sender.AccountNumber(), "offering", 0)

	// accepted by receiver
	params = bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(s.receiver)
	_, err = bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkId, s.receiver.AccountNumber(), "transferring", 10*time.Second)
}

func TestOwnershipTestSuite(t *testing.T) {
	suite.Run(t, NewOwnershipTestSuite(4))
}
