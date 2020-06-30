// SPDX-License-Identifier: ISC
// Copyright (c) 2014-2019 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package test

import (
	"testing"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
	"github.com/stretchr/testify/suite"
)

type QueryTestSuite struct {
	BaseTestSuite
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (q *QueryTestSuite) TestGetAsset() {
	actual, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22")
	q.NoError(err)

	createdAt, _ := time.Parse("2006-01-02T15:04:05.000000Z", "2018-09-07T07:46:25.000000Z")
	expected := &asset.Asset{
		ID:   "2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22",
		Name: "HA25124377",
		Metadata: map[string]string{
			"Source":     "Bitmark Health",
			"Saved Time": "2018-09-07T07:45:41.948Z",
		},
		Fingerprint: "016ef802c0f912ed69a5afc0e6c08fbe96de3284e7cc6e685111d5f1705049f20b695443bc2d7bae7fe2091d9e7a880a50a51c2d0be1963a99b9914f60f2462040",
		Registrant:  "eTicVBQqmGzxNMGiZGtKzDdufXZsiFKH3SR8FcVYM7MQTZ47k3",
		Status:      "confirmed",
		BlockNumber: 8696,
		Offset:      8581,
		CreatedAt:   createdAt,
	}
	q.Equal(actual, expected)
}

func (q *QueryTestSuite) TestGetNonExistingAsset() {
	_, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d55")
	q.EqualError(err, "[4000] message: asset not found reason: ")
}

func (q *QueryTestSuite) TestListAssetByAtAndTo() {
	limit := 10
	at := 5

	builder := asset.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(limit).Pending(true)

	assets, err := asset.List(builder)
	q.NoError(err)
	q.True(len(assets) <= limit)
	for _, a := range assets {
		q.True(a.Offset <= at)
	}
}

func (q *QueryTestSuite) TestListAssetByAssetIDs() {
	existingAssetIDs := []string{
		"c54294134a632c478e978bcd7088e368828474a0d3716b884dd16c2a397edff357e76f90163061934f2c2acba1a77a5dcf6833beca000992e63e19dfaa5aee2a",
		"81c8b35d99bf89e561153a774bbbb57dde490be1cad98fe6e1e1cbb3b3e2520a00e854882fa5dfaf1118630ecca53171a74df24383fe44fa6a571fba9d235738"}

	builder := asset.NewQueryParamsBuilder().AssetIDs(existingAssetIDs)

	assets, err := asset.List(builder)
	q.NoError(err)
	q.True(len(assets) == len(existingAssetIDs))
}

func (q *QueryTestSuite) TestListNonExistingAssets() {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizP").Limit(10)
	assets, _ := asset.List(builder)
	q.True(len(assets) == 0)
}

func (q *QueryTestSuite) TestGetBitmark() {
	bitmarkID := "5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef"
	bitmark, err := bitmark.Get(bitmarkID)
	q.NoError(err)
	q.Equal(bitmark.ID, bitmarkID)
}

func (q *QueryTestSuite) TestGetBitmarkWithAsset() {
	bitmarkID := "5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef"
	assetID := "0e0b4e3bd771811d35a23707ba6197aa1dd5937439a221eaf8e7909309e7b31b6c0e06a1001c261a099abf04c560199db898bc154cf128aa9efa5efd36030c64"

	bitmark, asset, err := bitmark.GetWithAsset(bitmarkID)
	q.NoError(err)

	q.Equal(bitmark.ID, bitmarkID)
	q.Equal(asset.ID, assetID)
}

func (q *QueryTestSuite) TestGetNonExistingBitmark() {
	_, err := bitmark.Get("2bc5189e77b55f8f671c62cb46650c3b")
	q.EqualError(err, "[4000] message: bitmark not found reason: ")
}

func (q *QueryTestSuite) TestListBitmark() {
	limit := 1
	builder := bitmark.NewQueryParamsBuilder().
		IssuedBy("e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog").
		OwnedBy("eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9").
		ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e158386b0ad90515c69e7d1fd6df8f3d523e3550741e88d0d04798627a57b0006c9").
		LoadAsset(true).
		Limit(limit)

	bitmarks, _, err := bitmark.List(builder)
	q.NoError(err)
	q.True(len(bitmarks) <= limit)
}

func (q *QueryTestSuite) TestListBitmarkByBitmarkIDs() {
	existingBitmarkIDs := []string{
		"889f46d55ddbf6fae2da6fe14ca31b79ab84fe7cd104de735dc8cf9319eb68b5",
		"0d9a70dbad56820ac538417be3cacdcb643f295a1f2cf4812ad9fb4b56818221"}

	builder := bitmark.NewQueryParamsBuilder().
		BitmarkIDs(existingBitmarkIDs)

	bitmarks, _, err := bitmark.List(builder)
	q.NoError(err)

	q.Equal(len(bitmarks), len(existingBitmarkIDs), "The length should be the same.")
}

func (q *QueryTestSuite) TestListBitmarkByAtAndTo() {
	limit := 10
	at := 5
	builder := bitmark.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(limit)

	bitmarks, _, err := bitmark.List(builder)
	q.NoError(err)

	for _, b := range bitmarks {
		q.True(b.Offset <= at)
	}
}

func (q *QueryTestSuite) TestListNonExistingBitmarks() {
	builder := bitmark.NewQueryParamsBuilder().ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e15").Limit(10)

	bitmarks, _, err := bitmark.List(builder)
	q.NoError(err)
	q.True(len(bitmarks) == 0)
}

func (q *QueryTestSuite) TestGetTx() {
	txID := "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80"
	actual, err := tx.Get(txID)
	q.NoError(err)
	q.Equal(actual.ID, txID)
}

func (q *QueryTestSuite) TestGetTxWithAsset() {
	actualTx, actualAsset, err := tx.GetWithAsset("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80")
	q.NoError(err)

	expected := &tx.Tx{
		ID:      "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80",
		AssetID: "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
	}

	q.Equal(actualTx.ID, expected.ID)
	q.Equal(actualAsset.ID, expected.AssetID)
}

// FIXME
// func TestGetNonExistingTx(t *testing.T) {
// 	_, err := tx.Get("67ef8bfee0ef7b8c33eda34ba21c8b2b")
// 	if err.Error() != "[4000] message: tx not found reason: " {
// 		t.Fatalf("incorrect error message")
// 	}
// }

func (q *QueryTestSuite) TestListTxs() {
	builder := tx.NewQueryParamsBuilder().
		ReferencedBitmark("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80").
		LoadAsset(true).
		Limit(10)

	txs, _, err := tx.List(builder)
	q.NoError(err)
	q.True(len(txs) <= 10)
}

func (q *QueryTestSuite) TestListTxsByAtAndTo() {
	at := 5
	builder := tx.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(10).Pending(true)

	txs, _, err := tx.List(builder)
	q.NoError(err)

	for _, tx := range txs {
		q.True(tx.Offset <= at)
	}
}

func (q *QueryTestSuite) TestListNonExsitingTxs() {
	builder := tx.NewQueryParamsBuilder().ReferencedBitmark("2bc5189e77b55f8f671c62cb46650c3b").Limit(10)

	txs, _, err := tx.List(builder)
	q.NoError(err)
	q.Equal(len(txs), 0)
}

/*
func printBeautifulJSON(t *testing.T, v interface{}) {
	item, _ := json.MarshalIndent(v, "", "\t")
	t.Log("\n" + string(item))
}
*/
