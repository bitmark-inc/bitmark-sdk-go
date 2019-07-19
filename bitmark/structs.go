package bitmark

import (
	"encoding/json"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/asset"
)

type Bitmark struct {
	ID          string         `json:"id"`
	AssetID     string         `json:"asset_id"`
	Asset       *asset.Asset   `json:"asset"`
	LatestTxID  string         `json:"head_id"` // TODO: rename api field
	Issuer      string         `json:"issuer"`
	Owner       string         `json:"owner"`
	Status      string         `json:"status"` // issuing, transferring, offering, settled
	Offer       *TransferOffer `json:"offer"`
	BlockNumber int            `json:"block_number"`
	Commit      int            `json:"offset"` // TODO: rename api field
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"confirmed_at"` // TODO: rename api field
}

type TransferOffer struct {
	ID        string                        `json:"id"`
	From      string                        `json:"from"`
	To        string                        `json:"to"`
	Record    *CountersignedTransferRequest `json:"record"`
	ExtraInfo map[string]string             `json:"extra_info"`
	CreatedAt time.Time                     `json:"created_at"`
	Open      bool                          `json:"open"`
}

type Share struct {
	ID        string `json:"share_id"`
	Owner     string `json:"owner"`
	Balance   uint64 `json:"balance"`
	Available uint64 `json:"available"`
}

type ShareOffer struct {
	ID        string          `json:"id"`
	ShareID   string          `json:"share_id"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Record    GrantRequest    `json:"record"`
	ExtraInfo json.RawMessage `json:"extra_info"`
	CreatedAt time.Time       `json:"created_at"`
}
