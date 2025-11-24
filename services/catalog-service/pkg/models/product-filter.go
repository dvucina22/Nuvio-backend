package models

type AttributeFilter struct {
	AttributeID int64   `json:"attributeId"`
	ValueIDs    []int64 `json:"valueIds"`
}

type ProductFilter struct {
	Search      *string `json:"search,omitempty"`
	BrandIDs    []int64 `json:"brandIds,omitempty"`
	CategoryIDs []int64 `json:"categoryIds,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`

	AttributeFilters []AttributeFilter `json:"attributeFilters,omitempty"`

	VariantAttributeFilters []AttributeFilter `json:"variantAttributeFilters,omitempty"`

	PriceMin *float64 `json:"priceMin,omitempty"`
	PriceMax *float64 `json:"priceMax,omitempty"`

	Sort   string `json:"sort,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}
