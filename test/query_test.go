package test

import (
	"fmt"
	"testing"

	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
)

func TestGetAsset(t *testing.T) {
	asset, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", asset)
}

func TestListAsset(t *testing.T) {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizPX6Mkyvod4vLQFdDemZvWsxiGr").Limit(10)

	it := asset.NewIterator(builder)
	for it.Before() {
		for _, b := range it.Values() {
			t.Logf("%+v\n", b)
		}
	}
	if it.Err() != nil {
		t.Error(it.Err())
	}
}

func TestGetBitmark(t *testing.T) {
	bitmark, err := bitmark.Get("5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef", true)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", bitmark)
}

func TestListBitmark(t *testing.T) {
	builder := bitmark.NewQueryParamsBuilder().
		IssuedBy("e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog").
		OwnedBy("eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9", true).
		ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e158386b0ad90515c69e7d1fd6df8f3d523e3550741e88d0d04798627a57b0006c9").
		LoadAsset(true).
		Limit(10)

	it := bitmark.NewIterator(builder)
	for it.Before() {
		for _, b := range it.Values() {
			t.Logf("%+v\n", b)
		}
	}
	if it.Err() != nil {
		t.Error(it.Err())
	}
}

func TestGetTx(t *testing.T) {
	tx, err := tx.Get("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80", true)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v", tx)
}

func TestListProvenance(t *testing.T) {
	builder := tx.NewQueryParamsBuilder().
		ReferencedBitmark("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80").
		LoadAsset(true).
		Limit(10)

	it := tx.NewIterator(builder)
	for it.Before() {
		for _, tx := range it.Values() {
			t.Logf("%+v\n", tx)
		}
	}
	if it.Err() != nil {
		t.Error(it.Err())
	}
}
