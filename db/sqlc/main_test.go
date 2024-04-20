package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

var testStore Store

func TestMain(m *testing.M) {
	viper.AddConfigPath("../../.")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	testDbSource := viper.GetString("testDbSource")

	poolConfig, err := pgxpool.ParseConfig(testDbSource)
	if err != nil {
		log.Fatalln("Unable to parse testDbSource:", err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to create connection pool:", err)
	}
	defer conn.Close()

	testStore = NewPgxStore(conn)
	os.Exit(m.Run())
}
