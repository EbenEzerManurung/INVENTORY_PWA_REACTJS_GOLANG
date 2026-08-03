package migrations

import (
	"database/sql"
	"fmt"
	"log"
	
)

// Migration represents a database migration
type Migration struct {
	DB *sql.DB
}

// NewMigration creates a new migration instance
func NewMigration(db *sql.DB) *Migration {
	return &Migration{DB: db}
}

// Run executes all migrations
func (m *Migration) Run() error {
	log.Println("🚀 Running migrations...")

	// Create tables in correct order (with foreign key dependencies)
	migrations := []string{
		m.createUsersTable(),
		m.createProductsTable(),
		m.createStockTable(),
		m.createTransactionInTable(),
		m.createTransactionOutTable(),
	}

	for _, migration := range migrations {
		if _, err := m.DB.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %v", err)
		}
	}

	log.Println("✅ Migrations completed successfully")
	return nil
}

// Reset drops all tables and recreates them
func (m *Migration) Reset() error {
	log.Println("🔄 Resetting database...")

	// Drop tables in reverse order (to handle foreign key constraints)
	tables := []string{
		"transaction_out",
		"transaction_in",
		"stock",
		"products",
		"users",
	}

	// Disable foreign key checks temporarily
	if _, err := m.DB.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %v", err)
	}
	defer m.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
		if _, err := m.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to drop table %s: %v", table, err)
		}
		log.Printf("✅ Dropped table: %s", table)
	}

	// Run migrations
	if err := m.Run(); err != nil {
		return err
	}

	log.Println("✅ Database reset completed successfully")
	return nil
}

// ==================== TABLE DEFINITIONS ====================

// createUsersTable returns SQL for users table
func (m *Migration) createUsersTable() string {
	return `
	CREATE TABLE IF NOT EXISTS users (
		id INT(11) NOT NULL AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL,
		password VARCHAR(255) NOT NULL,
		fullname VARCHAR(100) NOT NULL,
		email VARCHAR(100) DEFAULT NULL,
		role ENUM('superadmin','head','produksi') NOT NULL,
		profile_image VARCHAR(255) DEFAULT NULL,
		avatar TEXT DEFAULT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		last_login TIMESTAMP NULL DEFAULT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY username (username),
		UNIQUE KEY email (email)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
	`
}

// createProductsTable returns SQL for products table
func (m *Migration) createProductsTable() string {
	return `
	CREATE TABLE IF NOT EXISTS products (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_code VARCHAR(50) NOT NULL,
		product_name VARCHAR(100) NOT NULL,
		category VARCHAR(50) DEFAULT NULL,
		unit VARCHAR(20) DEFAULT NULL,
		min_stock INT(11) DEFAULT 0,
		description TEXT DEFAULT NULL,
		image TEXT DEFAULT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY product_code (product_code),
		KEY idx_products_name (product_name),
		KEY idx_products_code (product_code)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
	`
}

// createStockTable returns SQL for stock table
func (m *Migration) createStockTable() string {
	return `
	CREATE TABLE IF NOT EXISTS stock (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_id INT(11) DEFAULT NULL,
		quantity INT(11) DEFAULT 0,
		last_updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY product_id (product_id),
		CONSTRAINT stock_ibfk_1 FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
	`
}

// createTransactionInTable returns SQL for transaction_in table
func (m *Migration) createTransactionInTable() string {
	return `
	CREATE TABLE IF NOT EXISTS transaction_in (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_id INT(11) DEFAULT NULL,
		quantity INT(11) NOT NULL,
		note TEXT DEFAULT NULL,
		created_by INT(11) DEFAULT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY created_by (created_by),
		KEY idx_transaction_in_product (product_id),
		CONSTRAINT transaction_in_ibfk_1 FOREIGN KEY (product_id) REFERENCES products (id),
		CONSTRAINT transaction_in_ibfk_2 FOREIGN KEY (created_by) REFERENCES users (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
	`
}

// createTransactionOutTable returns SQL for transaction_out table
func (m *Migration) createTransactionOutTable() string {
	return `
	CREATE TABLE IF NOT EXISTS transaction_out (
		id INT(11) NOT NULL AUTO_INCREMENT,
		product_id INT(11) DEFAULT NULL,
		quantity INT(11) NOT NULL,
		note TEXT DEFAULT NULL,
		status ENUM('pending','approved','rejected') DEFAULT 'pending',
		created_by INT(11) DEFAULT NULL,
		approved_by INT(11) DEFAULT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		approved_at TIMESTAMP NULL DEFAULT NULL,
		PRIMARY KEY (id),
		KEY created_by (created_by),
		KEY approved_by (approved_by),
		KEY idx_transaction_out_product (product_id),
		KEY idx_transaction_out_status (status),
		CONSTRAINT transaction_out_ibfk_1 FOREIGN KEY (product_id) REFERENCES products (id),
		CONSTRAINT transaction_out_ibfk_2 FOREIGN KEY (created_by) REFERENCES users (id),
		CONSTRAINT transaction_out_ibfk_3 FOREIGN KEY (approved_by) REFERENCES users (id)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
	`
}

// ==================== HELPER FUNCTIONS ====================

// TableExists checks if a table exists in the database
func (m *Migration) TableExists(tableName string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`
	err := m.DB.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetTableNames returns all table names in the database
func (m *Migration) GetTableNames() ([]string, error) {
	rows, err := m.DB.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

// TruncateTable truncates a table (removes all data but keeps structure)
func (m *Migration) TruncateTable(tableName string) error {
	query := fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	_, err := m.DB.Exec(query)
	return err
}

// ResetSequences resets auto-increment sequences for all tables
func (m *Migration) ResetSequences() error {
	tables, err := m.GetTableNames()
	if err != nil {
		return err
	}

	for _, table := range tables {
		query := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table)
		if _, err := m.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to reset sequence for %s: %v", table, err)
		}
	}
	return nil
}