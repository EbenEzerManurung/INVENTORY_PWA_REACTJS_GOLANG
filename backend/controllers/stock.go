package controllers

import (
    "database/sql"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "inventory-backend/config"
    "inventory-backend/models"
    "inventory-backend/utils"

    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
)

// ==================== GET STOCK ====================
func GetStock(c *gin.Context) {
    search := strings.TrimSpace(c.Query("search"))
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    offset := (page - 1) * limit

    var whereClause string
    var args []interface{}
    
    if search != "" {
        whereClause = "WHERE p.product_name LIKE ? OR p.product_code LIKE ? OR p.category LIKE ?"
        searchPattern := "%" + search + "%"
        args = append(args, searchPattern, searchPattern, searchPattern)
    }

    var total int
    countQuery := fmt.Sprintf(`
        SELECT COUNT(*) FROM stock s 
        JOIN products p ON s.product_id = p.id 
        %s
    `, whereClause)
    err := config.DB.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get stock", err.Error())
        return
    }

    query := fmt.Sprintf(`
        SELECT s.id, s.product_id, s.quantity, s.last_updated,
               p.id, p.product_code, p.product_name, p.category, p.unit, p.min_stock, p.description, p.created_at
        FROM stock s 
        JOIN products p ON s.product_id = p.id 
        %s
        ORDER BY s.last_updated DESC 
        LIMIT ? OFFSET ?
    `, whereClause)

    args = append(args, limit, offset)
    rows, err := config.DB.Query(query, args...)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get stock", err.Error())
        return
    }
    defer rows.Close()

    var stocks []models.Stock
    for rows.Next() {
        var stock models.Stock
        var product models.Product
        err := rows.Scan(
            &stock.ID, &stock.ProductID, &stock.Quantity, &stock.LastUpdated,
            &product.ID, &product.ProductCode, &product.ProductName, 
            &product.Category, &product.Unit, &product.MinStock, &product.Description, &product.CreatedAt,
        )
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan stock", err.Error())
            return
        }
        stock.Product = product
        stocks = append(stocks, stock)
    }

    utils.PaginatedResponse(c, stocks, total, page, limit)
}

// ==================== GET STOCK BY PRODUCT ID ====================
func GetStockByProductID(c *gin.Context) {
    productID, err := strconv.Atoi(c.Param("productId"))
    if err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
        return
    }

    var stock models.Stock
    var product models.Product

    err = config.DB.QueryRow(`
        SELECT s.id, s.product_id, s.quantity, s.last_updated,
               p.id, p.product_code, p.product_name, p.category, p.unit, p.min_stock, p.description, p.created_at
        FROM stock s 
        JOIN products p ON s.product_id = p.id 
        WHERE s.product_id = ?
    `, productID).Scan(
        &stock.ID, &stock.ProductID, &stock.Quantity, &stock.LastUpdated,
        &product.ID, &product.ProductCode, &product.ProductName,
        &product.Category, &product.Unit, &product.MinStock, &product.Description, &product.CreatedAt,
    )

    if err == sql.ErrNoRows {
        utils.Error(c, http.StatusNotFound, "Stock not found", "")
        return
    }
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get stock", err.Error())
        return
    }

    stock.Product = product
    utils.Success(c, "Stock retrieved successfully", stock)
}

// ==================== ADJUST STOCK ====================
func AdjustStock(c *gin.Context) {
    productId, err := strconv.Atoi(c.Param("productId"))
    if err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid product ID", err.Error())
        return
    }

    var req struct {
        Quantity int    `json:"quantity" binding:"required"`
        Note     string `json:"note"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
        return
    }

    userID, _ := c.Get("userID")

    // Check if product exists
    var exists bool
    err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", productId).Scan(&exists)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to check product", err.Error())
        return
    }
    if !exists {
        utils.Error(c, http.StatusNotFound, "Product not found", "")
        return
    }

    tx, err := config.DB.Begin()
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
        return
    }
    defer tx.Rollback()

    // Update stock
    _, err = tx.Exec(`
        UPDATE stock 
        SET quantity = quantity + ? 
        WHERE product_id = ?
    `, req.Quantity, productId)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to update stock", err.Error())
        return
    }

    // Create transaction in for adjustment
    note := "Stock adjustment"
    if req.Note != "" {
        note = req.Note + " (adjustment)"
    }

    _, err = tx.Exec(`
        INSERT INTO transaction_in (product_id, quantity, note, created_by) 
        VALUES (?, ?, ?, ?)
    `, productId, req.Quantity, note, userID)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to create transaction record", err.Error())
        return
    }

    if err := tx.Commit(); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
        return
    }

    // Get updated stock
    var newQuantity int
    config.DB.QueryRow("SELECT quantity FROM stock WHERE product_id = ?", productId).Scan(&newQuantity)

    utils.Success(c, "Stock adjusted successfully", gin.H{
        "product_id":    productId,
        "adjustment":    req.Quantity,
        "new_quantity":  newQuantity,
        "note":          note,
    })
}

// ==================== EXPORT STOCK ====================
func ExportStock(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT p.product_code, p.product_name, p.category, p.unit, p.min_stock, 
               COALESCE(s.quantity, 0) as quantity
        FROM products p
        LEFT JOIN stock s ON p.id = s.product_id
        ORDER BY p.product_name
    `)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to export stock", err.Error())
        return
    }
    defer rows.Close()

    f := excelize.NewFile()
    sheet := "Stock"
    f.SetSheetName("Sheet1", sheet)

    headers := []string{"Product Code", "Product Name", "Category", "Unit", "Min Stock", "Current Stock", "Status"}
    for i, header := range headers {
        cell := fmt.Sprintf("%s1", string(rune('A'+i)))
        f.SetCellValue(sheet, cell, header)
    }

    row := 2
    for rows.Next() {
        var productCode, productName, category, unit string
        var minStock, quantity int

        err := rows.Scan(&productCode, &productName, &category, &unit, &minStock, &quantity)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan stock", err.Error())
            return
        }

        status := "Normal"
        if quantity == 0 {
            status = "Out of Stock"
        } else if quantity <= minStock {
            status = "Low Stock"
        }

        f.SetCellValue(sheet, fmt.Sprintf("A%d", row), productCode)
        f.SetCellValue(sheet, fmt.Sprintf("B%d", row), productName)
        f.SetCellValue(sheet, fmt.Sprintf("C%d", row), category)
        f.SetCellValue(sheet, fmt.Sprintf("D%d", row), unit)
        f.SetCellValue(sheet, fmt.Sprintf("E%d", row), minStock)
        f.SetCellValue(sheet, fmt.Sprintf("F%d", row), quantity)
        f.SetCellValue(sheet, fmt.Sprintf("G%d", row), status)
        row++
    }

    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Disposition", "attachment; filename=stock.xlsx")
    
    if err := f.Write(c.Writer); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to write excel file", err.Error())
        return
    }
}