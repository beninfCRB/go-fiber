package repository

import (
	"context"
	"database/sql"
	"go-fiber-restapi/domain"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type productRepository struct {
	db *goqu.Database
}


func NewProduct(con *sql.DB) domain.ProductRepository{
	return &productRepository{
		db : goqu.New("default",con),
	}
}

func (c *productRepository) FindAll(ctx context.Context) (result []domain.Product,err error){
	dataset := c.db.From("products").Where(goqu.C("deleted_at").IsNull())

	err= dataset.ScanStructsContext(ctx,&result)
	return
}

func (c *productRepository) FindById(ctx context.Context, id string) (result domain.Product,err error){
	dataset := c.db.From("products").Where(goqu.C("deleted_at").IsNull(),goqu.C("id").Eq(id))

	_,err = dataset.ScanStructContext(ctx,&result)
	return
}

func (c *productRepository) Save(ctx context.Context, product *domain.Product) error{
	executor := c.db.Insert("products").Rows(product).Executor()

	_,err := executor.ExecContext(ctx)
	return err
}

func (c *productRepository) Update(ctx context.Context, product *domain.Product) error{
	executor := c.db.Update("products").Set(product).Where(goqu.C("id").Eq(product.ID)).Executor()

	_,err := executor.ExecContext(ctx)
	return err
}

func (c *productRepository) Delete(ctx context.Context, id string) error{
	executor := c.db.Update("products").Set(goqu.Record{"deleted_at": sql.NullTime{Time: time.Now(), Valid: true}}).Where(goqu.C("id").Eq(id)).Executor()

	_,err := executor.ExecContext(ctx)
	return err
}

