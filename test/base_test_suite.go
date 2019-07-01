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
	bitmarkIds   []string
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

	assetId := s.mustRegisterAsset("", []byte(time.Now().String()))
	s.bitmarkIds = s.mustIssueBitmarks(assetId, s.bitmarkCount)
}

func (s *BaseTestSuite) TearDownTest() {
	s.bitmarkIndex++
}

func (s *BaseTestSuite) mustRegisterAsset(name string, content []byte) string {
	params, _ := asset.NewRegistrationParams(name, nil)
	params.SetFingerprint(content)
	params.Sign(s.sender)
	assetId, err := asset.Register(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}

	return assetId
}

func (s *BaseTestSuite) mustIssueBitmarks(assetId string, quantity int) []string {
	params := bitmark.NewIssuanceParams(assetId, quantity)
	params.Sign(s.sender)
	bitmarkIds, err := bitmark.Issue(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}
	return bitmarkIds
}

func (s *BaseTestSuite) mustDirectTransfer(bitmarkId string) {
	params, err := bitmark.NewTransferParams(s.receiver.AccountNumber())
	if !s.NoError(err) {
		s.T().FailNow()
	}
	params.FromBitmark(bitmarkId)
	params.Sign(s.sender)
	_, err = bitmark.Transfer(params)
	if !s.NoError(err) {
		s.T().FailNow()
	}
}

func (s *BaseTestSuite) mustCreateOffer(bitmarkId string) {
	params, err := bitmark.NewOfferParams(s.receiver.AccountNumber(), nil)
	if !s.NoError(err) {
		s.T().Fatal(err)
	}

	params.FromBitmark(bitmarkId)
	params.Sign(s.sender)
	if !s.NoError(bitmark.Offer(params)) {
		s.T().Fatal(err)
	}
}

func (s *BaseTestSuite) verifyBitmark(bitmarkId, owner, status string, delay time.Duration) *bitmark.Bitmark {
	time.Sleep(delay)

	bmk, err := bitmark.Get(bitmarkId)
	if !s.NoError(err) || !s.Equal(owner, bmk.Owner) || !s.Equal(status, bmk.Status) {
		s.T().Logf("bitmark: %+v", bmk)
		s.T().FailNow()
	}
	return bmk
}
