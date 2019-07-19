package asset

import (
	"fmt"
	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, string(b), `{
			"assets":[
				{
					"name": "",
					"fingerprint": "fingerprint",
					"metadata": "",
					"registrant": "registrant",
					"signature": "signature"
				}
			]}`)
		fmt.Fprintln(w, `{"assets":[{"id":"asset_id"}]}`)
	}))
	defer ts.Close()

	sdk.Init(&sdk.Config{
		HTTPClient: ts.Client(),
		Network:    sdk.Testnet,
	})
	sdk.GetAPIClient().URLAuthority = ts.URL

	assetID, err := Register(&RegistrationParams{
		Name:        "",
		Metadata:    "",
		Fingerprint: "fingerprint",
		Registrant:  "registrant",
		Signature:   "signature",
	})
	assert.Equal(t, assetID, "asset_id")
	assert.NoError(t, err)
}

func TestGetAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{
			"asset": {
				"id": "be8af6c36ad4e4e15129fbbf6c2a6e75ac984a868cd9e76e75e8f60c5973b7722073dad1dbe4c70f6e65f4fbe42b100b1239ae48449cb75ec9367da47ce8d4a7",
				"name": "Logo 1",
				"fingerprint": "01eb9283dcbf7361b3534afdd6b7f24a784f06aae66498fd660a8f8dd8725b3759761ebc7124c7753095adb330c1e30223b407c2d049b8990c7edd173d8573eb18",
				"metadata": {},
				"registrant": "eabJKs8sDYT6FJGGUoA882T2Es1k8rho4cRiMPcjTuEURsC1Su",
				"status": "confirmed",
				"block_number": 26830,
				"block_offset": 1,
				"expires_at": null,
				"offset": 450434,
				"created_at": "2019-07-15T08:44:45.000000Z"
			}
		}`)
	}))
	defer ts.Close()

	sdk.Init(&sdk.Config{
		HTTPClient: ts.Client(),
		Network:    sdk.Testnet,
	})
	sdk.GetAPIClient().URLAuthority = ts.URL

	asset, err := Get("asset_id")
	assert.Equal(t, asset, &Asset{
		ID:          "be8af6c36ad4e4e15129fbbf6c2a6e75ac984a868cd9e76e75e8f60c5973b7722073dad1dbe4c70f6e65f4fbe42b100b1239ae48449cb75ec9367da47ce8d4a7",
		Name:        "Logo 1",
		Metadata:    map[string]string{},
		Fingerprint: "01eb9283dcbf7361b3534afdd6b7f24a784f06aae66498fd660a8f8dd8725b3759761ebc7124c7753095adb330c1e30223b407c2d049b8990c7edd173d8573eb18",
		Registrant:  "eabJKs8sDYT6FJGGUoA882T2Es1k8rho4cRiMPcjTuEURsC1Su",
		Status:      "confirmed",
		BlockNumber: 26830,
		Sequence:    450434,
		CreatedAt:   time.Date(2019, 07, 15, 8, 44, 45, 0, time.UTC),
	})
	assert.NoError(t, err)
}

func TestListAsset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Query()["asset_ids"], []string{"a", "b"})
		assert.Equal(t, r.URL.Query().Get("registrant"), "r")
		assert.Equal(t, r.URL.Query().Get("at"), "3")
		assert.Equal(t, r.URL.Query().Get("to"), "earlier")
		assert.Equal(t, r.URL.Query().Get("limit"), "10")
		assert.Equal(t, r.URL.Query().Get("pending"), "false")

		fmt.Fprintln(w, `{
			"assets": [
				{
					"id": "f779ce0be5802eb907353e19bf10a910bc9bb7caa82759a56686a508af9aecb07b895e0838f90d6c677249298704af7f52b6757b7d79e72139dc18cae5eb678b",
					"name": "SwiftSDK__rekey_test_7nLLBlUj",
					"fingerprint": "01cfba8139e3092eb62dddf6153a5cff59aa9bfe5c83702b12c276da93b4a6a577f46de2be95990894488c916c4fbeaa60d78cbfef03072906b0287878240a6cbc",
					"metadata": {
						"Random string": "AJC7wvqOhWi9Kg17nY6n"
					},
					"registrant": "fEWr2PV8GdSFCG2WMxuBQsPaQG6wPSXYaJYAKNBLbzjC7Ys8oR",
					"status": "confirmed",
					"block_number": 26836,
					"block_offset": 1,
					"expires_at": null,
					"offset": 450442,
					"created_at": "2019-07-15T09:29:39.000000Z"
				},
				{
					"id": "4c63b0089881457c601d2ac3fd4b6ac18b5b733a2999bbb038ab629a512da69d7ad5e60262604621c70e15bee12c97cb780824b3cea37f56e2b4407ef694f1c0",
					"name": "SwiftSDK_test_qi2HqL1j",
					"fingerprint": "01f968c7ebc555a5acb165539b0cb899c3a8310501f913f07a5c98d301524c3f840117ecd6a6c71c21a51a081d0b1b833e91bfb91ea280e9a6b28b98aa58cc7f33",
					"metadata": {
						"Random string": "jjdCaSHcNx41mAxnFk6F"
					},
					"registrant": "fEX1UjHkgN1GqF1YF9UxGFgq2RUatB4N92rd6f3ZWbwKTQuVDM",
					"status": "confirmed",
					"block_number": 26834,
					"block_offset": 1,
					"expires_at": null,
					"offset": 450439,
					"created_at": "2019-07-15T09:27:34.000000Z"
				}
			]
		}`)
	}))
	defer ts.Close()

	sdk.Init(&sdk.Config{
		HTTPClient: ts.Client(),
		Network:    sdk.Testnet,
	})
	sdk.GetAPIClient().URLAuthority = ts.URL

	builder := NewQueryParamsBuilder().Limit(101)
	_, err := builder.Build()
	assert.Error(t, err)

	builder = NewQueryParamsBuilder()
	params, err := builder.Build()
	assert.Equal(t, params, "pending=true")
	assert.NoError(t, err)

	builder = NewQueryParamsBuilder().
		AssetIDs([]string{"a", "b"}).
		RegisteredBy("r").
		To(utils.Earlier).
		At(3).
		Limit(10).Pending(false)
	assets, err := List(builder)
	assert.Equal(t, len(assets), 2)
	assert.NoError(t, err)
}
