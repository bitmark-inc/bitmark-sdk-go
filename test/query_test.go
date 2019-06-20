package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

func TestGetAsset(t *testing.T) {
	actual, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22")
	if err != nil {
		t.Error(err)
	}
	createdAt, _ := time.Parse("2006-01-02T15:04:05.000000Z", "2018-09-07T07:46:25.000000Z")
	expected := &asset.Asset{
		Id:   "2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22",
		Name: "HA25124377",
		Metadata: map[string]string{
			"Source":     "Bitmark Health",
			"Saved Time": "2018-09-07T07:45:41.948Z",
		},
		Fingerprint: "016ef802c0f912ed69a5afc0e6c08fbe96de3284e7cc6e685111d5f1705049f20b695443bc2d7bae7fe2091d9e7a880a50a51c2d0be1963a99b9914f60f2462040",
		Registrant:  "eTicVBQqmGzxNMGiZGtKzDdufXZsiFKH3SR8FcVYM7MQTZ47k3",
		Status:      "confirmed",
		BlockNumber: 8696,
		Sequence:    8581,
		CreatedAt:   createdAt,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("incorrect asset record:\nactual=%+v\nexpected=%+v", actual, expected)
	}
}

func TestGetNonExsitingAsset(t *testing.T) {
	_, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d55")
	if err.Error() != "[4000] message: asset not found reason: " {
		t.Fatalf("incorrect error message")
	}
}

func TestListAsset(t *testing.T) {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizPX6Mkyvod4vLQFdDemZvWsxiGr").Limit(10)

	assets, err := asset.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, a := range assets {
		printBeautifulJSON(t, a)
	}
}

func TestListAssetByAtAndTo(t *testing.T) {
	limit := 10
	at := 5

	builder := asset.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(limit).Pending(true)

	assets, err := asset.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, a := range assets {
		assert.True(t, a.Sequence <= at)
	}
}

func TestListAssetByAssetIds(t *testing.T) {
	existingAssetIds := []string{
		"c54294134a632c478e978bcd7088e368828474a0d3716b884dd16c2a397edff357e76f90163061934f2c2acba1a77a5dcf6833beca000992e63e19dfaa5aee2a",
		"81c8b35d99bf89e561153a774bbbb57dde490be1cad98fe6e1e1cbb3b3e2520a00e854882fa5dfaf1118630ecca53171a74df24383fe44fa6a571fba9d235738"}

	builder := asset.NewQueryParamsBuilder().AssetIds(existingAssetIds)

	assets, err := asset.List(builder)
	if err != nil {
		t.Error(err)
	}

	assert.True(t, len(assets) == len(existingAssetIds))
}

func TestListNonExsitingAssets(t *testing.T) {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizP").Limit(10)
	assets, _ := asset.List(builder)
	if len(assets) > 0 {
		t.Errorf("should return empty assets")
	}
}

func TestGetBitmark(t *testing.T) {
	bitmarkId := "5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef"
	bitmark, err := bitmark.Get(bitmarkId)
	if err != nil {
		t.Error(err)
	}
	printBeautifulJSON(t, bitmark)
	assert.Equal(t, bitmark.Id, bitmarkId)
}

func TestGetBitmarkWithAsset(t *testing.T) {
	bitmarkId := "5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef"
	assetId := "0e0b4e3bd771811d35a23707ba6197aa1dd5937439a221eaf8e7909309e7b31b6c0e06a1001c261a099abf04c560199db898bc154cf128aa9efa5efd36030c64"

	bitmark, asset, err := bitmark.GetWithAsset(bitmarkId)
	if err != nil {
		t.Error(err)
	}
	printBeautifulJSON(t, bitmark)
	printBeautifulJSON(t, asset)

	assert.Equal(t, bitmark.Id, bitmarkId)
	assert.Equal(t, asset.Id, assetId)
}

func TestGetNonExsitingBitmark(t *testing.T) {
	_, err := bitmark.Get("2bc5189e77b55f8f671c62cb46650c3b")
	if err.Error() != "[4000] message: bitmark not found reason: " {
		t.Fatalf("incorrect error message")
	}
}

func TestListBitmark(t *testing.T) {
	limit := 1
	builder := bitmark.NewQueryParamsBuilder().
		IssuedBy("e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog").
		OwnedBy("eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9").
		ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e158386b0ad90515c69e7d1fd6df8f3d523e3550741e88d0d04798627a57b0006c9").
		LoadAsset(true).
		Limit(limit)

	bitmarks, assets, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, b := range bitmarks {
		printBeautifulJSON(t, b)
	}

	for _, a := range assets {
		printBeautifulJSON(t, a)
	}

	assert.Equal(t, len(bitmarks), limit)
}

func TestListBitmarkByBitmarkIds(t *testing.T) {
	existingBitmarkIds := []string{
		"889f46d55ddbf6fae2da6fe14ca31b79ab84fe7cd104de735dc8cf9319eb68b5",
		"0d9a70dbad56820ac538417be3cacdcb643f295a1f2cf4812ad9fb4b56818221"}

	builder := bitmark.NewQueryParamsBuilder().
		BitmarkIds(existingBitmarkIds)

	bitmarks, _, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(bitmarks), len(existingBitmarkIds), "The length should be the same.")
}

func TestListBitmarkByAtAndTo(t *testing.T) {
	limit := 10
	at := 5
	builder := bitmark.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(limit)

	bitmarks, _, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, b := range bitmarks {
		assert.True(t, b.Commit <= at)
	}
}

func TestListNonExsitingBitmarks(t *testing.T) {
	builder := bitmark.NewQueryParamsBuilder().ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e15").Limit(10)

	bitmarks, _, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	if len(bitmarks) > 0 {
		t.Errorf("should return empty bitmarks")
	}
}

func TestGetTx(t *testing.T) {
	txId := "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80"
	actual, err := tx.Get(txId)
	if err != nil {
		t.Error(err)
	}

	printBeautifulJSON(t, actual)
	assert.Equal(t, txId, actual.Id)
}

func TestGetTxWithAsset(t *testing.T) {
	actualTx, actualAsset, err := tx.GetWithAsset("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80")
	if err != nil {
		t.Error(err)
	}

	expected := &tx.Tx{
		Id:      "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80",
		AssetId: "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
	}

	printBeautifulJSON(t, actualTx)
	printBeautifulJSON(t, actualAsset)

	assert.Equal(t, actualTx.Id, expected.Id)
	assert.Equal(t, actualAsset.Id, expected.AssetId)
}

func TestGetNonExsitingTx(t *testing.T) {
	_, err := tx.Get("67ef8bfee0ef7b8c33eda34ba21c8b2b")
	if err.Error() != "[4000] message: tx not found reason: " {
		t.Fatalf("incorrect error message")
	}
}

func TestListTxs(t *testing.T) {
	builder := tx.NewQueryParamsBuilder().
		ReferencedBitmark("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80").
		LoadAsset(true).
		Limit(10)

	txs, _, err := tx.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, tx := range txs {
		printBeautifulJSON(t, tx)
	}
}

func TestListTxsByAtAndTo(t *testing.T) {
	at := 5
	builder := tx.NewQueryParamsBuilder().At(at).To(utils.Earlier).Limit(10).Pending(true)

	txs, _, err := tx.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, tx := range txs {
		printBeautifulJSON(t, tx)
		assert.True(t, tx.Sequence <= at)
	}
}

func TestListNonExsitingTxs(t *testing.T) {
	builder := tx.NewQueryParamsBuilder().ReferencedBitmark("2bc5189e77b55f8f671c62cb46650c3b").Limit(10)

	txs, _, err := tx.List(builder)
	if err != nil {
		t.Error(err)
	}

	if len(txs) > 0 {
		t.Errorf("should return empty txs")
	}
}

func printBeautifulJSON(t *testing.T, v interface{}) {
	item, _ := json.MarshalIndent(v, "", "\t")
	t.Log("\n" + string(item))
}
