package models

import "time"

type TransactionIn struct {
    ID        int       `json:"id"`
    ProductID int       `json:"product_id"`
    Product   Product   `json:"product,omitempty"`
    Quantity  int       `json:"quantity"`
    Note      string    `json:"note"`
    CreatedBy int       `json:"created_by"`
    CreatedByName string    `json:"created_by_name"`
    CreatedAt time.Time `json:"created_at"`
}

type TransactionOut struct {
    ID            int        `json:"id"`
    ProductID     int        `json:"product_id"`
    Product       Product    `json:"product,omitempty"`
    Quantity      int        `json:"quantity"`
    Note          string     `json:"note"`
    Status        string     `json:"status"` // pending, approved, rejected
    CreatedBy     int        `json:"created_by"`
    CreatedByName string     `json:"created_by_name"` // Tambahan untuk nama user
    
    ApprovedBy    *int       `json:"approved_by"`
    CreatedAt     time.Time  `json:"created_at"`
    ApprovedAt    *time.Time `json:"approved_at"`
}

type TransactionInRequest struct {
    ProductID int    `json:"product_id" binding:"required"`
    Quantity  int    `json:"quantity" binding:"required,min=1"`
    Note      string `json:"note"`
}

type TransactionOutRequest struct {
    ProductID int    `json:"product_id" binding:"required"`
    Quantity  int    `json:"quantity" binding:"required,min=1"`
    Note      string `json:"note"`
}

type ApproveRequest struct {
    Status string `json:"status" binding:"required,oneof=approved rejected"`
    Note   string `json:"note"`
}