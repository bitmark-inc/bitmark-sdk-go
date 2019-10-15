// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"os"
	"testing"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/stretchr/testify/suite"
)

type GiveawayTestSuite struct {
	BaseTestSuite
}

func NewGiveawayTestSuite(bitmarkCount int) *GiveawayTestSuite {
	s := new(GiveawayTestSuite)
	s.bitmarkCount = bitmarkCount
	return s
}

func (s *GiveawayTestSuite) SetupSuite() {
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
}

func (s *GiveawayTestSuite) TestDirectTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustDirectTransfer(bitmarkID) // able to transfer right after the bitmark is issued
	s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "issuing", 5*time.Second)

	for {
		if txsAreReady([]string{bitmarkID}) {
			break
		}
		time.Sleep(15 * time.Second)
	}
	s.verifyBitmark(bitmarkID, s.receiver.AccountNumber(), "transferring", 5*time.Second)
}

func (s *GiveawayTestSuite) TestCountersignedTransfer() {
	bitmarkID := s.bitmarkIDs[s.bitmarkIndex]
	s.T().Logf("bitmark_id=%s", bitmarkID)

	s.mustCreateOffer(bitmarkID) // able to create a transfer offer right after the bitmark is issued
	s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "issuing", 5*time.Second)

	for {
		if txsAreReady([]string{bitmarkID}) {
			break
		}
		time.Sleep(15 * time.Second)
	}
	bmk := s.verifyBitmark(bitmarkID, s.sender.AccountNumber(), "offering", 0)

	params := bitmark.NewTransferResponseParams(bmk, "accept")
	params.Sign(s.receiver)
	_, err := bitmark.Respond(params)
	s.NoError(err)

	s.verifyBitmark(bitmarkID, s.receiver.AccountNumber(), "transferring", 5*time.Second)
}

func TestGiveawayTestSuite(t *testing.T) {
	suite.Run(t, NewGiveawayTestSuite(2))
}
