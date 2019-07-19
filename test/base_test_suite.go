package test

import (
	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/stretchr/testify/suite"
	"net/http"
	"os"
	"time"
)

type BaseTestSuite struct {
	suite.Suite

	sender   account.Account
	receiver account.Account

	bitmarkIndex int
	bitmarkCount int
	bitmarkIDs   []string
}

func (s *BaseTestSuite) SetupSuite() {
	network := os.Getenv("SDK_TEST_NETWORK")
	token := os.Getenv("SDK_TEST_API_TOKEN")
	sdk.Init(&sdk.Config{
		HTTPClient: http.DefaultClient,
		Network:    sdk.Network(network),
		APIToken:   token,
	})

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

func (s *BaseTestSuite) TearDownTest() {
	s.bitmarkIndex++
}

func (s *BaseTestSuite) mustRegisterAsset(name string, content []byte) string {
	params, _ := asset.NewRegistrationParams(name, nil)
	params.SetFingerprintFromData(content)
	params.Sign(s.sender)
	assetID, err := asset.Register(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}

	return assetID
}

func (s *BaseTestSuite) mustIssueBitmarks(assetID string, quantity int) []string {
	params := bitmark.NewIssuanceParams(assetID, quantity)
	params.Sign(s.sender)
	bitmarkIDs, err := bitmark.Issue(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	return bitmarkIDs
}

func (s *BaseTestSuite) mustDirectTransfer(bitmarkID string) {
	params, err := bitmark.NewTransferParams(s.receiver.AccountNumber())
	if !s.NoError(err) {
		s.T().FailNow()
	}
	params.FromBitmark(bitmarkID)
	params.Sign(s.sender)
	_, err = bitmark.Transfer(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}
}

func (s *BaseTestSuite) mustCreateOffer(bitmarkID string) {
	params, err := bitmark.NewOfferParams(s.receiver.AccountNumber(), nil)
	if !s.NoError(err) {
		s.T().Fatal(err)
	}

	params.FromBitmark(bitmarkID)
	params.Sign(s.sender)
	if !s.NoError(bitmark.Offer(params)) {
		s.T().Fatal(err)
	}
}

func (s *BaseTestSuite) verifyBitmark(bitmarkID, owner, status string, delay time.Duration) *bitmark.Bitmark {
	time.Sleep(delay)

	bmk, err := bitmark.Get(bitmarkID)
	if !s.NoError(err) || !s.Equal(owner, bmk.Owner) || !s.Equal(status, bmk.Status) {
		s.T().Logf("bitmark: %+v", bmk)
		s.T().FailNow()
	}
	return bmk
}
