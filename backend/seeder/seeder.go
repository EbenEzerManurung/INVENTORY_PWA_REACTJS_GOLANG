package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// ==================== KONSTANTA ====================

const fallbackImage = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

const (
	blueImage   = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAAAXNSR0IArs4c6QAAAARzQklUCAgICHwIZAAAAAlwSFlzAAAOxAAADsQBlSsOGwAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAGQSURBVHic7dqxTQJhGMfxC3EAOzt2MDE2YhsTsbGxMQFjbISNDRhjYwLC2FhYWVhZWBhbEBtLAxsLYmNjbKg7zBx3uY9e7uWv9z1Jnmx37DfP83sP5/jnn3/++eef/ywR9//+HxNARISRkBHhBOGW0fP6jOqAepm0M+qHeANr9FZ7vQaVz9e9XkNOJm1pB+P/4Hh93feUqkAVVNGrGta2Q+wYVZCVVh1BtlVtyHrKOFTt/7n2P6Y/7sL1hX9z/3t/V+uy5brbc9PyZSUj04aM+QhbmKjP9SLbZ5d2VZ4j2z5VQy5yBc5knprJkKNbrMZ4inPTZPI4pnczfYUNj9bPzrSMG8m15flKxy2N7mUlI5OaiRw7hzJZ+9jG7/x44EfnyuKzjEwaMtZJymTuYQuH6WLb+J0fD/zoXFh8rpFJTUY6uT2CbTbQJkS+7cz5A9vF39P+EHsSX9/k9l2YztH5fuqM8/kd0wX6trldV1cnftYx6ZpN3zK3J/HXD77nml9sn/ghK/mJcBPj+SbnRSmZtGSUfP2L2zzNurLhVXQ/+6wrOz1Tv9Jk8hAx+Zm0n2dX9sHpPLiyH05bBhQNcFfUAEpYFOHW2T8QIgjJSPJKKIkDzYzMyp75jS7krKTVgWZGZmXP/IYEgYQBxsyMkJA40MzIrOyZ3+jCf2hiZIQQCUkDzUbKjIj4wY9MYfVbnCgRREQYCRkRThBuGT2vz6gOqJdJO6N+iDewRm+116tQ+Xzd6zXkZNKWdjD+D47X131PqQpUQRW9qmFtO8SOUQVZadURZFvVhqynjEPV/p9r/2P64y5cX/g39//vTVu2W667PTctX1YyMm3ImI+whYn6XC+yfXZpV+U5su1TNeQiV+BM5qmZDDm6xWqMpzg3TSaPY3o30/+A69qj/9FgSg/+j9qKXDYrGRkk/m3I2ilv42fHEpPfRS5H4t+GvHwJTw+oIrqspwIWAAAAAElFTkSuQmCC"
	greenImage  = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAAAXNSR0IArs4c6QAAAARzQklUCAgICHwIZAAAAAlwSFlzAAAOxAAADsQBlSsOGwAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAGQSURBVHic7dqxTQJhGMfxC3EAOzt2MDE2YhsTsbGxMQFjbISNDRhjYwLC2FhYWVhZWBhbEBtLAxsLYmNjbKg7zBx3uY9e7uWv9z1Jnmx37DfP83sP5/jnn3/++eef/ywR9//+HxNARISRkBHhBOGW0fP6jOqAepm0M+qHeANr9FZ7vQaVz9e9XkNOJm1pB+P/4Hh93feUqkAVVNGrGta2Q+wYVZCVVh1BtlVtyHrKOFTt/7n2P6Y/7sL1hX9z/3t/V+uy5brbc9PyZSUj04aM+QhbmKjP9SLbZ5d2VZ4j2z5VQy5yBc5knprJkKNbrMZ4inPTZPI4pnczfYUNj9bPzrSMG8m15flKxy2N7mUlI5OaiRw7hzJZ+9jG7/x44EfnyuKzjEwaMtZJymTuYQuH6WLb+J0fD/zoXFh8rpFJTUY6uT2CbTbQJkS+7cz5A9vF39P+EHsSX9/k9l2YztH5fuqM8/kd0wX6trldV1cnftYx6ZpN3zK3J/HXD77nml9sn/ghK/mJcBPj+SbnRSmZtGSUfP2L2zzNurLhVXQ/+6wrOz1Tv9Jk8hAx+Zm0n2dX9sHpPLiyH05bBhQNcFfUAEpYFOHW2T8QIgjJSPJKKIkDzYzMyp75jS7krKTVgWZGZmXP/IYEgYQBxsyMkJA40MzIrOyZ3+jCf2hiZIQQCUkDzUbKjIj4wY9MYfVbnCgRREQYCRkRThBuGT2vz6gOqJdJO6N+iDewRm+116tQ+Xzd6zXkZNKWdjD+D47X131PqQpUQRW9qmFtO8SOUQVZadURZFvVhqynjEPV/p9r/2P64y5cX/g39//vTVu2W667PTctX1YyMm3ImI+whYn6XC+yfXZpV+U5su1TNeQiV+BM5qmZDDm6xWqMpzg3TSaPY3o30/+A69qj/9FgSg/+j9qKXDYrGRkk/m3I2ilv42fHEpPfRS5H4t+GvHwJTw+oIrqspwIWAAAAAElFTkSuQmCC"
	redImage    = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyCAYAAAAeP4ixAAAAAXNSR0IArs4c6QAAAARzQklUCAgICHwIZAAAAAlwSFlzAAAOxAAADsQBlSsOGwAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAGQSURBVHic7dqxTQJhGMfxC3EAOzt2MDE2YhsTsbGxMQFjbISNDRhjYwLC2FhYWVhZWBhbEBtLAxsLYmNjbKg7zBx3uY9e7uWv9z1Jnmx37DfP83sP5/jnn3/++eef/ywR9//+HxNARISRkBHhBOGW0fP6jOqAepm0M+qHeANr9FZ7vQaVz9e9XkNOJm1pB+P/4Hh93feUqkAVVNGrGta2Q+wYVZCVVh1BtlVtyHrKOFTt/7n2P6Y/7sL1hX9z/3t/V+uy5brbc9PyZSUj04aM+QhbmKjP9SLbZ5d2VZ4j2z5VQy5yBc5knprJkKNbrMZ4inPTZPI4pnczfYUNj9bPzrSMG8m15flKxy2N7mUlI5OaiRw7hzJZ+9jG7/x44EfnyuKzjEwaMtZJymTuYQuH6WLb+J0fD/zoXFh8rpFJTUY6uT2CbTbQJkS+7cz5A9vF39P+EHsSX9/k9l2YztH5fuqM8/kd0wX6trldV1cnftYx6ZpN3zK3J/HXD77nml9sn/ghK/mJcBPj+SbnRSmZtGSUfP2L2zzNurLhVXQ/+6wrOz1Tv9Jk8hAx+Zm0n2dX9sHpPLiyH05bBhQNcFfUAEpYFOHW2T8QIgjJSPJKKIkDzYzMyp75jS7krKTVgWZGZmXP/IYEgYQBxsyMkJA40MzIrOyZ3+jCf2hiZIQQCUkDzUbKjIj4wY9MYfVbnCgRREQYCRkRThBuGT2vz6gOqJdJO6N+iDewRm+116tQ+Xzd6zXkZNKWdjD+D47X131PqQpUQRW9qmFtO8SOUQVZadURZFvVhqynjEPV/p9r/2P64y5cX/g39//vTVu2W667PTctX1YyMm3ImI+whYn6XC+yfXZpV+U5su1TNeQiV+BM5qmZDDm6xWqMpzg3TSaPY3o30/+A69qj/9FgSg/+j9qKXDYrGRkk/m3I2ilv42fHEpPfRS5H4t+GvHwJTw+oIrqspwIWAAAAAElFTkSuQmCC"
)

var categories = []string{
	"Electronics", "Accessories", "Office", "Stationery",
	"Furniture", "Tools", "Materials", "Packaging",
	"Cleaning", "Safety", "Food", "Beverages",
}

var units = []string{
	"Pcs", "Box", "Rim", "Pack", "Set",
	"Kg", "Gram", "Liter", "Meter", "Unit",
}

var productNames = []string{
	"Laptop", "Mouse", "Keyboard", "Monitor", "Printer",
	"Paper", "Pen", "USB", "Hard Drive", "Webcam",
	"Chair", "Table", "Cabinet", "Shelf", "Light",
	"Cable", "Adapter", "Battery", "Charger", "Speaker",
	"Headset", "Microphone", "Camera", "Scanner", "Projector",
	"Router", "Switch", "Hub", "Modem", "Access Point",
	"Screwdriver", "Hammer", "Wrench", "Pliers", "Tape",
	"Glue", "Paint", "Brush", "Roller", "Sandpaper",
	"Box", "Bag", "Wrap", "Tape", "Stapler",
	"Soap", "Detergent", "Disinfectant", "Tissue", "Towel",
	"Helmet", "Gloves", "Mask", "Goggles", "Vest",
	"Coffee", "Tea", "Sugar", "Milk", "Water",
}

var adjectives = []string{
	"Premium", "Pro", "Ultra", "Max", "Elite",
	"Basic", "Standard", "Advanced", "Smart", "Digital",
	"Professional", "Industrial", "Commercial", "Heavy Duty", "Lightweight",
	"Portable", "Compact", "Ergonomic", "High Speed", "Energy Saving",
	"Waterproof", "Dustproof", "Shockproof", "Fireproof", "Rustproof",
}

// ==================== FUNGSI UNTUK RESET DATABASE ====================

func ResetDatabase(db *sql.DB) {
	fmt.Println("🗑️  Resetting all data...")

	// Nonaktifkan foreign key checks sementara
	_, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err != nil {
		log.Fatal("Failed to disable foreign key checks:", err)
	}
	defer db.Exec("SET FOREIGN_KEY_CHECKS = 1")

	// Hapus data dengan urutan yang aman
	tables := []string{
		"transaction_out",
		"transaction_in",
		"stock",
		"products",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Printf("⚠️  Failed to delete from %s: %v", table, err)
		} else {
			fmt.Printf("✅ Deleted all data from %s\n", table)
		}

		// Reset auto increment
		_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = 1", table))
		if err != nil {
			log.Printf("⚠️  Failed to reset auto increment for %s: %v", table, err)
		} else {
			fmt.Printf("✅ Reset auto increment for %s\n", table)
		}
	}

	fmt.Println("✅ Database reset completed")
}

// ==================== SEEDER FUNCTIONS ====================

func SeedUsers(db *sql.DB) {
	fmt.Println("\n📦 Seeding users...")

	users := []struct {
		Username string
		Password string
		Fullname string
		Role     string
	}{
		{"superadmin", "admin123", "Super Admin", "superadmin"},
		{"head", "admin123", "Department Head", "head"},
		{"produksi", "admin123", "Production Staff", "produksi"},
	}

	for _, u := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash password:", err)
		}

		_, err = db.Exec(`
			INSERT INTO users (username, password, fullname, role) 
			VALUES (?, ?, ?, ?)
		`, u.Username, hashedPassword, u.Fullname, u.Role)

		if err != nil {
			log.Fatal("Failed to insert user:", err)
		}
		fmt.Printf("✅ User %s created\n", u.Username)
	}
}

func SeedProducts(db *sql.DB) {
	fmt.Println("\n📦 Seeding 1000 products...")

	rand.Seed(time.Now().UnixNano())
	images := []string{blueImage, greenImage, redImage, fallbackImage}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	for i := 1; i <= 1000; i++ {
		code := fmt.Sprintf("PRD-%04d", i)
		adj := adjectives[rand.Intn(len(adjectives))]
		name := productNames[rand.Intn(len(productNames))]
		productName := fmt.Sprintf("%s %s", adj, name)
		category := categories[rand.Intn(len(categories))]
		unit := units[rand.Intn(len(units))]
		minStock := rand.Intn(46) + 5

		descs := []string{
			fmt.Sprintf("High quality %s for professional use", name),
			fmt.Sprintf("Premium %s with advanced features", name),
			fmt.Sprintf("Durable %s for daily use", name),
			fmt.Sprintf("Ergonomic %s designed for comfort", name),
			fmt.Sprintf("Efficient %s with great performance", name),
		}
		description := descs[rand.Intn(len(descs))]
		image := images[rand.Intn(len(images))]

		_, err := tx.Exec(`
			INSERT INTO products (product_code, product_name, category, unit, min_stock, description, image) 
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, code, productName, category, unit, minStock, description, image)

		if err != nil {
			tx.Rollback()
			log.Fatal("Failed to insert product:", err)
		}

		if i%100 == 0 {
			fmt.Printf("✅ %d products created\n", i)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}
	fmt.Println("✅ 1000 products created successfully")
}

func SeedStock(db *sql.DB) {
	fmt.Println("\n📦 Seeding stock for 1000 products...")

	rand.Seed(time.Now().UnixNano())

	rows, err := db.Query("SELECT id FROM products")
	if err != nil {
		log.Fatal("Failed to get product IDs:", err)
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatal("Failed to scan product ID:", err)
		}
		productIDs = append(productIDs, id)
	}

	if len(productIDs) == 0 {
		fmt.Println("⚠️  No products found, skipping stock seeding")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	for i, pid := range productIDs {
		quantity := rand.Intn(491) + 10

		_, err := tx.Exec(`
			INSERT INTO stock (product_id, quantity) 
			VALUES (?, ?)
		`, pid, quantity)

		if err != nil {
			tx.Rollback()
			log.Fatal("Failed to insert stock:", err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("✅ %d stock records created\n", i+1)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}
	fmt.Println("✅ 1000 stock records created successfully")
}

func SeedTransactions(db *sql.DB) {
	fmt.Println("\n📦 Seeding transactions...")

	rand.Seed(time.Now().UnixNano())

	var superadminUserID int
	err := db.QueryRow("SELECT id FROM users WHERE username = 'superadmin' LIMIT 1").Scan(&superadminUserID)
	if err != nil {
		log.Fatal("Failed to get superadmin user ID:", err)
	}

	var headUserID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'head' LIMIT 1").Scan(&headUserID)
	if err != nil {
		log.Fatal("Failed to get head user ID:", err)
	}

	var produksiUserID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'produksi' LIMIT 1").Scan(&produksiUserID)
	if err != nil {
		log.Fatal("Failed to get produksi user ID:", err)
	}

	rows, err := db.Query("SELECT id FROM products")
	if err != nil {
		log.Fatal("Failed to get product IDs:", err)
	}
	defer rows.Close()

	var productIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatal("Failed to scan product ID:", err)
		}
		productIDs = append(productIDs, id)
	}

	if len(productIDs) == 0 {
		fmt.Println("⚠️  No products found, skipping transaction seeding")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	fmt.Println("📦 Seeding 1000 transaction_in records...")
	for i := 0; i < 1000; i++ {
		pid := productIDs[rand.Intn(len(productIDs))]
		quantity := rand.Intn(196) + 5
		daysAgo := rand.Intn(30)
		date := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02 15:04:05")
		note := fmt.Sprintf("Stock entry - batch %d", i+1)

		userID := superadminUserID
		if rand.Intn(2) == 0 {
			userID = headUserID
		}

		_, err := tx.Exec(`
			INSERT INTO transaction_in (product_id, quantity, note, created_by, created_at) 
			VALUES (?, ?, ?, ?, ?)
		`, pid, quantity, note, userID, date)

		if err != nil {
			tx.Rollback()
			log.Fatal("Failed to insert transaction_in:", err)
		}

		_, err = tx.Exec(`
			UPDATE stock SET quantity = quantity + ? WHERE product_id = ?
		`, quantity, pid)

		if err != nil {
			tx.Rollback()
			log.Fatal("Failed to update stock:", err)
		}

		if (i+1)%100 == 0 {
			fmt.Printf("✅ %d transaction_in records created\n", i+1)
		}
	}

	fmt.Println("📦 Seeding 1000 transaction_out records...")
for i := 0; i < 1000; i++ {
    pid := productIDs[rand.Intn(len(productIDs))]
    quantity := rand.Intn(50) + 1

    // Status: approved, pending, rejected (tanpa blank)
    // Distribusi: 60% approved, 30% pending, 10% rejected
    status := "approved"
    statusRand := rand.Intn(10)
    if statusRand < 3 {
        status = "pending"
    } else if statusRand < 4 {
        status = "rejected"
    }

    daysAgo := rand.Intn(30)
    date := time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02 15:04:05")
    note := fmt.Sprintf("Stock request - batch %d", i+1)
    userID := produksiUserID

    _, err := tx.Exec(`
        INSERT INTO transaction_out (product_id, quantity, note, status, created_by, created_at) 
        VALUES (?, ?, ?, ?, ?, ?)
    `, pid, quantity, note, status, userID, date)

    if err != nil {
        tx.Rollback()
        log.Fatal("Failed to insert transaction_out:", err)
    }

    // Hanya kurangi stock jika status approved
    if status == "approved" {
        _, err = tx.Exec(`
            UPDATE stock SET quantity = quantity - ? WHERE product_id = ?
        `, quantity, pid)

        if err != nil {
            tx.Rollback()
            log.Fatal("Failed to update stock:", err)
        }
    }

    if (i+1)%100 == 0 {
        fmt.Printf("✅ %d transaction_out records created\n", i+1)
    }
}

	if err := tx.Commit(); err != nil {
		log.Fatal("Failed to commit transaction:", err)
	}
	fmt.Println("✅ 1000 transaction_in and 1000 transaction_out records created successfully")
}

// ==================== MAIN ====================

func main() {
	// Konfigurasi database - SESUAIKAN DENGAN DATABASE ANDA
	dsn := "root:@tcp(127.0.0.1:3306)/inventory_db?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("✅ Connected to database")
	fmt.Println("Database:", dsn)
	fmt.Println()

	// Tampilkan peringatan
	fmt.Println("⚠️  WARNING: This will DELETE ALL EXISTING DATA!")
	fmt.Print("Continue? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "y" && confirm != "Y" && confirm != "yes" && confirm != "Yes" {
		fmt.Println("❌ Aborted.")
		return
	}
	fmt.Println()

	// Reset database dengan foreign key check dimatikan
	ResetDatabase(db)

	// Run seeder
	fmt.Println("\n🚀 Running seeder...")
	fmt.Println("========================================")

	SeedUsers(db)
	SeedProducts(db)
	SeedStock(db)
	SeedTransactions(db)

	fmt.Println("========================================")
	fmt.Println("✅ Seeder completed successfully!")
	fmt.Println("📊 Summary:")
	fmt.Println("   - Users: 3 (superadmin, head, produksi)")
	fmt.Println("   - Products: 1000")
	fmt.Println("   - Stock: 1000")
	fmt.Println("   - Transaction In: 1000")
	fmt.Println("   - Transaction Out: 1000")
}