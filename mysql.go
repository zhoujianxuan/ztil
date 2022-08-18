package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli/v2"
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

	//result := make(map[int]*DatabaseResult)
	var result []*DatabaseResult
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

		result = append(result, &DatabaseResult{
			Tp:    tp,
			Value: row,
		})
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
		UsageText: "mysql <user> <password> <url> <database> <sql> [-e]",
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

			var exhibit bool
			if c.NArg() == 6 {
				exhibit = true
			}

			list, cols, err := RAW(user, pass, url, database, sqlStr)
			if err != nil {
				return err
			}

			if exhibit {
				for _, col := range cols {
					fmt.Print(col, " ")
				}
				fmt.Println()
				for _, row := range list {
					for _, r := range row {
						fmt.Print(r, " ")
					}
					fmt.Println()
				}
				return nil
			}

			f, err := os.Create(fmt.Sprintf("%d%s", time.Now().UnixNano(), ".csv"))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_, err = f.WriteString("\xEF\xBB\xBF")
			if err != nil {
				return err
			}

			writer := csv.NewWriter(f)
			err = writer.Write(cols)
			if err != nil {
				return err
			}

			for _, data := range list {
				err = writer.Write(data)
				if err != nil {
					return err
				}
			}
			writer.Flush()
			return nil
		},
	}
}
