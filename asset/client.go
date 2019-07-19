package asset

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

type registrationRequest struct {
	Assets []*RegistrationParams `json:"assets"`
}

type registeredItem struct {
	ID        string `json:"id"`
	Duplicate bool   `json:"duplicate"`
}

func Register(params *RegistrationParams) (string, error) {
	r := registrationRequest{
		Assets: []*RegistrationParams{params},
	}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(r)

	client := sdk.GetAPIClient()
	req, _ := client.NewRequest("POST", "/v3/register-asset", body)

	var result struct {
		Assets []registeredItem `json:"assets"`
	}
	if err := client.Do(req, &result); err != nil {
		return "", err
	}
	return result.Assets[0].ID, nil
}

func Get(assetID string) (*Asset, error) {
	client := sdk.GetAPIClient()

	req, err := client.NewRequest("GET", "/v3/assets/"+assetID+"?pending=true", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Asset *Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Asset, nil
}

func List(builder *QueryParamsBuilder) ([]*Asset, error) {
	params, err := builder.Build()

	if err != nil {
		return nil, err
	}

	client := sdk.GetAPIClient()
	req, err := client.NewRequest("GET", "/v3/assets?"+params, nil)

	if err != nil {
		return nil, err
	}

	var result struct {
		Assets []*Asset `json:"assets"`
	}

	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Assets, nil
}

type QueryParamsBuilder struct {
	params url.Values
	err    error
}

func NewQueryParamsBuilder() *QueryParamsBuilder {
	return &QueryParamsBuilder{params: url.Values{}}
}

func (qb *QueryParamsBuilder) RegisteredBy(registrant string) *QueryParamsBuilder {
	qb.params.Set("registrant", registrant)
	return qb
}

func (qb *QueryParamsBuilder) AssetIDs(assetIDs []string) *QueryParamsBuilder {
	for _, assetID := range assetIDs {
		qb.params.Add("asset_ids", assetID)
	}
	return qb
}

func (qb *QueryParamsBuilder) Pending(pending bool) *QueryParamsBuilder {
	qb.params.Set("pending", strconv.FormatBool(pending))
	return qb
}

func (qb *QueryParamsBuilder) Limit(size int) *QueryParamsBuilder {
	if size > 100 {
		qb.err = errors.New("invalid size: max = 100")
	}
	qb.params.Set("limit", strconv.Itoa(size))
	return qb
}

func (qb *QueryParamsBuilder) At(at int) *QueryParamsBuilder {
	qb.params.Set("at", strconv.Itoa(at))
	return qb
}

func (qb *QueryParamsBuilder) To(direction utils.Direction) *QueryParamsBuilder {
	qb.params.Set("to", string(direction))
	return qb
}

func (qb *QueryParamsBuilder) Build() (string, error) {
	if qb.err != nil {
		return "", qb.err
	}

	if qb.params.Get("pending") == "" {
		qb.params.Set("pending", "true")
	}

	return qb.params.Encode(), nil
}
