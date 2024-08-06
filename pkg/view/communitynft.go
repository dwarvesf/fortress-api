package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type GetNftMetadataResponse struct {
	Data *NftMetadata `json:"data"`
} // @name GetNftMetadataResponse

type NftMetadata struct {
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Image           string      `json:"image"`
	BackgroundColor string      `json:"background_color"`
	Attributes      []Attribute `json:"attributes"`
} // @name NftInfo

type Attribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
} // @name Attribute

func ToNftMetadata(nft *model.NftMetadata) *NftMetadata {
	attributes := make([]Attribute, 0)
	for _, attr := range nft.Attributes {
		attributes = append(attributes, Attribute{
			TraitType: attr.TraitType,
			Value:     attr.Value,
		})
	}
	return &NftMetadata{
		Name:            nft.Name,
		Description:     nft.Description,
		Image:           nft.Image,
		BackgroundColor: nft.BackgroundColor,
		Attributes:      attributes,
	}
}
