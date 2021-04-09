package main

import (
	"log"
)

func main() {
	log.Println("Opening DB")
	db, err := OpenSQLDB("sqlite3", "fizzboxd.db")
	if err != nil {
		log.Fatalf("failed to open database: %v\n", err)
	}

	log.Println("Closing DB")
	db.Close()

	log.Println("bye")
}
