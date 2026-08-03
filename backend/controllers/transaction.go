package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"inventory-backend/config"
	"inventory-backend/models"
	"inventory-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ==================== CREATE TRANSACTION IN ====================
func CreateTransactionIn(c *gin.Context) {
	var req models.TransactionInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, _ := c.Get("userID")

	tx, err := config.DB.Begin()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO transaction_in (product_id, quantity, note, created_by) 
		VALUES (?, ?, ?, ?)
	`, req.ProductID, req.Quantity, req.Note, userID)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create transaction", err.Error())
		return
	}

	transactionID, _ := result.LastInsertId()

	_, err = tx.Exec(`
		UPDATE stock 
		SET quantity = quantity + ? 
		WHERE product_id = ?
	`, req.Quantity, req.ProductID)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction in created successfully", gin.H{"id": transactionID})
}

// ==================== GET TRANSACTIONS IN ====================
func GetTransactionsIn(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var whereClause string
	var args []interface{}

	if search != "" {
		whereClause = "WHERE p.product_name LIKE ? OR p.product_code LIKE ? OR u.fullname LIKE ?"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM transaction_in t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		%s
	`, whereClause)
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}

	// ✅ Query dengan mengambil fullname user sebagai created_by_name
	query := fmt.Sprintf(`
		SELECT t.id, t.product_id, t.quantity, t.note, t.created_by, t.created_at,
			   p.id, p.product_code, p.product_name, p.category, p.unit,
			   u.id, u.username, u.fullname as created_by_name
		FROM transaction_in t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		%s
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}
	defer rows.Close()

	var transactions []models.TransactionIn
	for rows.Next() {
		var t models.TransactionIn
		var p models.Product
		var u models.User
		err := rows.Scan(
			&t.ID, &t.ProductID, &t.Quantity, &t.Note, &t.CreatedBy, &t.CreatedAt,
			&p.ID, &p.ProductCode, &p.ProductName, &p.Category, &p.Unit,
			&u.ID, &u.Username, &u.Fullname,
		)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to scan transactions", err.Error())
			return
		}
		t.Product = p
		t.CreatedByName = u.Fullname // ✅ Set nama user
		transactions = append(transactions, t)
	}

	utils.PaginatedResponse(c, transactions, total, page, limit)
}

// ==================== GET TRANSACTION IN BY ID ====================
func GetTransactionInByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var t models.TransactionIn
	var p models.Product
	var u models.User

	query := `
		SELECT t.id, t.product_id, t.quantity, t.note, t.created_by, t.created_at,
			   p.id, p.product_code, p.product_name, p.category, p.unit,
			   u.id, u.username, u.fullname
		FROM transaction_in t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		WHERE t.id = ?
	`

	err = config.DB.QueryRow(query, id).Scan(
		&t.ID, &t.ProductID, &t.Quantity, &t.Note, &t.CreatedBy, &t.CreatedAt,
		&p.ID, &p.ProductCode, &p.ProductName, &p.Category, &p.Unit,
		&u.ID, &u.Username, &u.Fullname,
	)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	t.Product = p
	t.CreatedByName = u.Fullname
	utils.Success(c, "Transaction retrieved successfully", t)
}

// ==================== UPDATE TRANSACTION IN ====================
func UpdateTransactionIn(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req struct {
		Quantity int    `json:"quantity"`
		Note     string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	// Get old quantity
	var oldQuantity int
	var productID int
	err = config.DB.QueryRow(
		"SELECT product_id, quantity FROM transaction_in WHERE id = ?",
		id,
	).Scan(&productID, &oldQuantity)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
		return
	}
	defer tx.Rollback()

	// Update transaction
	_, err = tx.Exec(`
		UPDATE transaction_in 
		SET quantity = ?, note = ?
		WHERE id = ?
	`, req.Quantity, req.Note, id)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update transaction", err.Error())
		return
	}

	// Update stock (adjust quantity difference)
	diff := req.Quantity - oldQuantity
	if diff != 0 {
		_, err = tx.Exec(`
			UPDATE stock 
			SET quantity = quantity + ? 
			WHERE product_id = ?
		`, diff, productID)

		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction updated successfully", nil)
}

// ==================== DELETE TRANSACTION IN ====================
func DeleteTransactionIn(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
		return
	}
	defer tx.Rollback()

	// Get transaction details
	var productID, quantity int
	err = tx.QueryRow(
		"SELECT product_id, quantity FROM transaction_in WHERE id = ?",
		id,
	).Scan(&productID, &quantity)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	// Delete transaction
	_, err = tx.Exec("DELETE FROM transaction_in WHERE id = ?", id)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to delete transaction", err.Error())
		return
	}

	// Update stock (reverse the addition)
	_, err = tx.Exec(`
		UPDATE stock 
		SET quantity = quantity - ? 
		WHERE product_id = ?
	`, quantity, productID)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction deleted successfully", nil)
}

// ==================== EXPORT TRANSACTIONS IN ====================
func ExportTransactionsIn(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))

	var whereClause string
	var args []interface{}

	if search != "" {
		whereClause = "WHERE p.product_name LIKE ? OR p.product_code LIKE ? OR u.fullname LIKE ?"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query := `
		SELECT t.id, p.product_code, p.product_name, t.quantity, t.note,
			   u.fullname as created_by, t.created_at
		FROM transaction_in t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
	` + whereClause + `
		ORDER BY t.created_at DESC
	`

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to export data", err.Error())
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Transaction In"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "Product Code", "Product Name", "Quantity", "Note", "Created By", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheet, cell, header)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", string(rune('A'+len(headers)-1))), headerStyle)

	row := 2
	for rows.Next() {
		var id, quantity int
		var productCode, productName, note, createdBy, createdAt string

		err := rows.Scan(&id, &productCode, &productName, &quantity, &note, &createdBy, &createdAt)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to scan data", err.Error())
			return
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), id)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), productCode)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), productName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), quantity)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), note)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), createdBy)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), createdAt)
		row++
	}

	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheet, col, col, 20)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=transaction_in_"+time.Now().Format("2006-01-02_15-04")+".xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	c.Header("Pragma", "no-cache")

	if err := f.Write(c.Writer); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to write excel file", err.Error())
		return
	}
}

// ==================== CREATE TRANSACTION OUT ====================
func CreateTransactionOut(c *gin.Context) {
	var req models.TransactionOutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var currentStock int
	err := config.DB.QueryRow(
		"SELECT quantity FROM stock WHERE product_id = ?",
		req.ProductID,
	).Scan(&currentStock)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to check stock", err.Error())
		return
	}

	if currentStock < req.Quantity {
		utils.Error(c, http.StatusBadRequest, "Insufficient stock", fmt.Sprintf("Available: %d", currentStock))
		return
	}

	status := "pending"
	if role == "head" || role == "superadmin" {
		status = "approved"
	}

	result, err := config.DB.Exec(`
		INSERT INTO transaction_out (product_id, quantity, note, created_by, status) 
		VALUES (?, ?, ?, ?, ?)
	`, req.ProductID, req.Quantity, req.Note, userID, status)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create transaction", err.Error())
		return
	}

	transactionID, _ := result.LastInsertId()

	if status == "approved" {
		_, err = config.DB.Exec(`
			UPDATE stock 
			SET quantity = quantity - ? 
			WHERE product_id = ? AND quantity >= ?
		`, req.Quantity, req.ProductID, req.Quantity)

		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
			return
		}
	}

	message := "Transaction out created successfully, waiting for approval"
	if status == "approved" {
		message = "Transaction out approved successfully"
	}

	utils.Success(c, message, gin.H{
		"id":     transactionID,
		"status": status,
	})
}

// ==================== GET TRANSACTIONS OUT ====================
func GetTransactionsOut(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var whereClause string
	var args []interface{}

	conditions := []string{}
	if search != "" {
		conditions = append(conditions, "(p.product_name LIKE ? OR p.product_code LIKE ? OR u.fullname LIKE ?)")
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	if status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, status)
	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		%s
	`, whereClause)
	err := config.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.product_id, t.quantity, t.note, t.status,
			   t.created_by, t.approved_by, t.created_at, t.approved_at,
			   p.id, p.product_code, p.product_name, p.category, p.unit,
			   u.id, u.username, u.fullname as created_by_name
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		%s
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)
	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transactions", err.Error())
		return
	}
	defer rows.Close()

	var transactions []models.TransactionOut
	for rows.Next() {
		var t models.TransactionOut
		var p models.Product
		var u models.User
		var approvedBy sql.NullInt64
		var approvedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.ProductID, &t.Quantity, &t.Note, &t.Status,
			&t.CreatedBy, &approvedBy, &t.CreatedAt, &approvedAt,
			&p.ID, &p.ProductCode, &p.ProductName, &p.Category, &p.Unit,
			&u.ID, &u.Username, &u.Fullname,
		)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to scan transactions", err.Error())
			return
		}

		t.Product = p
		t.CreatedByName = u.Fullname

		if approvedBy.Valid {
			t.ApprovedBy = new(int)
			*t.ApprovedBy = int(approvedBy.Int64)
		}
		if approvedAt.Valid {
			t.ApprovedAt = &approvedAt.Time
		}

		transactions = append(transactions, t)
	}

	utils.PaginatedResponse(c, transactions, total, page, limit)
}

// ==================== GET TRANSACTION OUT BY ID ====================
func GetTransactionOutByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var t models.TransactionOut
	var p models.Product
	var u models.User
	var approvedBy sql.NullInt64
	var approvedAt sql.NullTime

	query := `
		SELECT t.id, t.product_id, t.quantity, t.note, t.status,
			   t.created_by, t.approved_by, t.created_at, t.approved_at,
			   p.id, p.product_code, p.product_name, p.category, p.unit,
			   u.id, u.username, u.fullname
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		WHERE t.id = ?
	`

	err = config.DB.QueryRow(query, id).Scan(
		&t.ID, &t.ProductID, &t.Quantity, &t.Note, &t.Status,
		&t.CreatedBy, &approvedBy, &t.CreatedAt, &approvedAt,
		&p.ID, &p.ProductCode, &p.ProductName, &p.Category, &p.Unit,
		&u.ID, &u.Username, &u.Fullname,
	)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	t.Product = p
	t.CreatedByName = u.Fullname

	if approvedBy.Valid {
		t.ApprovedBy = new(int)
		*t.ApprovedBy = int(approvedBy.Int64)
	}
	if approvedAt.Valid {
		t.ApprovedAt = &approvedAt.Time
	}

	utils.Success(c, "Transaction retrieved successfully", t)
}

// ==================== EXPORT TRANSACTIONS OUT ====================
func ExportTransactionsOut(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))

	var whereClause string
	var args []interface{}

	conditions := []string{}
	if search != "" {
		conditions = append(conditions, "(p.product_name LIKE ? OR p.product_code LIKE ? OR u.fullname LIKE ?)")
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}
	if status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, status)
	}

	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT t.id, p.product_code, p.product_name, t.quantity, t.note, t.status,
			   u.fullname as requested_by, t.created_at,
			   COALESCE(a.fullname, '') as approved_by
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		LEFT JOIN users a ON t.approved_by = a.id
	` + whereClause + `
		ORDER BY t.created_at DESC
	`

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to export data", err.Error())
		return
	}
	defer rows.Close()

	f := excelize.NewFile()
	sheet := "Transaction Out"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"ID", "Product Code", "Product Name", "Quantity", "Note", "Status", "Requested By", "Approved By", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheet, cell, header)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
			Color: "#FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", string(rune('A'+len(headers)-1))), headerStyle)

	row := 2
	for rows.Next() {
		var id, quantity int
		var productCode, productName, note, status, requestedBy, approvedBy, createdAt string

		err := rows.Scan(&id, &productCode, &productName, &quantity, &note, &status,
			&requestedBy, &createdAt, &approvedBy)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to scan data", err.Error())
			return
		}

		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), id)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), productCode)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), productName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), quantity)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), note)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), status)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), requestedBy)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), approvedBy)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), createdAt)
		row++
	}

	for i := 0; i < len(headers); i++ {
		col := string(rune('A' + i))
		f.SetColWidth(sheet, col, col, 20)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=transaction_out_"+time.Now().Format("2006-01-02_15-04")+".xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")
	c.Header("Pragma", "no-cache")

	if err := f.Write(c.Writer); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to write excel file", err.Error())
		return
	}
}

// ==================== APPROVE TRANSACTION OUT ====================
func ApproveTransactionOut(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	var req models.ApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	userID, _ := c.Get("userID")

	var transaction struct {
		ID        int
		ProductID int
		Quantity  int
		Status    string
	}

	err = config.DB.QueryRow(`
		SELECT id, product_id, quantity, status 
		FROM transaction_out 
		WHERE id = ? AND status = 'pending'
	`, id).Scan(&transaction.ID, &transaction.ProductID, &transaction.Quantity, &transaction.Status)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found or already processed", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	tx, err := config.DB.Begin()
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
		return
	}
	defer tx.Rollback()

	now := time.Now()

	if req.Status == "approved" {
		var stock int
		err = tx.QueryRow("SELECT quantity FROM stock WHERE product_id = ?", transaction.ProductID).Scan(&stock)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to check stock", err.Error())
			return
		}

		if stock < transaction.Quantity {
			utils.Error(c, http.StatusBadRequest, "Insufficient stock", fmt.Sprintf("Available: %d, Requested: %d", stock, transaction.Quantity))
			return
		}

		_, err = tx.Exec(`
			UPDATE stock 
			SET quantity = quantity - ? 
			WHERE product_id = ? AND quantity >= ?
		`, transaction.Quantity, transaction.ProductID, transaction.Quantity)

		if err != nil {
			utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
			return
		}
	}

	_, err = tx.Exec(`
		UPDATE transaction_out 
		SET status = ?, approved_by = ?, approved_at = ?
		WHERE id = ?
	`, req.Status, userID, now, id)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update transaction", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction "+req.Status+" successfully", gin.H{"id": id, "status": req.Status})
}

// ==================== REJECT TRANSACTION OUT ====================
func RejectTransactionOut(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	userID, _ := c.Get("userID")

	var status string
	err = config.DB.QueryRow(
		"SELECT status FROM transaction_out WHERE id = ?",
		id,
	).Scan(&status)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	if status != "pending" {
		utils.Error(c, http.StatusBadRequest, "Transaction is not pending", "")
		return
	}

	now := time.Now()
	_, err = config.DB.Exec(`
		UPDATE transaction_out 
		SET status = 'rejected', approved_by = ?, approved_at = ?
		WHERE id = ?
	`, userID, now, id)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to reject transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction rejected successfully", nil)
}

// ==================== CANCEL TRANSACTION OUT ====================
func CancelTransactionOut(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
		return
	}

	userID, _ := c.Get("userID")

	var status string
	err = config.DB.QueryRow(
		"SELECT status FROM transaction_out WHERE id = ? AND created_by = ?",
		id, userID,
	).Scan(&status)

	if err == sql.ErrNoRows {
		utils.Error(c, http.StatusNotFound, "Transaction not found or not owned by user", "")
		return
	}
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get transaction", err.Error())
		return
	}

	if status != "pending" {
		utils.Error(c, http.StatusBadRequest, "Only pending transactions can be cancelled", "")
		return
	}

	_, err = config.DB.Exec(`
		UPDATE transaction_out 
		SET status = 'cancelled', approved_by = ?
		WHERE id = ?
	`, userID, id)

	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to cancel transaction", err.Error())
		return
	}

	utils.Success(c, "Transaction cancelled successfully", nil)
}

// ==================== GET TRANSACTION OUT STATISTICS ====================
func GetTransactionOutStatistics(c *gin.Context) {
	var total, approved, pending, rejected, cancelled int

	err := config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out").Scan(&total)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get statistics", err.Error())
		return
	}

	err = config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'approved'").Scan(&approved)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get statistics", err.Error())
		return
	}

	err = config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'pending'").Scan(&pending)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get statistics", err.Error())
		return
	}

	err = config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'rejected'").Scan(&rejected)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get statistics", err.Error())
		return
	}

	err = config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'cancelled'").Scan(&cancelled)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to get statistics", err.Error())
		return
	}

	utils.Success(c, "Transaction statistics", gin.H{
		"total":     total,
		"approved":  approved,
		"pending":   pending,
		"rejected":  rejected,
		"cancelled": cancelled,
	})
}