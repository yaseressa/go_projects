package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)


var DB *sql.DB

func initDBFile(dbName string){
	var err error
	DB, err = sql.Open("sqlite3", dbName + ".db")
	if err != nil {
		log.Fatal(err)
	}
}

func createTable(tableName string, attributeStructure string){
	sqlStmt := "CREATE TABLE IF NOT EXISTS "+tableName+"("+attributeStructure+");"
	_, err := DB.Exec(sqlStmt)
	if err != nil{
		log.Fatalf("Error creating table: %q: %s\n", err, tableName)
	}
}

