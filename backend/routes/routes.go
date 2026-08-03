package routes

import (
	"inventory-backend/controllers"
	"inventory-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")

	// ==================== PUBLIC ROUTES ====================
	api.POST("/auth/login", controllers.Login)
	api.POST("/auth/register", controllers.Register)
	api.POST("/auth/refresh", controllers.RefreshToken)

	// ==================== PROTECTED ROUTES ====================
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())

	// ==================== AUTH ROUTES ====================
	protected.GET("/auth/me", controllers.GetMe)

	// ==================== USER PROFILE ROUTES ====================
	profileRoutes := protected.Group("/users/profile")
	{
		profileRoutes.GET("/", controllers.GetCurrentUserProfile)
		profileRoutes.PUT("/", controllers.UpdateUserProfile)
		profileRoutes.PUT("/password", controllers.ChangePassword)
		profileRoutes.POST("/upload-avatar", controllers.UploadAvatar)
		profileRoutes.DELETE("/avatar", controllers.RemoveAvatar)
	}

	// ==================== USER MANAGEMENT ROUTES ====================
	userRoutes := protected.Group("/users")
	userRoutes.Use(middleware.RoleMiddleware("superadmin"))
	{
		userRoutes.GET("/", controllers.GetUsers)
		userRoutes.GET("/:id", controllers.GetUserByID)
		userRoutes.POST("/", controllers.CreateUser)
		userRoutes.PUT("/:id", controllers.UpdateUser)
		userRoutes.DELETE("/:id", controllers.DeleteUser)
		userRoutes.PUT("/:id/role", controllers.UpdateUserRole)
		userRoutes.PUT("/:id/reset-password", controllers.ResetUserPassword)
	}

	// ==================== PRODUCT ROUTES ====================
	productRoutes := protected.Group("/products")
	{
		productRoutes.GET("/", controllers.GetProducts)
		productRoutes.GET("/:id", controllers.GetProductByID)
		productRoutes.POST("/", middleware.RoleMiddleware("superadmin", "head"), controllers.CreateProduct)
		productRoutes.PUT("/:id", middleware.RoleMiddleware("superadmin", "head"), controllers.UpdateProduct)
		productRoutes.DELETE("/:id", middleware.RoleMiddleware("superadmin"), controllers.DeleteProduct)
		productRoutes.GET("/export", controllers.ExportProducts)
	}

	// ==================== STOCK ROUTES ====================
	stockRoutes := protected.Group("/stock")
	{
		stockRoutes.GET("/", controllers.GetStock)
		stockRoutes.GET("/product/:productId", controllers.GetStockByProductID)
		stockRoutes.GET("/export", controllers.ExportStock)
		stockRoutes.PUT("/product/:productId",
			middleware.RoleMiddleware("superadmin", "head"),
			controllers.AdjustStock)
	}

	// ==================== TRANSACTION IN ROUTES ====================
	transInRoutes := protected.Group("/transactions/in")
	{
		transInRoutes.GET("/", controllers.GetTransactionsIn)
		transInRoutes.GET("/:id", controllers.GetTransactionInByID)
		transInRoutes.POST("/", controllers.CreateTransactionIn)
		transInRoutes.PUT("/:id",
			middleware.RoleMiddleware("superadmin", "head"),
			controllers.UpdateTransactionIn)
		transInRoutes.DELETE("/:id",
			middleware.RoleMiddleware("superadmin"),
			controllers.DeleteTransactionIn)
		transInRoutes.GET("/export", controllers.ExportTransactionsIn)
	}

	// ==================== TRANSACTION OUT ROUTES ====================
	transOutRoutes := protected.Group("/transactions/out")
	{
		transOutRoutes.GET("/", controllers.GetTransactionsOut)
		transOutRoutes.GET("/:id", controllers.GetTransactionOutByID)
		transOutRoutes.POST("/", controllers.CreateTransactionOut)
		transOutRoutes.PUT("/:id/approve",
			middleware.RoleMiddleware("superadmin", "head"),
			controllers.ApproveTransactionOut)
		transOutRoutes.PUT("/:id/reject",
			middleware.RoleMiddleware("superadmin", "head"),
			controllers.RejectTransactionOut)
		transOutRoutes.PUT("/:id/cancel",
			middleware.RoleMiddleware("superadmin", "head", "produksi"),
			controllers.CancelTransactionOut)
		transOutRoutes.GET("/export", controllers.ExportTransactionsOut)
		transOutRoutes.GET("/statistics", controllers.GetTransactionOutStatistics)
	}

	// ==================== AI CHAT ROUTES ====================
	aiRoutes := protected.Group("/ai")
	{
		aiRoutes.POST("/chat", controllers.AIChat)
	}

	// ==================== REPORT ROUTES ====================
	reportRoutes := protected.Group("/reports")
	{
		reportRoutes.GET("/stock", controllers.GetStockReport)
		reportRoutes.GET("/stock/export", controllers.ExportStockReport)
		reportRoutes.GET("/dashboard", controllers.GetDashboardStats)
		reportRoutes.GET("/transactions", controllers.GetTransactionReport)
		reportRoutes.GET("/products/category", controllers.GetProductCategoryReport)
		reportRoutes.GET("/summary", controllers.GetSummaryReport)
	}
}