package model

type SuccessOrder struct {
	OrderId   uint    `json:"order_id"`
	ProductId uint    `json:"product_id"`
	Price     float64 `json:"price"`
}

type DecreasePointSuccess struct {
	OrderId    uint   `json:"order_id"`
	PointLevel string `json:"point_level"`
}
