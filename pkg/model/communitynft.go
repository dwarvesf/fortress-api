package model

// Client store information of a Client
type NftMetadata struct {
	Name            string
	Description     string
	Image           string
	BackgroundColor string
	Attributes      []NftAttribute
}

type NftAttribute struct {
	TraitType string
	Value     string
}
