package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load env vars from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found. Using system env")
	}

	// Create config using DB info
	db_user := os.Getenv("DB_USER")
	db_password := os.Getenv("DB_PASSWD")
	db_host := os.Getenv("DB_HOST")
	db_port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	db_name := os.Getenv("DB_NAME")

	// Create the connection URL
	conn_url := conn_url(db_user, db_password, db_host, db_port, db_name)
	log.Println("Connection url:", conn_url)

	// Connect using the connection URL
	db, err := sql.Open("pgx", conn_url)
	if err != nil {
		log.Fatal(err)
	}
	// Defer the connection closing to when the function closes
	// Common in functions that represent the whole execution of the program
	defer db.Close()

	wait_for_db(db)
}

func conn_url(user, passwd, host string, port int, db string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, passwd, host, port, db)
}

func wait_for_db(db *sql.DB) {
	for {
		err := db.Ping()
		if err == nil {
			log.Println("Connected to database.")
			return
		}
		log.Println(err)
		log.Println("Waiting for database, retrying in 1s...")
		time.Sleep(time.Second)
	}
}
