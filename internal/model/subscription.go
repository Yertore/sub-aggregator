package model

import (
	"time"
)

type Subscription struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserID      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name,omitempty"`
	Price       int    `json:"price,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type ListFilter struct {
	UserID      string
	ServiceName string
	Limit       int
	Offset      int
}