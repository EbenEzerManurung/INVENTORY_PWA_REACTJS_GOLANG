package models

import "time"

type Stock struct {
    ID          int       `json:"id"`
    ProductID   int       `json:"product_id"`
    Product     Product   `json:"product,omitempty"`
    Quantity    int       `json:"quantity"`
    LastUpdated time.Time `json:"last_updated"`
}

type StockReport struct {
    ProductID   int    `json:"product_id"`
    ProductCode string `json:"product_code"`
    ProductName string `json:"product_name"`
    Category    string `json:"category"`
    Unit        string `json:"unit"`
    MinStock    int    `json:"min_stock"`
    Quantity    int    `json:"quantity"`
    TotalIn     int    `json:"total_in"`
    TotalOut    int    `json:"total_out"`
    Status      string `json:"status"`
}