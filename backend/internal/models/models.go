package models

import (
	"time"
)

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UUID         string     `gorm:"uniqueIndex;size:36" json:"uuid"`
	Name         string     `gorm:"size:100" json:"name"`
	Email        string     `gorm:"size:255" json:"email,omitempty"`
	PasswordHash string     `gorm:"size:255" json:"-"`
	Role         string     `gorm:"size:20;default:user" json:"role"`
	Status       string     `gorm:"size:20;default:active" json:"status"`
	Note         string     `gorm:"size:500" json:"note,omitempty"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`
	LastIP       string     `gorm:"size:45" json:"last_ip,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Node struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100" json:"name"`
	Address   string    `gorm:"size:255" json:"address"`
	APIPort   int       `gorm:"default:8443" json:"api_port"`
	APIKey    string    `gorm:"size:255" json:"-"`
	Location  string    `gorm:"size:100" json:"location"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Inbound struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NodeID    uint      `json:"node_id"`
	Protocol  string    `gorm:"size:50" json:"protocol"`  // vless | shadowsocks | trojan | hysteria2
	Transport string    `gorm:"size:50" json:"transport"` // tcp | ws | grpc | h2 | quic
	Port      int       `json:"port"`
	Settings  string    `gorm:"type:text" json:"settings"` // JSON — гибкие настройки протокола
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:100" json:"name"`
	DurationDays int       `json:"duration_days"`
	TrafficLimit int64     `json:"traffic_limit"` // байт, 0 = безлимит
	DeviceLimit  int       `json:"device_limit"`  // 0 = безлимит
	Price        float64   `json:"price"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subscription struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `json:"user_id"`
	PlanID      uint      `json:"plan_id"`
	Status      string    `gorm:"size:20;default:active" json:"status"`
	StartedAt   time.Time `json:"started_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	AutoRenew   bool      `gorm:"default:false" json:"auto_renew"`
	TrafficUsed int64     `gorm:"default:0" json:"traffic_used"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Device struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `json:"user_id"`
	Name            string     `gorm:"size:100" json:"name"`
	DeviceType      string     `gorm:"size:20" json:"device_type"` // android | ios | windows | mac | linux
	SubscriptionURL string     `gorm:"size:500" json:"subscription_url"`
	LastConnectedAt *time.Time `json:"last_connected_at,omitempty"`
	LastIP          string     `gorm:"size:45" json:"last_ip,omitempty"`
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TrafficLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `json:"user_id"`
	InboundID     uint      `json:"inbound_id"`
	NodeID        uint      `json:"node_id"`
	UploadBytes   int64     `json:"upload_bytes"`
	DownloadBytes int64     `json:"download_bytes"`
	RecordedAt    time.Time `gorm:"index" json:"recorded_at"`
}

type TrafficDaily struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `json:"user_id"`
	NodeID        uint   `json:"node_id"`
	Date          string `gorm:"size:10" json:"date"`
	UploadTotal   int64  `json:"upload_total"`
	DownloadTotal int64  `json:"download_total"`
}


type SubscriptionProfile struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	UserID          uint       `json:"user_id"`
	Name            string     `gorm:"size:100" json:"name"`
	UUID            string     `gorm:"uniqueIndex;size:36" json:"uuid"`
	PlanID          uint       `json:"plan_id"`
	NodeID          uint       `json:"node_id"`
	InboundIDs      string     `gorm:"type:text" json:"inbound_ids"` // JSON array: [1, 3, 5]
	ExpiresAt       time.Time  `json:"expires_at"`
	Status          string     `gorm:"size:20;default:active" json:"status"`
	TrafficUsed     int64      `gorm:"default:0" json:"traffic_used"`
	SubscriptionURL string     `gorm:"size:500" json:"subscription_url"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type EventLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `json:"user_id,omitempty"`
	Action    string    `gorm:"size:100;index" json:"action"`
	Details   string    `gorm:"type:text" json:"details,omitempty"`
	IPAddress string    `gorm:"size:45" json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	Type      string    `gorm:"size:20" json:"type"` // info | warning | error | success
	Title     string    `gorm:"size:255" json:"title"`
	Body      string    `gorm:"type:text" json:"body"`
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
