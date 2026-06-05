package main

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/luncher4/vpn-panel/internal/models"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/vpn.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Миграция
	db.AutoMigrate(
		&models.User{}, &models.Node{}, &models.Inbound{},
		&models.SubscriptionPlan{}, &models.Subscription{},
		&models.Device{}, &models.TrafficLog{}, &models.TrafficDaily{},
		&models.EventLog{}, &models.Notification{},
	)

	// Создаём админа, если нет
	var admin models.User
	if err := db.Where("email = ?", "admin").First(&admin).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		admin = models.User{
			UUID:         uuid.New().String(),
			Email:        "admin",
			PasswordHash: string(hash),
			Role:         "admin",
		}
		db.Create(&admin)
		log.Println("Admin created: admin / admin")
	}

	// Создаём дефолтную ноду, если нет
	var existing models.Node
	if err := db.First(&existing).Error; err != nil {
		node := models.Node{
			Name:     "main-vps",
			Address:  "45.134.39.18",
			APIPort:  8443,
			Location: "russia",
			IsActive: true,
		}
		db.Create(&node)
		log.Println("Default node created: main-vps")
	}

	// Дефолтные тарифные планы
	var plans []models.SubscriptionPlan
	db.Find(&plans)
	if len(plans) == 0 {
		db.Create(&[]models.SubscriptionPlan{
			{Name: "Триал", DurationDays: 1, TrafficLimit: 1e9, DeviceLimit: 1, Price: 0},
			{Name: "Недельный", DurationDays: 7, TrafficLimit: 50e9, DeviceLimit: 3, Price: 100},
			{Name: "Месячный", DurationDays: 30, TrafficLimit: 500e9, DeviceLimit: 10, Price: 500},
			{Name: "Годовой", DurationDays: 365, TrafficLimit: 0, DeviceLimit: 0, Price: 3000},
		})
		log.Println("Default plans created")
	}

	log.Println("Seed complete")
}
