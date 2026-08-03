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

// ==================== GET PRODUCTS ====================
func GetProducts(c *gin.Context) {
    search := strings.TrimSpace(c.Query("search"))
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    offset := (page - 1) * limit

    var whereClause string
    var args []interface{}
    
    if search != "" {
        whereClause = "WHERE product_name LIKE ? OR product_code LIKE ? OR category LIKE ?"
        searchPattern := "%" + search + "%"
        args = append(args, searchPattern, searchPattern, searchPattern)
    }

    var total int
    countQuery := "SELECT COUNT(*) FROM products " + whereClause
    err := config.DB.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get products", err.Error())
        return
    }

    query := fmt.Sprintf(`
        SELECT id, product_code, product_name, category, unit, min_stock, description, 
               IFNULL(image, '') as image, 
               created_at 
        FROM products %s 
        ORDER BY id DESC 
        LIMIT ? OFFSET ?
    `, whereClause)

    args = append(args, limit, offset)
    rows, err := config.DB.Query(query, args...)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get products", err.Error())
        return
    }
    defer rows.Close()

    var products []models.Product
    for rows.Next() {
        var product models.Product
        err := rows.Scan(
            &product.ID,
            &product.ProductCode,
            &product.ProductName,
            &product.Category,
            &product.Unit,
            &product.MinStock,
            &product.Description,
            &product.Image,
            &product.CreatedAt,
        )
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan products", err.Error())
            return
        }
        
        products = append(products, product)
    }

    utils.PaginatedResponse(c, products, total, page, limit)
}

// ==================== GET PRODUCT BY ID ====================
func GetProductByID(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }

    var product models.Product
    err = config.DB.QueryRow(`
        SELECT id, product_code, product_name, category, unit, min_stock, description, 
               COALESCE(image, '') as image, 
               created_at 
        FROM products WHERE id = ?
    `, id).Scan(
        &product.ID,
        &product.ProductCode,
        &product.ProductName,
        &product.Category,
        &product.Unit,
        &product.MinStock,
        &product.Description,
        &product.Image,
        &product.CreatedAt,
    )

    if err == sql.ErrNoRows {
        utils.Error(c, http.StatusNotFound, "Product not found", "")
        return
    }
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get product", err.Error())
        return
    }

    utils.Success(c, "Product retrieved successfully", product)
}

// ==================== CREATE PRODUCT ====================
func CreateProduct(c *gin.Context) {
    var req models.ProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
        return
    }

    tx, err := config.DB.Begin()
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to start transaction", err.Error())
        return
    }
    defer tx.Rollback()

    result, err := tx.Exec(`
        INSERT INTO products (product_code, product_name, category, unit, min_stock, description, image) 
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, req.ProductCode, req.ProductName, req.Category, req.Unit, req.MinStock, req.Description, req.Image)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to create product", err.Error())
        return
    }

    productID, _ := result.LastInsertId()

    _, err = tx.Exec("INSERT INTO stock (product_id, quantity) VALUES (?, 0)", productID)
    if err != nil {
        tx.Rollback()
        utils.Error(c, http.StatusInternalServerError, "Failed to initialize stock", err.Error())
        return
    }

    if err := tx.Commit(); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
        return
    }

    utils.Success(c, "Product created successfully", gin.H{"id": productID})
}

// ==================== UPDATE PRODUCT ====================
func UpdateProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }

    var req models.ProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
        return
    }

    _, err = config.DB.Exec(`
        UPDATE products 
        SET product_code = ?, product_name = ?, category = ?, unit = ?, min_stock = ?, description = ?, image = ?
        WHERE id = ?
    `, req.ProductCode, req.ProductName, req.Category, req.Unit, req.MinStock, req.Description, req.Image, id)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to update product", err.Error())
        return
    }

    utils.Success(c, "Product updated successfully", nil)
}

// ==================== DELETE PRODUCT ====================
func DeleteProduct(c *gin.Context) {
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

    _, err = tx.Exec("DELETE FROM stock WHERE product_id = ?", id)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to delete stock", err.Error())
        return
    }

    _, err = tx.Exec("DELETE FROM products WHERE id = ?", id)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to delete product", err.Error())
        return
    }

    if err := tx.Commit(); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to commit transaction", err.Error())
        return
    }

    utils.Success(c, "Product deleted successfully", nil)
}

// ==================== EXPORT PRODUCTS ====================
func ExportProducts(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT id, product_code, product_name, category, unit, min_stock, description, 
               COALESCE(image, '') as image, 
               created_at 
        FROM products ORDER BY id DESC
    `)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to export products", err.Error())
        return
    }
    defer rows.Close()

    f := excelize.NewFile()
    sheet := "Products"
    f.SetSheetName("Sheet1", sheet)

    headers := []string{"ID", "Product Code", "Product Name", "Category", "Unit", "Min Stock", "Description", "Image", "Created At"}
    for i, header := range headers {
        cell := fmt.Sprintf("%s1", string(rune('A'+i)))
        f.SetCellValue(sheet, cell, header)
    }

    row := 2
    for rows.Next() {
        var product models.Product
        err := rows.Scan(
            &product.ID,
            &product.ProductCode,
            &product.ProductName,
            &product.Category,
            &product.Unit,
            &product.MinStock,
            &product.Description,
            &product.Image,
            &product.CreatedAt,
        )
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan products", err.Error())
            return
        }

        f.SetCellValue(sheet, fmt.Sprintf("A%d", row), product.ID)
        f.SetCellValue(sheet, fmt.Sprintf("B%d", row), product.ProductCode)
        f.SetCellValue(sheet, fmt.Sprintf("C%d", row), product.ProductName)
        f.SetCellValue(sheet, fmt.Sprintf("D%d", row), product.Category)
        f.SetCellValue(sheet, fmt.Sprintf("E%d", row), product.Unit)
        f.SetCellValue(sheet, fmt.Sprintf("F%d", row), product.MinStock)
        f.SetCellValue(sheet, fmt.Sprintf("G%d", row), product.Description)
        f.SetCellValue(sheet, fmt.Sprintf("H%d", row), product.Image)
        f.SetCellValue(sheet, fmt.Sprintf("I%d", row), product.CreatedAt.Format("2006-01-02 15:04:05"))
        row++
    }

    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Disposition", "attachment; filename=products.xlsx")
    
    if err := f.Write(c.Writer); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to write excel file", err.Error())
        return
    }
}

// ==================== GET PRODUCT STOCK HISTORY ====================
func GetProductStockHistory(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        utils.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }

    // Check if product exists
    var exists bool
    err = config.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE id = ?)", id).Scan(&exists)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to check product", err.Error())
        return
    }
    if !exists {
        utils.Error(c, http.StatusNotFound, "Product not found", "")
        return
    }

    // Get current stock
    var currentStock int
    err = config.DB.QueryRow("SELECT COALESCE(quantity, 0) FROM stock WHERE product_id = ?", id).Scan(&currentStock)
    if err != nil {
        currentStock = 0
    }

    // Get transaction history (combine in and out)
    rows, err := config.DB.Query(`
        SELECT 
            'IN' as type,
            quantity,
            note,
            created_by,
            created_at
        FROM transaction_in 
        WHERE product_id = ?
        UNION ALL
        SELECT 
            'OUT' as type,
            quantity,
            note,
            created_by,
            created_at
        FROM transaction_out 
        WHERE product_id = ? AND status = 'approved'
        ORDER BY created_at DESC
        LIMIT 50
    `, id, id)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get transaction history", err.Error())
        return
    }
    defer rows.Close()

    var history []gin.H
    for rows.Next() {
        var transactionType string
        var quantity int
        var note string
        var createdBy int
        var createdAt string

        err := rows.Scan(&transactionType, &quantity, &note, &createdBy, &createdAt)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan history", err.Error())
            return
        }

        // Get user name
        var userName string
        config.DB.QueryRow("SELECT fullname FROM users WHERE id = ?", createdBy).Scan(&userName)

        history = append(history, gin.H{
            "type":          transactionType,
            "quantity":      quantity,
            "note":          note,
            "created_by":    createdBy,
            "created_by_name": userName,
            "created_at":    createdAt,
        })
    }

    // Get product details
    var product models.Product
    err = config.DB.QueryRow(`
        SELECT id, product_code, product_name, category, unit, min_stock, description, 
               COALESCE(image, '') as image, 
               created_at 
        FROM products WHERE id = ?
    `, id).Scan(
        &product.ID,
        &product.ProductCode,
        &product.ProductName,
        &product.Category,
        &product.Unit,
        &product.MinStock,
        &product.Description,
        &product.Image,
        &product.CreatedAt,
    )
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get product", err.Error())
        return
    }

    utils.Success(c, "Product stock history retrieved", gin.H{
        "product": product,
        "current_stock": currentStock,
        "history": history,
        "total_transactions": len(history),
    })
}