package mochipay

import (
	"time"
)

type TransactionType string

const (
	TransactionTypeSend    TransactionType = "out"
	TransactionTypeReceive TransactionType = "in"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusSubmitted TransactionStatus = "submitted"
	TransactionStatusSuccess   TransactionStatus = "success"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusExpired   TransactionStatus = "expired"
	TransactionStatusCancelled TransactionStatus = "cancelled" //nolint:all
)

type TransactionAction string

const (
	TransactionActionTransfer      TransactionAction = "transfer"
	TransactionActionAirdrop       TransactionAction = "airdrop"
	TransactionActionDeposit       TransactionAction = "deposit"
	TransactionActionWithdraw      TransactionAction = "withdraw"
	TransactionActionSwap          TransactionAction = "swap"
	TransactionActionVaultTransfer TransactionAction = "vault_transfer"
)

type TransactionPlatform string

const (
	TransactionPlatformDiscord TransactionPlatform = "discord"
)

type ListTransactionsRequest struct {
	Type         TransactionType       `json:"type"`
	Status       TransactionStatus     `json:"status"`
	ActionList   []TransactionAction   `json:"action_list"`
	TokenAddress string                `json:"token_address"`
	ProfileID    string                `json:"profile_id"`
	Platforms    []TransactionPlatform `json:"platforms"`
	ChainIDs     []string              `json:"chain_ids"`
	Page         int64                 `json:"page"`
	Size         int64                 `json:"size"`
	IsSender     *bool                 `json:"is_sender"`
	SortBy       string                `json:"sort_by"`
}

type ListTransactionsResponse struct {
	Data       []TransactionData `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

type Pagination struct {
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Size  int64 `json:"size"`
}

type TransactionData struct {
	Id                 string                 `json:"id"`
	FromProfileId      string                 `json:"from_profile_id"`
	OtherProfileId     string                 `json:"other_profile_id"`
	FromProfileSource  string                 `json:"from_profile_source"`
	OtherProfileSource string                 `json:"other_profile_source"`
	SourcePlatform     string                 `json:"source_platform"`
	Amount             string                 `json:"amount"`
	TokenId            string                 `json:"token_id"`
	ChainId            string                 `json:"chain_id"`
	InternalId         int64                  `json:"internal_id"`
	ExternalId         string                 `json:"external_id"`
	OnchainTxHash      string                 `json:"onchain_tx_hash"`
	Type               TransactionType        `json:"type"`
	Action             TransactionAction      `json:"action"`
	Status             TransactionStatus      `json:"status"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	ExpiredAt          *time.Time             `json:"expired_at"`
	SettledAt          *time.Time             `json:"settled_at"`
	Token              *Token                 `json:"token"`
	OriginalTxId       string                 `json:"original_tx_id"`
	OtherProfile       *MochiProfile          `json:"other_profile"`
	FromProfile        *MochiProfile          `json:"from_profile"`
	OtherProfiles      []MochiProfile         `json:"other_profiles"`
	AmountEachProfiles []AmountEachProfiles   `json:"amount_each_profiles"`
	UsdAmount          float64                `json:"usd_amount"`
	Metadata           map[string]interface{} `json:"metadata"`

	// used in airdrop response
	OtherProfileIds []string `json:"other_profile_ids"`
	TotalAmount     string   `json:"total_amount"`

	// used in swap response
	FromTokenId string `json:"from_token_id"`
	ToTokenId   string `json:"to_token_id"`
	FromToken   *Token `json:"from_token,omitempty"`
	ToToken     *Token `json:"to_token,omitempty"`
	FromAmount  string `json:"from_amount"`
	ToAmount    string `json:"to_amount"`
}

type Token struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Decimal     int64   `json:"decimal"`
	ChainId     string  `json:"chain_id"`
	Native      bool    `json:"native"`
	Address     string  `json:"address"`
	Icon        string  `json:"icon"`
	CoinGeckoId string  `json:"coin_gecko_id"`
	Price       float64 `json:"price"`
	Chain       *Chain  `json:"chain"`
}

type Chain struct {
	Id       string `json:"id"`
	ChainId  string `json:"chain_id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Rpc      string `json:"rpc"`
	Explorer string `json:"explorer"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
}

type MochiProfile struct {
	Id                 string               `json:"id"`
	CreatedAt          string               `json:"created_at"`
	UpdatedAt          string               `json:"updated_at"`
	ProfileName        string               `json:"profile_name"`
	Avatar             string               `json:"avatar"`
	AssociatedAccounts []AssociatedAccounts `json:"associated_accounts"`
	Type               string               `json:"type"`
	Application        *Application         `json:"application"`
}

type Application struct {
	Id                   int     `json:"id"`
	Name                 string  `json:"name"`
	OwnerProfileId       string  `json:"owner_profile_id"`
	ServiceFee           float64 `json:"service_fee"`
	ApplicationProfileId string  `json:"application_profile_id"`
	Active               bool    `json:"active"`
}

type AssociatedAccounts struct {
	Id                 string      `json:"id"`
	ProfileId          string      `json:"profile_id"`
	Platform           string      `json:"platform"`
	PlatformIdentifier string      `json:"platform_identifier"`
	PlatformMetadata   interface{} `json:"platform_metadata"`
	IsGuildMember      bool        `json:"is_guild_member"`
	CreatedAt          string      `json:"created_at"`
	UpdatedAt          string      `json:"updated_at"`
}

type AmountEachProfiles struct {
	ProfileId string  `json:"profile_id"`
	Amount    string  `json:"amount"`
	UsdAmount float64 `json:"usd_amount"`
}

type TokenInfo struct {
	Address     string `json:"address"`
	Chain       *Chain `json:"chain"`
	ChainID     string `json:"chain_id"`
	CoinGeckoID string `json:"coin_gecko_id"`
	Decimal     int    `json:"decimal"`
	Icon        string `json:"icon"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Native      bool   `json:"native"`
	Price       int    `json:"price"`
	Symbol      string `json:"symbol"`
}

type VaultRequest struct {
	Amount     string     `json:"amount"`
	Chain      string     `json:"chain"`
	ListNotify []string   `json:"list_notify"`
	Message    string     `json:"message"`
	Name       string     `json:"name"`
	Platform   string     `json:"platform"`
	PrivateKey string     `json:"private_key"`
	ProfileID  string     `json:"profile_id"`
	Receiver   string     `json:"receiver"`
	RequestID  int        `json:"request_id"`
	To         string     `json:"to"`
	Token      string     `json:"token"`
	TokenID    string     `json:"token_id"`
	TokenInfo  *TokenInfo `json:"token_info"`
	VaultID    int        `json:"vault_id"`
}

type TransactionMetadata struct {
	Message              string        `json:"message"`
	RecipientProfileType string        `json:"recipient_profile_type"`
	RequestID            int           `json:"request_id"`
	SenderProfileType    string        `json:"sender_profile_type"`
	TransferType         string        `json:"transfer_type"`
	TxHash               string        `json:"tx_hash"`
	VaultRequest         *VaultRequest `json:"vault_request"`
}

type BatchBalancesResponse struct {
	Data []BatchBalancesData `json:"data"`
}

type BatchBalancesData struct {
	Id        string `json:"id"`
	ProfileID string `json:"profile_id"`
	TokenID   string `json:"token_id"`
	Amount    string `json:"amount"`
	Token     Token  `json:"token"`
}

type TransferFromVaultRequest struct {
	RecipientIDs []string `json:"recipient_ids"`
	Amounts      []string `json:"amounts"`
	TokenID      string   `json:"token_id"`
	VaultID      string   `json:"vault_id"`
	References   string   `json:"references"`
	Description  string   `json:"description"`
}

type TransactionFromVaultResponse struct {
	Data []VaultTransaction `json:"data"`
}

type VaultTransaction struct {
	Timestamp    int64  `json:"timestamp"`
	TxId         int64  `json:"tx_id"`
	RecipientId  string `json:"recipient_id"`
	Amount       string `json:"amount"`
	Status       string `json:"status"`
	References   string `json:"references"`
	TokenId      string `json:"token_id"`
	TxFee        string `json:"tx_fee"`
	TxFeePercent string `json:"tx_fee_percent"`
}

type Metadata map[string]interface{}

type WithdrawFromVaultRequest struct {
	Address  string   `json:"address"`
	Amount   string   `json:"amount"`
	Metadata Metadata `json:"metadata,omitempty"`
	Platform string   `json:"platform"`
	TokenID  string   `json:"token_id"`
	VaultID  string   `json:"vault_id"`
}

type WithdrawFromVaultResponse struct {
	Message string `json:"message"`
}

type DepositToVaultRequest struct {
	Platform string `json:"platform"`
	Token    string `json:"token"`
	VaultID  string `json:"vault_id"`
}

type Contract struct {
	Address     string    `json:"address"`
	Chain       *Chain    `json:"chain"`
	ChainID     string    `json:"chain_id"`
	CreatedAt   time.Time `json:"created_at"`
	ID          string    `json:"id"`
	LastSweepAt time.Time `json:"last_sweep_at"`
}

type DepositToVault struct {
	ChainID        string     `json:"chain_id"`
	Contract       Contract   `json:"contract"`
	ContractID     string     `json:"contract_id"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiredAt      time.Time  `json:"expired_at"`
	Platform       string     `json:"platform"`
	PlatformUserID string     `json:"platform_user_id"`
	ProfileID      string     `json:"profile_id"`
	Token          *TokenInfo `json:"token"`
	TokenID        string     `json:"token_id"`
}

type DepositToVaultResponse struct {
	Data []DepositToVault `json:"data"`
}
