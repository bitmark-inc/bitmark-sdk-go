package main

import (
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
)

func queryBitmarks(queryParamsBuilder *bitmark.QueryParamsBuilder) ([]*bitmark.Bitmark, error) {
	bitmarks, _, err := bitmark.List(queryParamsBuilder)
	return bitmarks, err
}

func queryBitmarkById(bitmarkId string) (*bitmark.Bitmark, error) {
	bitmark, err := bitmark.Get(bitmarkId)
	return bitmark, err
}

func queryAssets(queryParamsBuilder *asset.QueryParamsBuilder) ([]*asset.Asset, error) {
	assets, err := asset.List(queryParamsBuilder)
	return assets, err
}

func queryAssetById(assetId string) (*asset.Asset, error) {
	asset, err := asset.Get(assetId)
	return asset, err
}

func queryTransactions(queryParamsBuilder *tx.QueryParamsBuilder) ([]*tx.Tx, error) {
	txs, _, err := tx.List(queryParamsBuilder)
	return txs, err
}

func queryTransactionById(txId string) (*tx.Tx, error) {
	tx, err := tx.Get(txId)
	return tx, err
}
