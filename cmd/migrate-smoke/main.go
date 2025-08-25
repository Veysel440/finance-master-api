package migrate_smoke

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/finance_master?parseTime=true&multiStatements=true"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// audit_logs var mı?
	if !existsTable(db, "audit_logs") {
		log.Fatal("missing table audit_logs")
	}

	// FK'ler var mı?
	if !existsFK(db, "transactions", "fk_tx_wallet") {
		log.Fatal("missing FK fk_tx_wallet")
	}
	if !existsFK(db, "transactions", "fk_tx_category") {
		log.Fatal("missing FK fk_tx_category")
	}

	// indeksler var mı?
	if !existsIndex(db, "transactions", "idx_tx_user_date_type") {
		log.Fatal("missing index idx_tx_user_date_type")
	}

	log.Println("schema smoke OK")
}

func existsTable(db *sql.DB, table string) bool {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=?`, table).Scan(&n)
	return n > 0
}
func existsFK(db *sql.DB, table, fk string) bool {
	var n int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.REFERENTIAL_CONSTRAINTS
		WHERE CONSTRAINT_SCHEMA=DATABASE() AND CONSTRAINT_NAME=? AND TABLE_NAME=?`, fk, table).Scan(&n)
	return n > 0
}
func existsIndex(db *sql.DB, table, idx string) bool {
	var n int
	_ = db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME=? AND INDEX_NAME=?`, table, idx).Scan(&n)
	return n > 0
}
