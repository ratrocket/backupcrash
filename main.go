package main

import (
	"fmt"
	"log"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
)

const (
	DBNAME  = "database.db"
	BACKUP1 = "backup1.db"
	BACKUP2 = "backup2.db"

	createTable = `CREATE TABLE products(
		id    INTEGER PRIMARY KEY,
		sku   TEXT,
		title TEXT
	)`
)

// This is reimplimented as sqlutil.BackupTo.
//
// THIS is the solution to the crash illustrated by the commented out
// parts of this program and the extensive write up in README.md.
func backupVacuum(conn *sqlite.Conn, backup string) (err error) {
	stmtTxt := "VACUUM INTO ?"
	if err := sqlitex.Exec(conn, stmtTxt, nil, backup); err != nil {
		return err
	}
	return nil
}

func main() {
	srcpool, err := sqlitex.Open(DBNAME, 0, 10)
	if err != nil {
		log.Fatal(err)
	}
	defer srcpool.Close()
	srcconn := srcpool.Get(nil)
	defer srcpool.Put(srcconn)

	if err := sqlitex.ExecScript(srcconn, createTable); err != nil {
		log.Fatal(err)
	}

	insertProduct(srcconn, "BK", "Black T")
	insertProduct(srcconn, "WT", "White T")

	// Make sure there are 2 products.
	fmt.Println("src products count (should be 2):", selectCountStar(srcconn, "products"))

	// WAY ZERO
	if err := backupVacuum(srcconn, BACKUP2); err != nil {
		log.Print("Way zero failure")
		log.Fatal(err)
	}

	// WAY ONE
	// _, err = srcconn.BackupToDB(DBNAME, BACKUP1)
	// if err != nil {
	// 	log.Print("Way one failure")
	// 	log.Fatal(err)
	// }

	// WAY TWO
	// dstpool, err := sqlitex.Open(BACKUP2, 0, 10)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer dstpool.Close()
	// dstconn := dstpool.Get(nil)
	// defer dstpool.Put(dstconn)
	// bu, err := srcconn.BackupInit(DBNAME, BACKUP2, dstconn)
	// if err != nil {
	// 	log.Print("Way two, failure 1 (BackupInit)")
	// 	log.Fatal(err)
	// }
	// if err := bu.Step(-1); err != nil {
	// 	log.Print("Way two, failure 2 (Step)")
	// 	log.Fatal(err)
	// }
	// if err := bu.Finish(); err != nil {
	// 	log.Print("Way two, failure 3 (Finish)")
	// 	log.Fatal(err)
	// }

	// Post-backup, insert another product into non-backup DB.
	insertProduct(srcconn, "OR", "Orange T")

	fmt.Println("src products count (should be 3):", selectCountStar(srcconn, "products"))
	// fmt.Println("dst products count (should be 2):", selectCountStar(dstconn, "products"))
	// don't have a convenient conn to check backup2.db...
}

// These functions fail on error for simplicity.

func insertProduct(conn *sqlite.Conn, sku, title string) {
	stmtTxt := "INSERT INTO products(sku, title) VALUES (?, ?)"
	if err := sqlitex.Exec(conn, stmtTxt, nil, sku, title); err != nil {
		log.Fatal(err)
	}
}

func selectCountStar(conn *sqlite.Conn, table string) int64 {
	var count int64
	fn := func(stmt *sqlite.Stmt) error {
		count = stmt.ColumnInt64(0)
		return nil
	}
	stmt := fmt.Sprintf("SELECT count(*) FROM %s", table)
	if err := sqlitex.Exec(conn, stmt, fn); err != nil {
		log.Fatal(err)
	}
	return count
}
