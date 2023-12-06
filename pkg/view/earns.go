package view

import (
	"time"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service/mochipay"
)

type MochiTransaction struct {
	ID                 string                 `json:"id"`
	FromProfileID      string                 `json:"fromProfileID"`
	OtherProfileID     string                 `json:"otherProfileID"`
	FromProfileSource  string                 `json:"fromProfileSource"`
	OtherProfileSource string                 `json:"otherProfileSource"`
	SourcePlatform     string                 `json:"sourcePlatform"`
	Amount             string                 `json:"amount"`
	TokenID            string                 `json:"tokenID"`
	ChainID            string                 `json:"chainID"`
	InternalID         int64                  `json:"internalID"`
	ExternalID         string                 `json:"externalID"`
	OnchainTxHash      string                 `json:"onchainTxHash"`
	Type               string                 `json:"type"`
	Action             string                 `json:"action"`
	Status             string                 `json:"status"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	ExpiredAt          *time.Time             `json:"expiredAt"`
	SettledAt          *time.Time             `json:"settledAt"`
	Token              *Token                 `json:"token"`
	OriginalTxID       string                 `json:"originalTxID"`
	OtherProfile       *MochiProfile          `json:"otherProfile"`
	FromProfile        *MochiProfile          `json:"fromProfile"`
	OtherProfiles      []MochiProfile         `json:"otherProfiles"`
	AmountEachProfiles []AmountEachProfiles   `json:"amountEachProfiles"`
	USDAmount          float64                `json:"usdAmount"`
	Metadata           map[string]interface{} `json:"metadata"`
	OtherProfileIds    []string               `json:"otherProfileIds"`
	TotalAmount        string                 `json:"totalAmount"`
	FromTokenId        string                 `json:"fromTokenId"`
	ToTokenId          string                 `json:"toTokenId"`
	FromToken          *Token                 `json:"fromToken,omitempty"`
	ToToken            *Token                 `json:"toToken,omitempty"`
	FromAmount         string                 `json:"fromAmount"`
	ToAmount           string                 `json:"toAmount"`
} // @name MochiTransaction

type Token struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Decimal     int64   `json:"decimal"`
	ChainID     string  `json:"chainID"`
	Native      bool    `json:"native"`
	Address     string  `json:"address"`
	Icon        string  `json:"icon"`
	CoinGeckoID string  `json:"coinGeckoID"`
	Price       float64 `json:"price"`
	Chain       *Chain  `json:"chain"`
} // @name Token

type Chain struct {
	ID       string `json:"id"`
	ChainID  string `json:"chainID"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	RPC      string `json:"rpc"`
	Explorer string `json:"explorer"`
	Icon     string `json:"icon"`
	Type     string `json:"type"`
} // @name Chain

type MochiProfile struct {
	ID                 string               `json:"id"`
	CreatedAt          string               `json:"createdAt"`
	UpdatedAt          string               `json:"updatedAt"`
	ProfileName        string               `json:"profileName"`
	Avatar             string               `json:"avatar"`
	AssociatedAccounts []AssociatedAccounts `json:"associatedAccounts"`
	Type               string               `json:"type"`
	Application        *Application         `json:"application"`
} // @name MochiProfile

type Application struct {
	ID                   int     `json:"id"`
	Name                 string  `json:"name"`
	OwnerProfileID       string  `json:"ownerProfileID"`
	ServiceFee           float64 `json:"serviceFee"`
	ApplicationProfileID string  `json:"applicationProfileID"`
	Active               bool    `json:"active"`
} // @name Application

type AssociatedAccounts struct {
	ID                 string      `json:"id"`
	ProfileID          string      `json:"profileID"`
	Platform           string      `json:"platform"`
	PlatformIdentifier string      `json:"platformIdentifier"`
	PlatformMetadata   interface{} `json:"platformMetadata"`
	IsGuildMember      bool        `json:"isGuildMember"`
	CreatedAt          string      `json:"createdAt"`
	UpdatedAt          string      `json:"updatedAt"`
} // @name AssociatedAccounts

type AmountEachProfiles struct {
	ProfileID string  `json:"profileID"`
	Amount    string  `json:"amount"`
	UsdAmount float64 `json:"usdAmount"`
} // @name AmountEachProfiles

func toMochiTransaction(data mochipay.TransactionData) MochiTransaction {
	return MochiTransaction{
		ID:                 data.Id,
		FromProfileID:      data.FromProfileId,
		OtherProfileID:     data.OtherProfileId,
		FromProfileSource:  data.FromProfileSource,
		OtherProfileSource: data.OtherProfileSource,
		SourcePlatform:     data.SourcePlatform,
		Amount:             data.Amount,
		TokenID:            data.TokenId,
		ChainID:            data.ChainId,
		InternalID:         data.InternalId,
		ExternalID:         data.ExternalId,
		OnchainTxHash:      data.OnchainTxHash,
		Type:               string(data.Type),
		Action:             string(data.Action),
		Status:             string(data.Status),
		CreatedAt:          data.CreatedAt,
		UpdatedAt:          data.UpdatedAt,
		ExpiredAt:          data.ExpiredAt,
		SettledAt:          data.SettledAt,
		Token:              toToken(data.Token),
		OriginalTxID:       data.OriginalTxId,
		OtherProfile:       toMochiProfile(data.OtherProfile),
		FromProfile:        toMochiProfile(data.FromProfile),
		OtherProfiles:      toMochiProfiles(data.OtherProfiles),
		AmountEachProfiles: toAmountEachProfiles(data.AmountEachProfiles),
		USDAmount:          data.UsdAmount,
		Metadata:           data.Metadata,
		OtherProfileIds:    data.OtherProfileIds,
		TotalAmount:        data.TotalAmount,
		FromTokenId:        data.FromTokenId,
		ToTokenId:          data.ToTokenId,
		FromToken:          toToken(data.FromToken),
		ToToken:            toToken(data.ToToken),
		FromAmount:         data.FromAmount,
		ToAmount:           data.ToAmount,
	}
}

func toToken(data *mochipay.Token) *Token {
	if data == nil {
		return nil
	}
	return &Token{
		ID:          data.Id,
		Name:        data.Name,
		Symbol:      data.Symbol,
		Decimal:     data.Decimal,
		ChainID:     data.ChainId,
		Native:      data.Native,
		Address:     data.Address,
		Icon:        data.Icon,
		CoinGeckoID: data.CoinGeckoId,
		Price:       data.Price,
		Chain:       toChain(data.Chain),
	}
}

func toChain(data *mochipay.Chain) *Chain {
	if data == nil {
		return nil
	}

	return &Chain{
		ID:       data.Id,
		ChainID:  data.ChainId,
		Name:     data.Name,
		Symbol:   data.Symbol,
		RPC:      data.Rpc,
		Explorer: data.Explorer,
		Icon:     data.Icon,
		Type:     data.Type,
	}
}

func toMochiProfile(data *mochipay.MochiProfile) *MochiProfile {
	if data == nil {
		return nil
	}

	return &MochiProfile{
		ID:                 data.Id,
		CreatedAt:          data.CreatedAt,
		UpdatedAt:          data.UpdatedAt,
		ProfileName:        data.ProfileName,
		Avatar:             data.Avatar,
		AssociatedAccounts: toAssociatedAccounts(data.AssociatedAccounts),
		Type:               data.Type,
		Application:        toApplication(data.Application),
	}
}

func toApplication(data *mochipay.Application) *Application {
	if data == nil {
		return nil
	}

	return &Application{
		ID:                   data.Id,
		Name:                 data.Name,
		OwnerProfileID:       data.OwnerProfileId,
		ServiceFee:           data.ServiceFee,
		ApplicationProfileID: data.ApplicationProfileId,
		Active:               data.Active,
	}
}

func toMochiProfiles(data []mochipay.MochiProfile) []MochiProfile {
	result := make([]MochiProfile, len(data))
	for _, item := range data {
		result = append(result, *toMochiProfile(&item))
	}
	return result
}

func toAssociatedAccounts(data []mochipay.AssociatedAccounts) []AssociatedAccounts {
	result := make([]AssociatedAccounts, len(data))
	for i, item := range data {
		result[i] = AssociatedAccounts{
			ID:                 item.Id,
			ProfileID:          item.ProfileId,
			Platform:           item.Platform,
			PlatformIdentifier: item.PlatformIdentifier,
			PlatformMetadata:   item.PlatformMetadata,
			IsGuildMember:      item.IsGuildMember,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		}
	}
	return result
}

func toAmountEachProfiles(data []mochipay.AmountEachProfiles) []AmountEachProfiles {
	result := make([]AmountEachProfiles, len(data))
	for i, item := range data {
		result[i] = AmountEachProfiles{
			ProfileID: item.ProfileId,
			Amount:    item.Amount,
			UsdAmount: item.UsdAmount,
		}
	}
	return result
}

func ToEmployeeEarnsTransactions(report model.EmployeeEarnTransactions) []MochiTransaction {
	var result []MochiTransaction
	for _, r := range report {
		result = append(result, toMochiTransaction(r))
	}
	return result
}

type GetEmployeeEarnTransactionsResponse struct {
	Pagination
	Total int64 `json:"total"`

	Data []MochiTransaction `json:"data"`
} // @name GetEmployeeEarnTransactionsResponse

type EmployeeTotalEarn struct {
	TotalEarnsICY string  `json:"totalEarnsICY"`
	TotalEarnsUSD float64 `json:"totalEarnsUSD"`
} // @name EmployeeTotalEarn

type GetEmployeeTotalEarnResponse struct {
	Data EmployeeTotalEarn `json:"data"`
} // @name GetEmployeeTotalEarnResponse
