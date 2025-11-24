package models

type ProductAttribute struct {
	AttributeID int64  `json:"attributeId"`
	Name        string `json:"name"`
	ValueID     int64  `json:"valueId"`
	Value       string `json:"value"`
}
