package dto

type ProductData struct{
	ID string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateProductRequest struct{
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type UpdateProductRequest struct{
	ID string `json:"-"`
	Code string `json:"code" validate:"required"`
	Name string `json:"name" validate:"required"`
}