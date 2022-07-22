package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"

	_ "github.com/go-sql-driver/mysql"
)

func getMysqlConnection(user, pass, url, database string) (*sql.DB, error) {
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pass, url, database))
}

type DatabaseResult struct {
	Tp    map[string]string
	Value map[string][]byte
}

func RAW(user, pass, url, database, sqlStr string) ([][]string, []string, error) {
	db, err := getMysqlConnection(user, pass, url, database)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close()

	rows, err := db.Query(sqlStr)
	if err != nil {
		return nil, nil, err
	}

	cols, _ := rows.Columns()
	types, _ := rows.ColumnTypes()

	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i, _ := range cols {
		scans[i] = &values[i]
	}

	index := 0
	result := make(map[int]*DatabaseResult)
	for rows.Next() {
		err = rows.Scan(scans...)
		if err != nil {
			log.Fatal(err)
		}

		tp := make(map[string]string)
		row := make(map[string][]byte)
		j := 0
		for k, v := range values {
			key := cols[k]
			//这里把[]byte根据条件转换
			row[key] = v
			tp[key] = types[j].DatabaseTypeName()
			j++
		}

		result[index] = &DatabaseResult{
			Tp:    tp,
			Value: row,
		}
		index++
	}

	var data [][]string
	for i := range result {
		row := result[i]
		var rowData []string
		for _, col := range cols {
			if v, ok := row.Value[col]; ok {
				rowData = append(rowData, string(v))
			}
		}
		data = append(data, rowData)
	}
	return data, cols, nil
}

func NewMySQLCommand() *cli.Command {
	return &cli.Command{
		Name:      "mysql",
		Aliases:   []string{"sql"},
		Usage:     "select sql",
		UsageText: "mysql <user> <password> <url> <database> <sql>",
		Action: func(c *cli.Context) error {
			if c.NArg() < 5 {
				fmt.Println("Periodic incomplete")
				return nil
			}

			user := c.Args().Get(0)
			pass := c.Args().Get(1)
			url := c.Args().Get(2)
			database := c.Args().Get(3)
			sqlStr := c.Args().Get(4)

			list, cols, err := RAW(user, pass, url, database, sqlStr)
			if err != nil {
				return err
			}

			f, err := os.Create(fmt.Sprintf("%d%s", time.Now().UnixNano(), ".csv"))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_, _ = f.WriteString("\xEF\xBB\xBF")
			writer := csv.NewWriter(f)
			_ = writer.Write(cols)
			for _, data := range list {
				_ = writer.Write(data)
			}
			writer.Flush()
			return nil
		},
	}
}
