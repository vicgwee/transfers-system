package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"transfers/api"
	db "transfers/db/sqlc"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	dbSource := viper.GetString("dbSource")
	serverAddress := viper.GetString("serverAddress")
	
	poolConfig, err := pgxpool.ParseConfig(dbSource)
	if err != nil {
		log.Fatalln("Unable to parse dbSource:", err)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}
	defer pool.Close()
	store := db.NewPgxStore(pool)
	server := api.NewServer(store)
	err = server.Run(serverAddress)
	if err != nil {
		log.Fatal("Err when running server:", err)
	}
}
