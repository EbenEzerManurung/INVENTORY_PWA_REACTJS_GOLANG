package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"inventory-backend/config"
	"inventory-backend/utils"

	"github.com/gin-gonic/gin"
)

type AIChatRequest struct {
	Message string `json:"message" binding:"required"`
	UserID  int    `json:"user_id"`
}

type AIChatResponse struct {
	Response string      `json:"response"`
	Data     interface{} `json:"data,omitempty"`
}

func AIChat(c *gin.Context) {
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	fmt.Printf("🤖 AI Chat Request: %s\n", req.Message)

	userMessage := req.Message

	dbData, err := queryDatabase(userMessage)
	if err != nil {
		fmt.Printf("❌ Database query error: %v\n", err)
		utils.Error(c, http.StatusInternalServerError, "Failed to query database", err.Error())
		return
	}

	response := generateSmartResponse(userMessage, dbData)

	result := AIChatResponse{
		Response: response,
		Data:     dbData,
	}

	fmt.Printf("✅ AI Response generated (length: %d)\n", len(response))
	utils.Success(c, "AI response generated", result)
}

// ==================== SMART RESPONSE GENERATOR ====================
func generateSmartResponse(query string, dbData map[string]interface{}) string {
	q := strings.ToLower(query)
	var response strings.Builder

	// ==================== GREETING ====================
	if strings.TrimSpace(q) == "" ||
		strings.Contains(q, "hello") || strings.Contains(q, "halo") ||
		strings.Contains(q, "hi") || strings.Contains(q, "hai") ||
		strings.Contains(q, "hey") || strings.Contains(q, "good morning") ||
		strings.Contains(q, "good afternoon") || strings.Contains(q, "good evening") ||
		strings.Contains(q, "good night") || strings.Contains(q, "selamat") ||
		strings.Contains(q, "hallo") || strings.Contains(q, "helo") {

		hour := time.Now().Hour()
		var timeGreeting, emoji string
		if hour >= 5 && hour < 12 {
			timeGreeting, emoji = "Good Morning", "🌅"
		} else if hour >= 12 && hour < 17 {
			timeGreeting, emoji = "Good Afternoon", "☀️"
		} else if hour >= 17 && hour < 20 {
			timeGreeting, emoji = "Good Evening", "🌤️"
		} else {
			timeGreeting, emoji = "Good Evening", "🌙"
		}

		response.WriteString(fmt.Sprintf("%s **%s!** 👋\n\n", emoji, timeGreeting))
		response.WriteString("Welcome to **AI Inventory Assistant**.\n\n")
		response.WriteString("Ask me about:\n")
		response.WriteString("  • 📊 Stock summary\n")
		response.WriteString("  • 🏆 Top products\n")
		response.WriteString("  • ⏳ Pending approvals\n")
		response.WriteString("  • 👤 Users & roles\n")
		response.WriteString("  • ⚠️ Low stock items\n\n")
		response.WriteString("_What would you like to know?_ 🚀")

		return response.String()
	}

	// ==================== TOTAL PRODUCTS ====================
	if strings.Contains(q, "total product") || strings.Contains(q, "jumlah produk") ||
		strings.Contains(q, "berapa produk") || strings.Contains(q, "product count") ||
		strings.Contains(q, "total produk") {

		var totalProducts int
		err := config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
		if err == nil {
			response.WriteString(fmt.Sprintf("📊 **Total Products: %d**", totalProducts))
			return response.String()
		}
	}

	// ==================== PENDING ====================
	if strings.Contains(q, "pending") || strings.Contains(q, "menunggu") ||
		strings.Contains(q, "belum di approve") || strings.Contains(q, "belum disetujui") {

		if pending, ok := dbData["pending_transactions"]; ok {
			if pendingList, ok := pending.([]map[string]interface{}); ok {
				if len(pendingList) > 0 {
					totalQty := 0
					for _, t := range pendingList {
						if qty, ok := t["quantity"].(int); ok {
							totalQty += qty
						}
					}

					wantDetail := strings.Contains(q, "detail") || strings.Contains(q, "rinci") ||
						strings.Contains(q, "semua") || strings.Contains(q, "all") ||
						strings.Contains(q, "tampilkan") || strings.Contains(q, "show") ||
						strings.Contains(q, "list") || strings.Contains(q, "daftar") ||
						strings.Contains(q, "lengkap") || strings.Contains(q, "full") ||
						strings.Contains(q, "lihat") || strings.Contains(q, "lihatkan")

					if wantDetail {
						response.WriteString(fmt.Sprintf("⏳ **%d transactions PENDING APPROVAL**\n\n", len(pendingList)))
						response.WriteString(fmt.Sprintf("📦 Total units: **%d**\n\n", totalQty))

						response.WriteString("**📋 Latest pending transactions:**\n")
						maxShow := 10
						if len(pendingList) < maxShow {
							maxShow = len(pendingList)
						}
						for i, t := range pendingList {
							if i >= maxShow {
								response.WriteString(fmt.Sprintf("_... and %d more_\n", len(pendingList)-maxShow))
								break
							}
							response.WriteString(fmt.Sprintf("  %d. %s: **%d** units\n", i+1, t["product_name"], t["quantity"]))
							if requestedBy, ok := t["requested_by"].(string); ok {
								response.WriteString(fmt.Sprintf("     📝 Requested by: %s\n", requestedBy))
							}
						}

						if summary, ok := dbData["summary"].(map[string]interface{}); ok {
							response.WriteString("\n📊 **Summary:**\n")
							response.WriteString(fmt.Sprintf("  • Pending: **%d**\n", summary["pending_requests"]))
							response.WriteString(fmt.Sprintf("  • Approved: **%d**\n", summary["approved_requests"]))
							response.WriteString(fmt.Sprintf("  • Rejected: **%d**\n", summary["rejected_requests"]))
						}
						return response.String()
					}

					response.WriteString("📊 **Pending Approval Summary**\n\n")
					response.WriteString(fmt.Sprintf("  • Total pending: **%d**\n", len(pendingList)))
					response.WriteString(fmt.Sprintf("  • Total units: **%d**\n", totalQty))

					if summary, ok := dbData["summary"].(map[string]interface{}); ok {
						response.WriteString(fmt.Sprintf("  • Approved: **%d**\n", summary["approved_requests"]))
						response.WriteString(fmt.Sprintf("  • Rejected: **%d**\n", summary["rejected_requests"]))
					}

					response.WriteString("\n💡 _Type \"show pending details\" to see the list._")
					return response.String()
				}
				response.WriteString("✅ **No pending transactions.**")
				return response.String()
			}
		}
	}

	// ==================== HEAD USERS ====================
	if strings.Contains(q, "head") || strings.Contains(q, "superadmin") ||
		strings.Contains(q, "siapa") || strings.Contains(q, "who") ||
		strings.Contains(q, "daftar user") || strings.Contains(q, "list users") {

		response.WriteString("👤 **Users with Head & Superadmin Roles:**\n\n")

		rows, err := config.DB.Query(`
			SELECT id, username, fullname, role,
			       DATE_FORMAT(created_at, '%d %M %Y') as created_at
			FROM users
			WHERE role IN ('head', 'superadmin')
			ORDER BY
				CASE role
					WHEN 'superadmin' THEN 1
					WHEN 'head' THEN 2
				END,
				fullname ASC
		`)
		if err == nil {
			defer rows.Close()
			var headUsers []map[string]interface{}
			var superAdminUsers []map[string]interface{}

			for rows.Next() {
				var id int
				var username, fullname, role, createdAt string
				rows.Scan(&id, &username, &fullname, &role, &createdAt)

				userData := map[string]interface{}{
					"id":         id,
					"username":   username,
					"fullname":   fullname,
					"role":       role,
					"created_at": createdAt,
				}

				if role == "superadmin" {
					superAdminUsers = append(superAdminUsers, userData)
				} else {
					headUsers = append(headUsers, userData)
				}
			}

			if len(superAdminUsers) == 0 && len(headUsers) == 0 {
				response.WriteString("❌ **No users found with Head or Superadmin roles.**")
				return response.String()
			}

			if len(superAdminUsers) > 0 {
				response.WriteString("**🛡️ Super Admin:**\n")
				for i, u := range superAdminUsers {
					response.WriteString(fmt.Sprintf("  %d. **%s** (@%s)\n", i+1, u["fullname"], u["username"]))
				}
				response.WriteString("\n")
			}

			if len(headUsers) > 0 {
				response.WriteString("**👔 Head Department:**\n")
				for i, u := range headUsers {
					response.WriteString(fmt.Sprintf("  %d. **%s** (@%s)\n", i+1, u["fullname"], u["username"]))
				}
				response.WriteString("\n")
			}

			response.WriteString("📊 **Summary:**\n")
			response.WriteString(fmt.Sprintf("  • Total Head: **%d**\n", len(headUsers)))
			response.WriteString(fmt.Sprintf("  • Total Super Admin: **%d**\n", len(superAdminUsers)))
			response.WriteString(fmt.Sprintf("  • Total: **%d**\n", len(headUsers)+len(superAdminUsers)))

			return response.String()
		}
	}

	// ==================== APPROVED ====================
	if strings.Contains(q, "approved") || strings.Contains(q, "disetujui") ||
		strings.Contains(q, "sudah di approve") {

		if approved, ok := dbData["approved_transactions"]; ok {
			if approvedList, ok := approved.([]map[string]interface{}); ok {
				if len(approvedList) > 0 {
					response.WriteString(fmt.Sprintf("✅ **%d approved transactions**\n\n", len(approvedList)))

					totalQty := 0
					for _, t := range approvedList {
						if qty, ok := t["quantity"].(int); ok {
							totalQty += qty
						}
					}
					response.WriteString(fmt.Sprintf("📦 Total units: **%d**\n\n", totalQty))

					response.WriteString("**📋 Latest approved:**\n")
					for i, t := range approvedList {
						if i >= 5 {
							response.WriteString(fmt.Sprintf("_... and %d more_\n", len(approvedList)-5))
							break
						}
						response.WriteString(fmt.Sprintf("  %d. %s: **%d** units\n", i+1, t["product_name"], t["quantity"]))
						if approvedBy, ok := t["approved_by"].(string); ok && approvedBy != "" {
							response.WriteString(fmt.Sprintf("     ✅ By: %s\n", approvedBy))
						}
					}
					return response.String()
				}
				response.WriteString("✅ **No approved transactions found.**")
				return response.String()
			}
		}
	}

	// ==================== REJECTED ====================
	if strings.Contains(q, "rejected") || strings.Contains(q, "ditolak") {
		if rejected, ok := dbData["rejected_transactions"]; ok {
			if rejectedList, ok := rejected.([]map[string]interface{}); ok {
				if len(rejectedList) > 0 {
					response.WriteString(fmt.Sprintf("❌ **%d rejected transactions**\n\n", len(rejectedList)))
					for i, t := range rejectedList {
						if i >= 5 {
							response.WriteString(fmt.Sprintf("_... and %d more_\n", len(rejectedList)-5))
							break
						}
						response.WriteString(fmt.Sprintf("  %d. %s: **%d** units\n", i+1, t["product_name"], t["quantity"]))
						if rejectedBy, ok := t["rejected_by"].(string); ok && rejectedBy != "" {
							response.WriteString(fmt.Sprintf("     ❌ By: %s\n", rejectedBy))
						}
					}
					return response.String()
				}
				response.WriteString("✅ **No rejected transactions found.**")
				return response.String()
			}
		}
	}

	// ==================== TOP PRODUCTS ====================
	if strings.Contains(q, "top") || strings.Contains(q, "tertinggi") ||
		strings.Contains(q, "terbanyak") || strings.Contains(q, "highest") {

		if top, ok := dbData["top_products"]; ok {
			if topList, ok := top.([]map[string]interface{}); ok && len(topList) > 0 {
				response.WriteString("🏆 **Products with Highest Stock:**\n\n")
				for i, p := range topList {
					response.WriteString(fmt.Sprintf("  %d. **%s** — %d units\n", i+1, p["product_name"], p["quantity"]))
				}
				return response.String()
			}
		}
	}

	// ==================== LOW STOCK ====================
	if strings.Contains(q, "low stock") || strings.Contains(q, "menipis") ||
		strings.Contains(q, "kurang") || strings.Contains(q, "habis") ||
		strings.Contains(q, "lowest") || strings.Contains(q, "terendah") {

		if low, ok := dbData["low_stock"]; ok {
			if lowList, ok := low.([]map[string]interface{}); ok && len(lowList) > 0 {
				response.WriteString("⚠️ **Products with Low Stock Alert:**\n\n")
				for i, p := range lowList {
					response.WriteString(fmt.Sprintf("  %d. **%s** — %d units (Min: %d)\n", i+1, p["product_name"], p["quantity"], p["min_stock"]))
				}
				return response.String()
			}
			response.WriteString("✅ **All products are in healthy stock condition!** 🎉")
			return response.String()
		}
	}

	// ==================== SUMMARY ====================
	if strings.Contains(q, "summary") || strings.Contains(q, "ringkasan") ||
		strings.Contains(q, "overview") || strings.Contains(q, "rekap") ||
		strings.Contains(q, "stock report") || strings.Contains(q, "laporan") {

		if summary, ok := dbData["summary"].(map[string]interface{}); ok {
			response.WriteString("📊 **Inventory Summary**\n")
			response.WriteString("─" + strings.Repeat("─", 28) + "\n\n")
			response.WriteString(fmt.Sprintf("  • Total Products: **%d**\n", summary["total_products"]))
			response.WriteString(fmt.Sprintf("  • Total Stock: **%d** units\n", summary["total_stock"]))
			response.WriteString(fmt.Sprintf("  • Total Incoming: **%d**\n", summary["total_in"]))
			response.WriteString(fmt.Sprintf("  • Total Outgoing: **%d**\n", summary["total_out"]))
			response.WriteString(fmt.Sprintf("  • Pending: **%d**\n", summary["pending_requests"]))
			response.WriteString(fmt.Sprintf("  • Approved: **%d**\n", summary["approved_requests"]))
			response.WriteString(fmt.Sprintf("  • Rejected: **%d**\n", summary["rejected_requests"]))
			return response.String()
		}
	}

	// ==================== DEFAULT ====================
	response.WriteString("💡 **I can help you with:**\n\n")
	response.WriteString("  • 📊 \"Show stock summary\"\n")
	response.WriteString("  • 🏆 \"Products with highest stock\"\n")
	response.WriteString("  • ⏳ \"What transactions are pending approval?\"\n")
	response.WriteString("  • ✅ \"Show approved transactions\"\n")
	response.WriteString("  • 👤 \"Who are the Head users?\"\n")
	response.WriteString("  • ⚠️ \"Show low stock products\"\n\n")
	response.WriteString("_Try asking with specific keywords! 🚀_")

	return response.String()
}

// ==================== QUERY DATABASE ====================
func queryDatabase(query string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	q := strings.ToLower(query)

	stats, _ := getBasicStats()
	result["basic_stats"] = stats

	summary, _ := getSummary()
	result["summary"] = summary

	if strings.Contains(q, "pending") || strings.Contains(q, "approve") ||
		strings.Contains(q, "belum diapprove") || strings.Contains(q, "menunggu") {
		pending, _ := getPendingTransactions()
		result["pending_transactions"] = pending
	}

	if strings.Contains(q, "approved") || strings.Contains(q, "disetujui") ||
		strings.Contains(q, "sudah diapprove") {
		approved, _ := getApprovedTransactions()
		result["approved_transactions"] = approved
	}

	if strings.Contains(q, "rejected") || strings.Contains(q, "ditolak") {
		rejected, _ := getRejectedTransactions()
		result["rejected_transactions"] = rejected
	}

	if strings.Contains(q, "top") || strings.Contains(q, "tertinggi") ||
		strings.Contains(q, "terbanyak") || strings.Contains(q, "highest") {
		top, _ := getTopProducts(5)
		result["top_products"] = top
	}

	if strings.Contains(q, "low stock") || strings.Contains(q, "menipis") ||
		strings.Contains(q, "kurang") || strings.Contains(q, "habis") ||
		strings.Contains(q, "lowest") || strings.Contains(q, "terendah") {
		low, _ := getLowStock()
		result["low_stock"] = low
	}

	return result, nil
}

// ==================== DATABASE QUERIES ====================

func getTopProducts(limit int) ([]map[string]interface{}, error) {
	rows, err := config.DB.Query(`
		SELECT p.product_code, p.product_name, p.category, COALESCE(s.quantity, 0) as quantity
		FROM products p
		LEFT JOIN stock s ON p.id = s.product_id
		ORDER BY quantity DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var code, name, category string
		var quantity int
		rows.Scan(&code, &name, &category, &quantity)
		products = append(products, map[string]interface{}{
			"product_code": code,
			"product_name": name,
			"category":     category,
			"quantity":     quantity,
		})
	}
	return products, nil
}

func getPendingTransactions() ([]map[string]interface{}, error) {
	rows, err := config.DB.Query(`
		SELECT t.id, p.product_name, t.quantity, t.note, u.fullname as requested_by, t.created_at
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		WHERE t.status = 'pending'
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, quantity int
		var productName, note, requestedBy, createdAt string
		rows.Scan(&id, &productName, &quantity, &note, &requestedBy, &createdAt)
		transactions = append(transactions, map[string]interface{}{
			"id":           id,
			"product_name": productName,
			"quantity":     quantity,
			"note":         note,
			"requested_by": requestedBy,
			"created_at":   createdAt,
		})
	}
	return transactions, nil
}

func getApprovedTransactions() ([]map[string]interface{}, error) {
	rows, err := config.DB.Query(`
		SELECT t.id, p.product_name, t.quantity, t.note, u.fullname as requested_by,
			   COALESCE(a.fullname, '') as approved_by, t.created_at, t.approved_at
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		LEFT JOIN users a ON t.approved_by = a.id
		WHERE t.status = 'approved'
		ORDER BY t.approved_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, quantity int
		var productName, note, requestedBy, approvedBy, createdAt, approvedAt string
		rows.Scan(&id, &productName, &quantity, &note, &requestedBy, &approvedBy, &createdAt, &approvedAt)
		transactions = append(transactions, map[string]interface{}{
			"id":           id,
			"product_name": productName,
			"quantity":     quantity,
			"note":         note,
			"requested_by": requestedBy,
			"approved_by":  approvedBy,
			"created_at":   createdAt,
			"approved_at":  approvedAt,
		})
	}
	return transactions, nil
}

func getRejectedTransactions() ([]map[string]interface{}, error) {
	rows, err := config.DB.Query(`
		SELECT t.id, p.product_name, t.quantity, t.note, u.fullname as requested_by,
			   COALESCE(a.fullname, '') as rejected_by, t.created_at, t.approved_at
		FROM transaction_out t
		JOIN products p ON t.product_id = p.id
		JOIN users u ON t.created_by = u.id
		LEFT JOIN users a ON t.approved_by = a.id
		WHERE t.status = 'rejected'
		ORDER BY t.approved_at DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, quantity int
		var productName, note, requestedBy, rejectedBy, createdAt, rejectedAt string
		rows.Scan(&id, &productName, &quantity, &note, &requestedBy, &rejectedBy, &createdAt, &rejectedAt)
		transactions = append(transactions, map[string]interface{}{
			"id":           id,
			"product_name": productName,
			"quantity":     quantity,
			"note":         note,
			"requested_by": requestedBy,
			"rejected_by":  rejectedBy,
			"created_at":   createdAt,
			"rejected_at":  rejectedAt,
		})
	}
	return transactions, nil
}

func getSummary() (map[string]interface{}, error) {
	var totalProducts, totalStock, totalIn, totalOut, pendingOut, approvedOut, rejectedOut int

	// Total Products dari database (harusnya 1000)
	config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM stock").Scan(&totalStock)
	config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_in").Scan(&totalIn)
	config.DB.QueryRow("SELECT COALESCE(SUM(quantity), 0) FROM transaction_out WHERE status = 'approved'").Scan(&totalOut)
	config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'pending'").Scan(&pendingOut)
	config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'approved'").Scan(&approvedOut)
	config.DB.QueryRow("SELECT COUNT(*) FROM transaction_out WHERE status = 'rejected'").Scan(&rejectedOut)

	return map[string]interface{}{
		"total_products":    totalProducts,
		"total_stock":       totalStock,
		"total_in":          totalIn,
		"total_out":         totalOut,
		"pending_requests":  pendingOut,
		"approved_requests": approvedOut,
		"rejected_requests": rejectedOut,
		"net_stock":         totalStock,
	}, nil
}

func getLowStock() ([]map[string]interface{}, error) {
	rows, err := config.DB.Query(`
		SELECT p.product_code, p.product_name, p.min_stock, COALESCE(s.quantity, 0) as quantity
		FROM products p
		LEFT JOIN stock s ON p.id = s.product_id
		WHERE COALESCE(s.quantity, 0) <= p.min_stock
		ORDER BY quantity ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var code, name string
		var minStock, quantity int
		rows.Scan(&code, &name, &minStock, &quantity)
		products = append(products, map[string]interface{}{
			"product_code": code,
			"product_name": name,
			"min_stock":    minStock,
			"quantity":     quantity,
		})
	}
	return products, nil
}

func getBasicStats() (map[string]interface{}, error) {
	var totalProducts, totalUsers int
	config.DB.QueryRow("SELECT COUNT(*) FROM products").Scan(&totalProducts)
	config.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)

	return map[string]interface{}{
		"total_products": totalProducts,
		"total_users":    totalUsers,
	}, nil
}

func getStockStatus(quantity, minStock int) string {
	if quantity == 0 {
		return "Out of Stock"
	}
	if quantity <= minStock {
		return "Low Stock"
	}
	return "Healthy"
}