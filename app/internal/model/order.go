package model

type SuccessOrder struct {
	OrderId   uint `json:"order_id"`
	ProductId uint `json:"product_id"`
}

type DecreasePointSuccess struct {
	OrderId    uint   `json:"order_id"`
	PointLevel string `json:"point_level"`
}
