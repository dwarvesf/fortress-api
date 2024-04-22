package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetNftMetadataResponse struct {
	Data *NftMetadata `json:"data"`
} // @name GetNftMetadataResponse

type NftMetadata struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       string      `json:"image"`
	Attributes  []attribute `json:"attributes"`
} // @name NftInfo

type attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

func ToNftMetadata(nft *model.NftMetadata) *NftMetadata {
	attributes := make([]attribute, 0)
	for _, attr := range nft.Attributes {
		attributes = append(attributes, attribute{
			TraitType: attr.TraitType,
			Value:     attr.Value,
		})
	}
	return &NftMetadata{
		Name:        nft.Name,
		Description: nft.Description,
		Image:       nft.Image,
		Attributes:  attributes,
	}
}
