package model

import "time"

type TransferRequestResponse struct {
	ProfileID   string `json:"profile_id"`
	RequestCode string `json:"request_code"`
	Status      string `json:"status"`
	TxID        int    `json:"tx_id"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
	Amount      string `json:"amount"`
	TokenName   string `json:"token_name"`
	TokenID     string `json:"token_id"`
}

type IcyInfo struct {
	CirculatedIcyBalance string  `json:"circulated_icy_balance"`
	IcySatoshiRate       string  `json:"icy_satoshi_rate"`
	IcyUsdRate           string  `json:"icy_usd_rate"`
	MinIcyToSwap         string  `json:"min_icy_to_swap"`
	MinSatoshiFee        string  `json:"min_satoshi_fee"`
	SatoshiBalance       string  `json:"satoshi_balance"`
	SatoshiPerUsd        float64 `json:"satoshi_per_usd"`
	SatoshiUsdRate       string  `json:"satoshi_usd_rate"`
	ServiceFeeRate       float64 `json:"service_fee_rate"`
}

type IcyInfoResponse struct {
	Data    IcyInfo `json:"data"`
	Message string  `json:"message"`
}

type GenerateSignature struct {
	BtcAmount string `json:"btc_amount"`
	Deadline  string `json:"deadline"`
	IcyAmount string `json:"icy_amount"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

type GenerateSignatureResponse struct {
	Data    GenerateSignature `json:"data"`
	Message string            `json:"message"`
	Error   *string           `json:"error"`
}

type GenerateSignatureRequest struct {
	BtcAmount  string `json:"btc_amount"`
	IcyAmount  string `json:"icy_amount"`
	BtcAddress string `json:"btc_address"`
}

type SwapResponse struct {
	TxHash string `json:"tx_hash"`
}

type TransferResponse struct {
	TxHash string `json:"tx_hash"`
}

const (
	IcySwapBtcStatusSuccess = "success"
	IcySwapBtcStatusFailed  = "failed"
	IcySwapBtcStatusPending = "pending"
)

type IcySwapBtcRequest struct {
	ID                UUID       `sql:",type:uuid" json:"id" gorm:"default:uuid()"`
	ProfileID         string     `json:"profile_id" gorm:"profile_id"`
	RequestCode       string     `json:"request_code" gorm:"request_code"`
	TxStatus          string     `json:"tx_status" gorm:"tx_status"`
	TxID              int        `json:"tx_id" gorm:"tx_id"`
	BtcAddress        string     `json:"btc_address" gorm:"btc_address"`
	Timestamp         int64      `json:"timestamp" gorm:"timestamp"`
	Amount            string     `json:"amount" gorm:"amount"`
	TokenName         string     `json:"token_name" gorm:"token_name"`
	TokenID           string     `json:"token_id" gorm:"token_id"`
	SwapRequestStatus string     `json:"swap_request_status" gorm:"swap_request_status"`
	SwapRequestError  string     `json:"swap_request_error" gorm:"swap_request_error"`
	RevertStatus      string     `json:"revert_status" gorm:"revert_status"`
	RevertError       string     `json:"revert_error" gorm:"revert_error"`
	WithdrawStatus    string     `json:"withdraw_status" gorm:"withdraw_status"`
	WithdrawError     string     `json:"withdraw_error" gorm:"withdraw_error"`
	TxSwap            string     `json:"tx_swap" gorm:"tx_swap"`
	TxDeposit         string     `json:"tx_deposit" gorm:"tx_deposit"`
	CreatedAt         time.Time  `sql:"default:now()" json:"created_at"`
	UpdatedAt         *time.Time `sql:"default:now()" json:"updated_at"`
}
