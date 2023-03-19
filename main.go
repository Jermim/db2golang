package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func Throw(err error) {
	panic(err)
}

func main() {
	var destDir string
	var mysqlConnectionString string

	flag.StringVar(&destDir, "dir", "", "--dir destination folder")
	flag.StringVar(&mysqlConnectionString, "mysql", "", "--mysql user:pass@/db")
	flag.Parse()

	if destDir == "" {
		fmt.Println("--dir option is required")
		return
	}

	if fs, err := os.Stat(destDir); err != nil {
		fmt.Printf("could not locate directory: %s\n", destDir)
		return
	} else if !fs.IsDir() {
		fmt.Printf("%s does not seem to be a directory\n", destDir)
		return
	}

	if mysqlConnectionString == "" {
		fmt.Printf("--mysql option is required\n")
		return
	}

	db, err := sql.Open("mysql", mysqlConnectionString)

	if err != nil {
		Throw(err)
	}

	defer db.Close()

	rows, err := db.Query("show tables")

	if err != nil {
		Throw(err)
	}

	defer rows.Close()

	var tableName string

	for rows.Next() {
		if err := rows.Scan(&tableName); err != nil {
			Throw(err)
		}

		infos, err := db.Query(fmt.Sprintf("describe %s", tableName))

		if err != nil {
			Throw(err)
		}

		var Table Table = Table{TableName: tableName}

		for infos.Next() {
			var col Column

			if err := infos.Scan(&col.Name, &col.Type, &col.Nullable, &col.Key, &col.f5, &col.increment); err != nil {
				Throw(err)
			}

			Table.Columns = append(Table.Columns, &col)
		}

		if err := Table.Generate(destDir); err != nil {
			Throw(err)
		}
	}

}
