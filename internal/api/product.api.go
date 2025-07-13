package api

import (
	"context"
	"go-fiber-restapi/domain"
	"go-fiber-restapi/dto"
	"go-fiber-restapi/internal/utility"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type productApi struct {
	productService domain.ProductService
}

func NewProduct(app *fiber.App, productService domain.ProductService) {
	c := productApi{
		productService: productService,
	}

	app.Get("/products", c.Index)
	app.Post("/products", c.Create)
	app.Put("/products/:id", c.Update)
}

func (c productApi) Index(ctx *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	res, err := c.productService.GetAll(ctxWithTimeout)

	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.JSON(dto.CreateResponseSuccess(res))
}

func (c productApi) Create(ctx *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.CreateProductRequest
	
	if err:= ctx.BodyParser(&req); err != nil{
		return ctx.SendStatus(http.StatusUnprocessableEntity)
	}

	fails := utility.Validate(req)
	if len(fails) > 0 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.CreateResponseErrorData("validation error", fails))
	}

	err := c.productService.Create(ctxWithTimeout, req)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.Status(http.StatusCreated).JSON(dto.CreateResponseSuccess(req))
}

func (c productApi) Update(ctx *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctx.Context(), 10*time.Second)
	defer cancel()

	var req dto.UpdateProductRequest
	
	if err:= ctx.BodyParser(&req); err != nil{
		return ctx.SendStatus(http.StatusUnprocessableEntity)
	}

	fails := utility.Validate(req)
	if len(fails) > 0 {
		return ctx.Status(http.StatusBadRequest).JSON(dto.CreateResponseErrorData("validation error", fails))
	}

	req.ID = ctx.Params("id")
	err := c.productService.Update(ctxWithTimeout, req)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(dto.CreateResponseError(err.Error()))
	}

	return ctx.Status(http.StatusOK).JSON(dto.CreateResponseSuccess(req))
}
