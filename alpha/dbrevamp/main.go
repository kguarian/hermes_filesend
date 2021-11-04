package main

import (
	"database/sql"
	"fmt"
	"hermes/server"
	"log"
	"net"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

func main() {

	var sqliteDatabase *sql.DB
	sqliteDatabase = InitiateEverything()

	defer sqliteDatabase.Close() // Defer Closing the database
	var devarray []server.Device = make([]server.Device, 1)

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

	sqliteDatabase, err := sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	defer sqliteDatabase.Close() // Defer Closing the database
	createUrlStatusTable(sqliteDatabase)

	device1, err := server.NewDevice("';drop", "", net.ParseIP(server.IP_server))
	if err != nil {
		log.Panicf("Error! Line 34!\n")
	}
	devarray[0] = device1
	DB_InsertDeviceSlice(sqliteDatabase, device1.Userid, devarray)
	sl, err := DB_GetDeviceSlice(sqliteDatabase, device1.Userid)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	fmt.Printf("sql retval: %v\n", sl)
}
