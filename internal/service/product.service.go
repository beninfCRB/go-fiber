package service

import (
	"context"
	"database/sql"
	"errors"
	"go-fiber-restapi/domain"
	"go-fiber-restapi/dto"
	"time"

	"github.com/google/uuid"
)

type productService struct{
	productRepository domain.ProductRepository
}

func NewProduct(productRepository domain.ProductRepository) domain.ProductService{
	return &productService{
		productRepository: productRepository,
	}
}

func (c productService) GetAll(ctx context.Context) ([]dto.ProductData,error){
	products,err:=c.productRepository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var productData []dto.ProductData
	for _, v := range products {
		productData = append(productData, dto.ProductData{
			ID: v.ID,
			Code: v.Code,
			Name: v.Name,
		})
	}

	return productData, nil
}

func (c productService) Create(ctx context.Context,req dto.CreateProductRequest) error{
	product := domain.Product{
		ID: uuid.NewString(),
		Code: req.Code,
		Name: req.Name,
		CreatedAt: sql.NullTime{Time: time.Now(),Valid: true},
	}

	return c.productRepository.Save(ctx,&product)
}

func (c productService) Update(ctx context.Context,req dto.UpdateProductRequest) error{
	persisted,err := c.productRepository.FindById(ctx,req.ID)
	if err != nil{
		return err
	}
	if persisted.ID == ""{
		return errors.New("data product tidak ditemukan")
	}
	persisted.Code = req.Code
	persisted.Name = req.Name
	persisted.UpdatedAt = sql.NullTime{Time: time.Now(),Valid: true}

	return c.productRepository.Update(ctx,&persisted)
}