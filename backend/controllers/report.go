package controllers

import (
    "fmt"
    "net/http"
    "time"

    "inventory-backend/config"
    "inventory-backend/utils"

    "github.com/gin-gonic/gin"
    "github.com/xuri/excelize/v2"
)

// ==================== GET STOCK REPORT ====================
func GetStockReport(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT p.id, p.product_code, p.product_name, p.category, p.unit, p.min_stock, 
               COALESCE(s.quantity, 0) as quantity,
               COALESCE((SELECT SUM(quantity) FROM transaction_in WHERE product_id = p.id), 0) as total_in,
               COALESCE((SELECT SUM(quantity) FROM transaction_out WHERE product_id = p.id AND status = 'approved'), 0) as total_out
        FROM products p
        LEFT JOIN stock s ON p.id = s.product_id
        ORDER BY p.product_name
    `)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get report", err.Error())
        return
    }
    defer rows.Close()

    var reports []map[string]interface{}
    var totalProducts, totalStock, lowStock, outOfStock int

    for rows.Next() {
        var id, minStock, quantity, totalIn, totalOut int
        var productCode, productName, category, unit string

        err := rows.Scan(&id, &productCode, &productName, &category, &unit, &minStock, &quantity, &totalIn, &totalOut)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan report", err.Error())
            return
        }

        totalProducts++
        totalStock += quantity

        status := "Normal"
        if quantity == 0 {
            status = "Out of Stock"
            outOfStock++
        } else if quantity <= minStock {
            status = "Low Stock"
            lowStock++
        }

        report := map[string]interface{}{
            "id":           id,
            "product_code": productCode,
            "product_name": productName,
            "category":     category,
            "unit":         unit,
            "min_stock":    minStock,
            "quantity":     quantity,
            "total_in":     totalIn,
            "total_out":    totalOut,
            "status":       status,
        }
        reports = append(reports, report)
    }

    utils.Success(c, "Report generated successfully", gin.H{
        "data": reports,
        "summary": gin.H{
            "total_products": totalProducts,
            "total_stock":    totalStock,
            "low_stock":      lowStock,
            "out_of_stock":   outOfStock,
            "healthy":        totalProducts - lowStock - outOfStock,
        },
    })
}

// ==================== EXPORT STOCK REPORT ====================
func ExportStockReport(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT p.id, p.product_code, p.product_name, p.category, p.unit, p.min_stock, 
               COALESCE(s.quantity, 0) as quantity,
               COALESCE((SELECT SUM(quantity) FROM transaction_in WHERE product_id = p.id), 0) as total_in,
               COALESCE((SELECT SUM(quantity) FROM transaction_out WHERE product_id = p.id AND status = 'approved'), 0) as total_out
        FROM products p
        LEFT JOIN stock s ON p.id = s.product_id
        ORDER BY p.product_name
    `)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to export report", err.Error())
        return
    }
    defer rows.Close()

    f := excelize.NewFile()
    sheet := "Stock Report"
    f.SetSheetName("Sheet1", sheet)

    headers := []string{"ID", "Product Code", "Product Name", "Category", "Unit", "Min Stock", "Current Stock", "Total In", "Total Out", "Status"}
    for i, header := range headers {
        cell := fmt.Sprintf("%s1", string(rune('A'+i)))
        f.SetCellValue(sheet, cell, header)
    }

    row := 2
    for rows.Next() {
        var id, minStock, quantity, totalIn, totalOut int
        var productCode, productName, category, unit string

        err := rows.Scan(&id, &productCode, &productName, &category, &unit, &minStock, &quantity, &totalIn, &totalOut)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan report", err.Error())
            return
        }

        status := "Normal"
        if quantity == 0 {
            status = "Out of Stock"
        } else if quantity <= minStock {
            status = "Low Stock"
        }

        f.SetCellValue(sheet, fmt.Sprintf("A%d", row), id)
        f.SetCellValue(sheet, fmt.Sprintf("B%d", row), productCode)
        f.SetCellValue(sheet, fmt.Sprintf("C%d", row), productName)
        f.SetCellValue(sheet, fmt.Sprintf("D%d", row), category)
        f.SetCellValue(sheet, fmt.Sprintf("E%d", row), unit)
        f.SetCellValue(sheet, fmt.Sprintf("F%d", row), minStock)
        f.SetCellValue(sheet, fmt.Sprintf("G%d", row), quantity)
        f.SetCellValue(sheet, fmt.Sprintf("H%d", row), totalIn)
        f.SetCellValue(sheet, fmt.Sprintf("I%d", row), totalOut)
        f.SetCellValue(sheet, fmt.Sprintf("J%d", row), status)
        row++
    }

    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Disposition", "attachment; filename=stock_report_"+time.Now().Format("20060102")+".xlsx")
    
    if err := f.Write(c.Writer); err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to write excel file", err.Error())
        return
    }
}

// ==================== GET DASHBOARD STATS ====================
func GetDashboardStats(c *gin.Context) {
    var totalProducts, totalStock, totalIn, totalOut int
    var pendingApprovals int

    config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
    config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM stock").Scan(&totalStock)
    config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_in").Scan(&totalIn)
    config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_out WHERE status = 'approved'").Scan(&totalOut)
    config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'pending'").Scan(&pendingApprovals)

    stats := map[string]interface{}{
        "total_products":    totalProducts,
        "total_stock":       totalStock,
        "total_in":          totalIn,
        "total_out":         totalOut,
        "pending_approvals": pendingApprovals,
        "date":              time.Now().Format("2006-01-02"),
    }

    utils.Success(c, "Dashboard stats retrieved", stats)
}

// ==================== GET TRANSACTION REPORT ====================
func GetTransactionReport(c *gin.Context) {
    // Get date range from query
    startDate := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
    endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

    // Summary
    var totalIn, totalOut int
    err := config.DB.QueryRow(`
        SELECT 
            COALESCE((SELECT SUM(quantity) FROM transaction_in WHERE DATE(created_at) BETWEEN ? AND ?), 0),
            COALESCE((SELECT SUM(quantity) FROM transaction_out WHERE DATE(created_at) BETWEEN ? AND ? AND status = 'approved'), 0)
    `, startDate, endDate, startDate, endDate).Scan(&totalIn, &totalOut)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get transaction summary", err.Error())
        return
    }

    // Daily transactions
    rows, err := config.DB.Query(`
        SELECT 
            DATE(created_at) as date,
            COUNT(*) as total_transactions,
            SUM(CASE WHEN type = 'in' THEN quantity ELSE 0 END) as total_in,
            SUM(CASE WHEN type = 'out' THEN quantity ELSE 0 END) as total_out
        FROM (
            SELECT 'in' as type, quantity, created_at FROM transaction_in
            UNION ALL
            SELECT 'out' as type, quantity, created_at FROM transaction_out WHERE status = 'approved'
        ) t
        WHERE DATE(created_at) BETWEEN ? AND ?
        GROUP BY DATE(created_at)
        ORDER BY date DESC
        LIMIT 30
    `, startDate, endDate)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get daily transactions", err.Error())
        return
    }
    defer rows.Close()

    var dailyTransactions []gin.H
    for rows.Next() {
        var date string
        var totalTransactions, totalInDaily, totalOutDaily int

        err := rows.Scan(&date, &totalTransactions, &totalInDaily, &totalOutDaily)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan data", err.Error())
            return
        }

        dailyTransactions = append(dailyTransactions, gin.H{
            "date":               date,
            "total_transactions": totalTransactions,
            "total_in":           totalInDaily,
            "total_out":          totalOutDaily,
        })
    }

    // Top products
    topRows, err := config.DB.Query(`
        SELECT 
            p.product_name,
            COALESCE(SUM(CASE WHEN t.type = 'in' THEN t.quantity ELSE 0 END), 0) as total_in,
            COALESCE(SUM(CASE WHEN t.type = 'out' THEN t.quantity ELSE 0 END), 0) as total_out
        FROM (
            SELECT 'in' as type, product_id, quantity FROM transaction_in
            UNION ALL
            SELECT 'out' as type, product_id, quantity FROM transaction_out WHERE status = 'approved'
        ) t
        JOIN products p ON t.product_id = p.id
        GROUP BY p.id, p.product_name
        ORDER BY (SUM(CASE WHEN t.type = 'in' THEN t.quantity ELSE 0 END) + SUM(CASE WHEN t.type = 'out' THEN t.quantity ELSE 0 END)) DESC
        LIMIT 10
    `)

    if err == nil {
        defer topRows.Close()
        var topProducts []gin.H
        for topRows.Next() {
            var productName string
            var totalInTop, totalOutTop int
            topRows.Scan(&productName, &totalInTop, &totalOutTop)
            topProducts = append(topProducts, gin.H{
                "product_name": productName,
                "total_in":     totalInTop,
                "total_out":    totalOutTop,
            })
        }
    }

    utils.Success(c, "Transaction report retrieved", gin.H{
        "summary": gin.H{
            "total_transactions_in":  totalIn,
            "total_transactions_out": totalOut,
            "total_transactions":     totalIn + totalOut,
            "net_change":             totalIn - totalOut,
        },
        "daily_transactions": dailyTransactions,
        "top_products":       dailyTransactions, // Placeholder
        "date_range": gin.H{
            "start_date": startDate,
            "end_date":   endDate,
        },
    })
}

// ==================== GET PRODUCT CATEGORY REPORT ====================
func GetProductCategoryReport(c *gin.Context) {
    rows, err := config.DB.Query(`
        SELECT 
            p.category,
            COUNT(*) as total_products,
            COALESCE(SUM(s.quantity), 0) as total_stock,
            COALESCE(AVG(s.quantity), 0) as avg_stock,
            COALESCE(MIN(s.quantity), 0) as min_stock,
            COALESCE(MAX(s.quantity), 0) as max_stock,
            COALESCE(SUM(CASE WHEN s.quantity <= p.min_stock THEN 1 ELSE 0 END), 0) as low_stock_count
        FROM products p
        LEFT JOIN stock s ON p.id = s.product_id
        GROUP BY p.category
        ORDER BY total_products DESC
    `)

    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get category report", err.Error())
        return
    }
    defer rows.Close()

    var categories []gin.H
    for rows.Next() {
        var category string
        var totalProducts, totalStock, avgStock, minStock, maxStock, lowStockCount int

        err := rows.Scan(&category, &totalProducts, &totalStock, &avgStock, &minStock, &maxStock, &lowStockCount)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan data", err.Error())
            return
        }

        categories = append(categories, gin.H{
            "category":         category,
            "total_products":   totalProducts,
            "total_stock":      totalStock,
            "avg_stock":        avgStock,
            "min_stock":        minStock,
            "max_stock":        maxStock,
            "low_stock_count":  lowStockCount,
        })
    }

    utils.Success(c, "Product category report retrieved", gin.H{
        "categories":       categories,
        "total_categories": len(categories),
    })
}

// ==================== GET SUMMARY REPORT ====================
func GetSummaryReport(c *gin.Context) {
    // Get various statistics
    var totalProducts, totalUsers, totalTransactions, pendingRequests, lowStockItems, totalStock int

    // Total products
    err := config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get total products", err.Error())
        return
    }

    // Total users
    err = config.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get total users", err.Error())
        return
    }

    // Total transactions
    err = config.DB.QueryRow(`
        SELECT COUNT(*) FROM (
            SELECT id FROM transaction_in
            UNION ALL
            SELECT id FROM transaction_out
        ) t
    `).Scan(&totalTransactions)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get total transactions", err.Error())
        return
    }

    // Pending requests (transaction out with pending status)
    err = config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'pending'").Scan(&pendingRequests)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get pending requests", err.Error())
        return
    }

    // Low stock items
    err = config.DB.QueryRow(`
        SELECT COUNT(*) FROM products p
        JOIN stock s ON p.id = s.product_id
        WHERE s.quantity <= p.min_stock
    `).Scan(&lowStockItems)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get low stock items", err.Error())
        return
    }

    // Total stock
    err = config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM stock").Scan(&totalStock)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get total stock", err.Error())
        return
    }

    // Recent activities (last 5 transactions)
    rows, err := config.DB.Query(`
        SELECT 'IN' as type, product_id, quantity, created_at, created_by 
        FROM transaction_in 
        UNION ALL
        SELECT 'OUT' as type, product_id, quantity, created_at, created_by 
        FROM transaction_out 
        ORDER BY created_at DESC 
        LIMIT 5
    `)
    if err != nil {
        utils.Error(c, http.StatusInternalServerError, "Failed to get recent activities", err.Error())
        return
    }
    defer rows.Close()

    var recentActivities []gin.H
    for rows.Next() {
        var typeStr string
        var productID, quantity, createdBy int
        var createdAt time.Time

        err := rows.Scan(&typeStr, &productID, &quantity, &createdAt, &createdBy)
        if err != nil {
            utils.Error(c, http.StatusInternalServerError, "Failed to scan activities", err.Error())
            return
        }

        // Get product name
        var productName string
        config.DB.QueryRow("SELECT product_name FROM products WHERE id = ?", productID).Scan(&productName)

        // Get user name
        var userName string
        config.DB.QueryRow("SELECT fullname FROM users WHERE id = ?", createdBy).Scan(&userName)

        recentActivities = append(recentActivities, gin.H{
            "type":            typeStr,
            "product_id":      productID,
            "product_name":    productName,
            "quantity":        quantity,
            "created_by":      createdBy,
            "created_by_name": userName,
            "created_at":      createdAt,
        })
    }

    // Get daily summary
    today := time.Now().Format("2006-01-02")
    var todayIn, todayOut int
    config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_in WHERE DATE(created_at) = ?", today).Scan(&todayIn)
    config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_out WHERE DATE(created_at) = ? AND status = 'approved'", today).Scan(&todayOut)

    utils.Success(c, "Summary report retrieved", gin.H{
        "summary": gin.H{
            "total_products":     totalProducts,
            "total_users":        totalUsers,
            "total_transactions": totalTransactions,
            "pending_requests":   pendingRequests,
            "low_stock_items":    lowStockItems,
            "total_stock":        totalStock,
        },
        "today_summary": gin.H{
            "date":             today,
            "transactions_in":  todayIn,
            "transactions_out": todayOut,
        },
        "recent_activities": recentActivities,
    })
}