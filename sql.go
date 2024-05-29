package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
	"gopkg.in/yaml.v3"
)

type (
	tQuery struct {
		cathegory string // data, duration, ...
		Name      string `yaml:"name"`
		Repeat    int    `yaml:"repeat"`
		Query     string `yaml:"query"`
	}

	tQueryDetails struct {
		timeRequest  time.Time
		timeResponse time.Time
		duration     time.Duration
	}
)

// Connect to MSSQL
func Connect() (*sqlx.DB, error) {
	query := url.Values{}
	query.Add("database", config.Database)
	if config.User == "" {
		query.Add("Trusted_Connection", "True")
	}
	if config.Port == 0 && config.Instance == "" {
		config.Port = 1433
	}

	if config.Instance == "" {
		u := &url.URL{
			Scheme:   "sqlserver",
			User:     url.UserPassword(config.User, config.Password),
			Host:     fmt.Sprintf("%s:%d", config.Server, config.Port),
			RawQuery: query.Encode(),
		}
		return sqlx.Open("sqlserver", u.String())
	}
	u := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(config.User, config.Password),
		Host:     config.Server,
		Path:     config.Instance,
		RawQuery: query.Encode(),
	}
	return sqlx.Open("sqlserver", u.String())
}

func getQueries(folder string) ([]tQuery, error) {
	queries := []tQuery{}
	cathegories := []string{"data", "duration"}
	for _, cathegory := range cathegories {
		files, err := os.ReadDir(filepath.Join(folder, cathegory))
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			file, err := os.ReadFile(filepath.Join(folder, cathegory, file.Name()))
			if err != nil {
				return nil, err
			}
			var query tQuery
			query.cathegory = cathegory
			if err := yaml.Unmarshal(file, &query); err != nil {
				return nil, err
			}
			queries = append(queries, query)
		}
	}
	return queries, nil
}

func queryData(query tQuery) (results []map[string]interface{}, columns []string, err error) {
	results = make([]map[string]interface{}, 0)
	columns = make([]string, 0)
	rows, err := db.Queryx(query.Query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	if columns, err = rows.Columns(); err != nil {
		return nil, nil, err
	}

	for rows.Next() {
		result := make(map[string]interface{})
		if err := rows.MapScan(result); err != nil {
			return nil, nil, err
		}
		results = append(results, result)

	}
	return results, columns, nil
}

func queryDurations(query tQuery) ([]tQueryDetails, error) {
	qds := []tQueryDetails{}
	for i := 0; i <= query.Repeat; i++ {
		qd := tQueryDetails{}
		fmt.Printf("Executing - repeat: %d, query: %s\r\n", i, query.Name)
		qd.timeRequest = time.Now()
		rows, err := db.Query(query.Query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		qd.timeResponse = time.Now()
		var results []string
		for rows.Next() {
			rows.Scan(results) // just to consume the results
		}
		// fmt.Println("Time taken: ", time.Since(timeStart))
		qd.duration = time.Since(qd.timeRequest)
		qds = append(qds, qd)
	}
	return qds, nil
}
