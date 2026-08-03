package models

import "time"

type Product struct {
    ID          int       `json:"id"`
    ProductCode string    `json:"product_code"`
    ProductName string    `json:"product_name"`
    Category    string    `json:"category"`
    Unit        string    `json:"unit"`
    MinStock    int       `json:"min_stock"`
    Description string    `json:"description"`
    Image       string    `json:"image"`
    CreatedAt   time.Time `json:"created_at"`
}

type ProductRequest struct {
    ProductCode string `json:"product_code" binding:"required"`
    ProductName string `json:"product_name" binding:"required"`
    Category    string `json:"category"`
    Unit        string `json:"unit" binding:"required"`
    MinStock    int    `json:"min_stock"`
    Description string `json:"description"`
    Image       string `json:"image"`
}