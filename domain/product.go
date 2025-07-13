package domain

import (
	"context"
	"database/sql"
	"go-fiber-restapi/dto"
)

type Product struct {
	ID  string `db:"id"`
	Code string `db:"code"`
	Name string `db:"name"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

type ProductRepository interface {
	FindAll(ctx context.Context) ([]Product, error)
	FindById(ctx context.Context, id string) (Product, error)
	Save(ctx context.Context, product *Product) error
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}

type ProductService interface{
	GetAll(ctx context.Context) ([]dto.ProductData,error)
	Create(ctx context.Context,req dto.CreateProductRequest) error
	Update(ctx context.Context,req dto.UpdateProductRequest) error
}