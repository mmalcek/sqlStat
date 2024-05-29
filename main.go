package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type tStats struct {
	connect time.Duration
	errors  []string
}

var (
	config tConfig
	stats  tStats
	db     *sqlx.DB
)

func main() {
	if err := getConfig(); err != nil {
		fmt.Println("Error getting config: " + err.Error())
		return
	}
	var err error
	timeStart := time.Now()
	if db, err = Connect(); err != nil {
		fmt.Println("Error creating connection pool: " + err.Error())
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Println("Error pinging database: " + err.Error())
		return
	}
	stats.connect = time.Since(timeStart)
	fmt.Println("Connected!")
	fmt.Println("Time taken: ", stats.connect)
	// Get queries from files
	queries, err := getQueries("./queries")
	if err != nil {
		fmt.Println("Error loading queries: " + err.Error())
		return
	}

	if err := createExcel(queries); err != nil {
		fmt.Println("Error creating excel: " + err.Error())
		return
	}
}
