package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/luncher4/vpn-panel/internal/models"
	"github.com/luncher4/vpn-panel/internal/services"
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

		// Статистика трафика через Xray gRPC
		traffic := map[string]int64{"up": 0, "down": 0}
		if client, err := services.NewXrayStatsClient(); err == nil {
			up, down, err := client.GetServerTraffic()
			if err == nil {
				traffic["up"] = up
				traffic["down"] = down
			}
			client.Close()
		}

		// Статистика по пользователям из TrafficLog за сегодня
		var todayUp, todayDown int64
		today := time.Now().Format("2006-01-02")
		db.Model(&models.TrafficDaily{}).Where("date = ?", today).
			Select("COALESCE(SUM(upload_total), 0)").Scan(&todayUp)
		db.Model(&models.TrafficDaily{}).Where("date = ?", today).
			Select("COALESCE(SUM(download_total), 0)").Scan(&todayDown)

		// Активные подписки с истекающим сроком (ближайшие 7 дней)
		var expiringSoon int64
		db.Model(&models.Subscription{}).Where("status = ? AND expires_at BETWEEN ? AND ?",
			"active", time.Now(), time.Now().Add(7*24*time.Hour)).Count(&expiringSoon)

		c.JSON(http.StatusOK, gin.H{
			"total_users":   userCount,
			"active_subs":   activeSubs,
			"expiring_soon": expiringSoon,
			"traffic_today": gin.H{"up_gb": int(todayUp / 1e9), "down_gb": int(todayDown / 1e9)},
			"xray_traffic":  gin.H{"up_gb": int(traffic["up"] / 1e9), "down_gb": int(traffic["down"] / 1e9)},
			"main_node":     nodeStats,
			"server_time":   time.Now(),
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



// ========== Subscription Profiles ==========

func ListProfiles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var profiles []models.SubscriptionProfile
		db.Preload("User").Preload("Plan").Order("created_at desc").Find(&profiles)
		c.JSON(http.StatusOK, profiles)
	}
}

func CreateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type CreateRequest struct {
			Name       string   `json:"name" binding:"required"`
			UserID     uint     `json:"user_id"`
			PlanID     uint     `json:"plan_id" binding:"required"`
			NodeID     uint     `json:"node_id"`
			InboundIDs []uint   `json:"inbound_ids"`
			ExpiresAt  string   `json:"expires_at"`
		}
		var req CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user := models.User{}
		if req.UserID > 0 {
			db.First(&user, req.UserID)
		}
		if user.ID == 0 {
			user = models.User{
				UUID: uuid.New().String(),
				Name: req.Name,
			}
			db.Create(&user)
		}

		profile := models.SubscriptionProfile{
			UserID: user.ID,
			Name:   req.Name,
			UUID:   user.UUID,
			PlanID: req.PlanID,
			NodeID: req.NodeID,
			Status: "active",
		}

		if len(req.InboundIDs) > 0 {
			data, _ := json.Marshal(req.InboundIDs)
			profile.InboundIDs = string(data)
		}

		if req.ExpiresAt != "" {
			t, err := time.Parse("2006-01-02", req.ExpiresAt)
			if err == nil {
				profile.ExpiresAt = t
			} else {
				profile.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
			}
		} else {
			profile.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
		}

		profile.SubscriptionURL = fmt.Sprintf("https://hvq.airydeck.su/sub/%s", user.UUID)
		db.Create(&profile)

		xray := services.NewXrayManager()
		for _, inboundID := range req.InboundIDs {
			var inbound models.Inbound
			if db.First(&inbound, inboundID).Error == nil {
				xray.AddClientToInbound(inbound.Port, user.UUID)
			}
		}

		c.JSON(http.StatusCreated, profile)
	}
}





// ========== Nodes ==========

func ListNodes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var nodes []models.Node
		db.Order("name asc").Find(&nodes)
		c.JSON(http.StatusOK, nodes)
	}
}

func CreateNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var node models.Node
		if err := c.ShouldBindJSON(&node); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&node)
		c.JSON(http.StatusCreated, node)
	}
}

func DeleteNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Node{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func ToggleInbound(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var inbound models.Inbound
		if err := db.First(&inbound, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		inbound.IsActive = !inbound.IsActive
		db.Save(&inbound)
		c.JSON(http.StatusOK, inbound)
	}
}

func ToggleNode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var node models.Node
		if err := db.First(&node, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		node.IsActive = !node.IsActive
		db.Save(&node)
		c.JSON(http.StatusOK, node)
	}
}

func UpdateProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var profile models.SubscriptionProfile
		if err := db.First(&profile, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		var req struct {
			Name       string   `json:"name"`
			PlanID     uint     `json:"plan_id"`
			NodeID     uint     `json:"node_id"`
			InboundIDs []uint   `json:"inbound_ids"`
			ExpiresAt  string   `json:"expires_at"`
			Status     string   `json:"status"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Name != "" { profile.Name = req.Name }
		if req.PlanID > 0 { profile.PlanID = req.PlanID }
		if req.NodeID > 0 { profile.NodeID = req.NodeID }
		if len(req.InboundIDs) > 0 {
			data, _ := json.Marshal(req.InboundIDs)
			profile.InboundIDs = string(data)
		}
		if req.ExpiresAt != "" {
			t, err := time.Parse("2006-01-02", req.ExpiresAt)
			if err == nil { profile.ExpiresAt = t }
		}
		if req.Status != "" { profile.Status = req.Status }
		db.Save(&profile)
		c.JSON(http.StatusOK, profile)
	}
}

func DeleteProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var profile models.SubscriptionProfile
		db.First(&profile, id)
		db.Delete(&models.SubscriptionProfile{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func GetProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var profile models.SubscriptionProfile
		if err := db.Preload("User").Preload("Plan").First(&profile, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		var inboundIDs []uint
		json.Unmarshal([]byte(profile.InboundIDs), &inboundIDs)
		var inbounds []models.Inbound
		db.Where("id IN ?", inboundIDs).Find(&inbounds)
		c.JSON(http.StatusOK, gin.H{"profile": profile, "inbounds": inbounds})
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

		// Валидация перед сохранением
		if err := services.ValidateInboundConfig(inbound.Protocol, inbound.Port, inbound.Security, inbound.Transport); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Если security не указан — ставим REALITY для VLESS
		if inbound.Security == "" && inbound.Protocol == "vless" {
			inbound.Security = "reality"
		}

		db.Create(&inbound)

		// Синхронизируем с Xray
		xray := services.NewXrayManager()
		if err := xray.AddInbounds([]models.Inbound{inbound}); err != nil {
			c.JSON(http.StatusOK, gin.H{"inbound": inbound, "xray_warning": err.Error()})
			return
		}

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


// ========== Subscription Plans ==========

func ListPlans(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var plans []models.SubscriptionPlan
		db.Order("price asc").Find(&plans)
		c.JSON(http.StatusOK, plans)
	}
}

func CreatePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var plan models.SubscriptionPlan
		if err := c.ShouldBindJSON(&plan); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Create(&plan)
		c.JSON(http.StatusCreated, plan)
	}
}

func DeletePlan(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.SubscriptionPlan{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

// ========== Subscriptions ==========

func ListSubscriptions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var subs []models.Subscription
		db.Preload("User").Preload("Plan").Order("created_at desc").Find(&subs)
		c.JSON(http.StatusOK, subs)
	}
}

func CreateSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sub models.Subscription
		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub.StartedAt = time.Now()
		if sub.ExpiresAt.IsZero() {
			sub.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
		}
		db.Create(&sub)
		c.JSON(http.StatusCreated, sub)
	}
}

func DeleteSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		db.Delete(&models.Subscription{}, id)
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func DeleteInbound(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var inbound models.Inbound
		db.First(&inbound, id)

		db.Delete(&models.Inbound{}, id)

		if inbound.Port > 0 {
			xray := services.NewXrayManager()
			xray.RemoveInbound(inbound.Port)
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}



// GenerateSubscriptionConfig отдаёт REALITY JSON конфиг по UUID
type ConfigEntry struct {
	Protocol string      `json:"protocol"`
	Config   interface{} `json:"config"`
}



// GetTemplates возвращает список шаблонов конфигов
func GetTemplates() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, services.GetConfigTemplates())
	}
}

// ValidateConfig проверяет конфиг inbound перед сохранением
func ValidateConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Protocol string `json:"protocol"`
			Port     int    `json:"port"`
			Security string `json:"security"`
			Transport string `json:"transport"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": err.Error()})
			return
		}
		if err := services.ValidateInboundConfig(req.Protocol, req.Port, req.Security, req.Transport); err != nil {
			c.JSON(http.StatusOK, gin.H{"valid": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"valid": true})
	}
}

func GenerateSubscriptionConfig(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		var profile models.SubscriptionProfile
		if err := db.Where("uuid = ? AND status = ?", uuid, "active").First(&profile).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found or inactive"})
			return
		}

		var node models.Node
		if err := db.First(&node, profile.NodeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
			return
		}

		// Парсим inbound IDs
		var inboundIDs []uint
		json.Unmarshal([]byte(profile.InboundIDs), &inboundIDs)
		var inbounds []models.Inbound
		db.Where("id IN ? AND is_active = ?", inboundIDs, true).Find(&inbounds)

		if len(inbounds) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no active inbounds"})
			return
		}

		// Определяем User-Agent для выбора формата
		ua := c.GetHeader("User-Agent")

		// Генерируем конфиги для каждого inbound
		var configs []ConfigEntry
		for _, ib := range inbounds {
			if ib.Protocol == "vless" && ib.Transport == "tcp" {
				// REALITY конфиг
				cfg := map[string]interface{}{
					"v": "2",
					"ps": profile.Name,
					"add": node.Address,
					"port": ib.Port,
					"id": uuid,
					"aid": 0,
					"scy": "auto",
					"net": "tcp",
					"type": "none",
					"tls": "reality",
					"flow": "xtls-rprx-vision",
					"sni": "www.microsoft.com",
					"pbk": services.GetPublicKey(),
					"sid": "19e72187",
					"fp": "chrome",
				}
				// Парсим доп. настройки из БД
				if ib.Settings != "" && ib.Settings != "{}" {
					var extra map[string]interface{}
					if json.Unmarshal([]byte(ib.Settings), &extra) == nil {
						if fp, ok := extra["fingerprint"]; ok { cfg["fp"] = fp }
						if sni, ok := extra["serverName"]; ok { cfg["sni"] = sni }
					}
				}
				configs = append(configs, ConfigEntry{Protocol: "vless", Config: cfg})
			}
		}

		// Формат ответа в зависимости от User-Agent
		if isClashUA(ua) {
			c.YAML(http.StatusOK, generateClashConfig(configs, node))
			return
		}

		// Стандартный JSON (v2rayNG / Nekobox)
		if len(configs) == 1 {
			c.JSON(http.StatusOK, configs[0].Config)
		} else {
			c.JSON(http.StatusOK, configs)
		}
	}
}

func isClashUA(ua string) bool {
	return strings.Contains(ua, "clash") || strings.Contains(ua, "Clash") || strings.Contains(ua, "stash") || strings.Contains(ua, "sing-box")
}

func generateClashConfig(configs []ConfigEntry, node models.Node) interface{} {
	proxies := make([]map[string]interface{}, 0)
	for _, entry := range configs {
		if c, ok := entry.Config.(map[string]interface{}); ok {
			proxy := map[string]interface{}{
				"name": entry.Protocol + "-" + fmt.Sprintf("%v", c["port"]),
				"type": "vless",
				"server": c["add"],
				"port": c["port"],
				"uuid": c["id"],
				"flow": c["flow"],
				"tls": true,
				"servername": c["sni"],
				"reality-opts": map[string]interface{}{
					"public-key": c["pbk"],
					"short-id": c["sid"],
				},
				"client-fingerprint": c["fp"],
			}
			proxies = append(proxies, proxy)
		}
	}
	return map[string]interface{}{
		"proxies": proxies,
		"proxy-groups": []map[string]interface{}{
			{"name": "Proxy", "type": "select", "proxies": []string{}},
		},
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
