package connection

import (
	"database/sql"
	"fmt"
	"go-fiber-restapi/internal/config"
	"log"

	_ "github.com/lib/pq"
)

func GetDatabase(conf config.Database) *sql.DB{
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable Timezone=%s",
		conf.Host, conf.Port, conf.User, conf.Password, conf.Name,conf.TZ)

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal("failed to connect database",err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("failed to ping database",err.Error())
	}

	return db
}