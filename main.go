package main

import (
	"go-fiber-restapi/internal/api"
	"go-fiber-restapi/internal/config"
	"go-fiber-restapi/internal/connection"
	"go-fiber-restapi/internal/repository"
	"go-fiber-restapi/internal/service"

	"github.com/gofiber/fiber/v2"
)

func main(){
	cnf := config.Get()
	dbConnection := connection.GetDatabase(cnf.Database)

	app := fiber.New()

	productRepository := repository.NewProduct(dbConnection)
	productService := service.NewProduct(productRepository)
	api.NewProduct(app, productService)

	app.Get("/hello-world", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	_ =app.Listen(cnf.Server.Host + ":" + cnf.Server.Port)
}