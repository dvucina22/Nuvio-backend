package models

type ProductAttribute struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}
