package config

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "ashish-tripathi"
	password = "postgres"
	dbname   = "ashish-tripathi"
)

type DBconfig struct {
	DBHost   string `env:"HOST"`
	DBPort   int    `env:"PORT"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	DBName   string `env:"NAME"`
}

type AppConfig struct {
	DBconfig DBconfig `envPrefix:"DB_"`
	Host     string   `env:"APP_HOST"`
	Port     int      `env:"APP_PORT"`
}

func DBinit(dbconfig *DBconfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbconfig.DBHost, dbconfig.DBPort, dbconfig.User, dbconfig.Password, dbconfig.DBName)
	fmt.Println(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic("failed to create database connection!!!")
	}
	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("not able to establish a connection, Ping failed!!! %s", err.Error()))
	}
	log.Printf("DB Successfully connected")
	return db, err
}

func LoadConfiguration() *AppConfig {
	godotenv.Load("local.env")
	appconfig := new(AppConfig)
	if err := env.Parse(appconfig); err != nil {
		panic("Failed to load environment variables")
	}
	return appconfig
}
