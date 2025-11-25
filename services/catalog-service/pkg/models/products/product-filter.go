package products

type AttributeFilter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type ProductFilter struct {
	Search      *string `json:"search,omitempty"`
	BrandIDs    []int64 `json:"brandIds,omitempty"`
	CategoryIDs []int64 `json:"categoryIds,omitempty"`
	IsActive    *bool   `json:"isActive,omitempty"`

	Attributes []AttributeFilter `json:"attributes,omitempty"`

	PriceMin *float64 `json:"priceMin,omitempty"`
	PriceMax *float64 `json:"priceMax,omitempty"`

	Sort   string `json:"sort,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`

	IsFavorite       *bool   `json:"isFavorite,omitempty"`
	IsFavoriteUserID *string `json:"-"`
}
