package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/luncher4/vpn-panel/internal/models"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func Login(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		now := time.Now()
		db.Model(&user).Updates(map[string]interface{}{
			"last_seen_at": now,
			"last_ip":      c.ClientIP(),
		})

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"role":    user.Role,
			"exp":     now.Add(24 * time.Hour).Unix(),
		})
		tokenStr, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Логируем событие входа
		db.Create(&models.EventLog{
			UserID:    &user.ID,
			Action:    "user_login",
			Details:   "Login successful",
			IPAddress: c.ClientIP(),
		})

		c.JSON(http.StatusOK, LoginResponse{Token: tokenStr, User: user})
	}
}

func Dashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userCount, activeSubs int64
		db.Model(&models.User{}).Count(&userCount)
		db.Model(&models.Subscription{}).Where("status = ?", "active").Count(&activeSubs)

		var node models.Node
		nodeStats := map[string]interface{}{}
		if err := db.First(&node).Error; err == nil {
			nodeStats["name"] = node.Name
			nodeStats["location"] = node.Location
			nodeStats["address"] = node.Address
		}

		c.JSON(http.StatusOK, gin.H{
			"total_users": userCount,
			"active_subs": activeSubs,
			"main_node":   nodeStats,
			"server_time": time.Now(),
		})
	}
}

func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		db.Order("created_at desc").Find(&users)
		c.JSON(http.StatusOK, users)
	}
}

func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.PasswordHash = string(hash)
		user.UUID = uuid.New().String()

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		db.Create(&models.EventLog{
			Action:    "user_created",
			Details:   "Created user: " + user.Email,
			IPAddress: c.ClientIP(),
		})

		c.JSON(http.StatusCreated, user)
	}
}

func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.Save(&user)
		c.JSON(http.StatusOK, user)
	}
}

func DeleteUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.User{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func UserStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var totalUpload, totalDownload int64
		db.Model(&models.TrafficLog{}).Where("user_id = ?", id).Select("COALESCE(SUM(upload_bytes), 0)").Scan(&totalUpload)
		db.Model(&models.TrafficLog{}).Where("user_id = ?", id).Select("COALESCE(SUM(download_bytes), 0)").Scan(&totalDownload)

		var deviceCount int64
		db.Model(&models.Device{}).Where("user_id = ? AND is_active = ?", id, true).Count(&deviceCount)

		var sub models.Subscription
		db.Where("user_id = ?", id).First(&sub)

		c.JSON(http.StatusOK, gin.H{
			"user_id":      id,
			"upload_gb":    float64(totalUpload) / 1e9,
			"download_gb":  float64(totalDownload) / 1e9,
			"devices":      deviceCount,
			"subscription": sub,
		})
	}
}

func ListInbounds(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var inbounds []models.Inbound
		db.Preload("Node").Order("port asc").Find(&inbounds)
		c.JSON(http.StatusOK, inbounds)
	}
}

func CreateInbound(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var inbound models.Inbound
		if err := c.ShouldBindJSON(&inbound); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&inbound)
		c.JSON(http.StatusCreated, inbound)
	}
}

func UpdateInbound(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var inbound models.Inbound
		if err := db.First(&inbound, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "inbound not found"})
			return
		}
		if err := c.ShouldBindJSON(&inbound); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Save(&inbound)
		c.JSON(http.StatusOK, inbound)
	}
}

func DeleteInbound(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Inbound{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func GenerateConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var inbounds []models.Inbound
		db.Where("is_active = ?", true).Find(&inbounds)

		if len(inbounds) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active inbounds"})
			return
		}

		// TODO: реальная генерация конфига под каждый протокол
		_ = user
		c.JSON(http.StatusOK, gin.H{
			"message":  "config generation stub",
			"user_id":  userID,
			"inbounds": len(inbounds),
		})
	}
}

func GenerateQR(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		// TODO: генерация QR-кода
		c.JSON(http.StatusOK, gin.H{
			"message": "qr generation stub",
			"user_id": userID,
		})
	}
}
