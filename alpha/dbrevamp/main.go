package main

import (
	"fmt"
	"hermes/server"
)

func main() {
	err := os.Remove("sqlite-database.db") // I delete the file to avoid duplicated records.  // SQLite is a file based database.
	if err != nil {
		log.Printf("Failed delete " + err.Error())
	}

	fmt.Println("Creating sqlite-database.db...")
	dbfile, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		log.Fatal("Create database file failed: " + err.Error())
	}
	dbfile.Close()
	fmt.Println("sqlite-database.db created")

	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	defer sqliteDatabase.Close()                                     // Defer Closing the database
	createUrlStatusTable(sqliteDatabase)

	device1 := server.NewDevice()
}
