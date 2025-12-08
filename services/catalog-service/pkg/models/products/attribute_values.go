package products

type AttributeItem struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}

type AttributeValues struct {
	Name  string          `json:"name"`
	Items []AttributeItem `json:"items"`
}
