package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/openhouse_2025")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("=== CLEANING UP TEST DATA ===")

	// Clean up test users
	result, err := db.Exec(`
		DELETE FROM users WHERE 
		nrp LIKE 'test%' OR 
		nrp LIKE 'debug%' OR 
		nrp LIKE 'seq%' OR 
		nrp LIKE 'con%' OR 
		nrp LIKE 'race%'
	`)
	if err != nil {
		log.Printf("Error cleaning up test users: %v", err)
	} else {
		affected, _ := result.RowsAffected()
		fmt.Printf("✅ Removed %d test users\n", affected)
	}

	// Clean up test reservations
	result, err = db.Exec(`
		DELETE FROM slot_reservations WHERE 
		nrp LIKE 'test%' OR 
		nrp LIKE 'debug%' OR 
		nrp LIKE 'seq%' OR 
		nrp LIKE 'con%' OR 
		nrp LIKE 'race%'
	`)
	if err != nil {
		log.Printf("Error cleaning up test reservations: %v", err)
	} else {
		affected, _ := result.RowsAffected()
		fmt.Printf("✅ Removed %d test reservations\n", affected)
	}

	// Reset LK PELMA back to normal (it was modified for testing)
	_, err = db.Exec(`
		UPDATE ukms SET current_slot = 101 
		WHERE name = 'LK PELMA' AND current_slot = 9998
	`)
	if err != nil {
		log.Printf("Error resetting LK PELMA: %v", err)
	} else {
		fmt.Printf("✅ Reset LK PELMA slot count\n")
	}

	fmt.Println("🧹 Database cleanup complete!")
}
