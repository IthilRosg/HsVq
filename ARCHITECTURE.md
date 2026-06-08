# HsVq Panel - Architecture

Date: 8 June 2026
Stack: Go + Gin + GORM/SQLite + React + TypeScript + Tailwind CSS
Transport: Xray-core v26.6.1 (REALITY + Hysteria 2 + Finalmask + VLESS PQ)
HTTPS: Caddy v2 (Let's Encrypt auto)

## 1. Project Structure



## 2. Data Models (SQLite via GORM)

### User
  ID, UUID (unique), Name, Email (optional), PasswordHash (admin only),
  Role (admin|user), Status (active|disabled), Note, LastSeenAt, LastIP, CreatedAt, UpdatedAt

### Node (Server)
  ID, Name, Address, APIPort, APIKey, Location (russia|germany|...),
  IsActive (toggle), CreatedAt, UpdatedAt

### Inbound (Connection on Node)
  ID, NodeID (FK), Protocol (vless|shadowsocks|trojan|hysteria2),
  Transport (tcp|ws|grpc|h2|quic), Port, Security (reality|tls|none),
  Settings (JSON - fingerprint, serverNames, privateKey...),
  IsActive (toggle), Tag, CreatedAt, UpdatedAt

### Plan (Tariff)
  ID, Name, DurationDays, TrafficLimit (0=unlimited), DeviceLimit,
  NodeLimit, ProfileLimit, Price, IsActive, CreatedAt, UpdatedAt

### Subscription
  ID, UserID (FK), PlanID (FK), Status (active|expired|suspended),
  StartedAt, ExpiresAt, AutoRenew, TrafficUsed, CreatedAt, UpdatedAt

### SubscriptionProfile
  ID, SubscriptionID (FK), Name ("Main"|"For Games"),
  NodeID (FK), InboundIDs (JSON array), UUID, Flow,
  SubscriptionURL, IsActive, CreatedAt, UpdatedAt

### Device
  ID, UserID (FK), Name, DeviceType (android|ios|windows|mac|linux),
  LastConnectedAt, LastIP, IsActive, CreatedAt

### TrafficLog (raw)
  ID, UserID (FK), InboundID (FK), NodeID (FK),
  UploadBytes, DownloadBytes, RecordedAt

### TrafficDaily (aggregated)
  ID, UserID (FK), NodeID (FK), Date, UploadTotal, DownloadTotal

### EventLog (audit)
  ID, UserID (nullable FK), Action, Details (JSON), IPAddress, CreatedAt

### Notification
  ID, UserID (FK), Type (info|warning|error|success), Title, Body, IsRead, CreatedAt

## 3. Relationships (ER)

Node 1---M Inbound
User 1---M Subscription 1---M SubscriptionProfile
User 1---M Device | TrafficLog | Notification
Plan 1---M Subscription
Node 1---M SubscriptionProfile | TrafficLog

## 4. API Endpoints

### Public (no JWT)
  POST /api/auth/login                   - login -> JWT
  GET  /sub/:uuid                        - subscription URL -> client config

### Protected (JWT required)
  GET    /api/dashboard                  - overview stats
  CRUD   /api/users                      - user management
  CRUD   /api/nodes                      - server management
  PATCH  /api/nodes/:id/toggle           - enable/disable node
  CRUD   /api/inbounds                   - inbound management
  PATCH  /api/inbounds/:id/toggle        - enable/disable inbound
  CRUD   /api/plans                      - tariff plans
  CRUD   /api/subscriptions              - subscriptions
  CRUD   /api/profiles                   - client profiles
  PUT    /api/profiles/:id               - edit profile (inbounds, status)
  GET    /api/configs/user/:id           - generate JSON config
  GET    /api/configs/user/:id/qr        - generate QR code

## 5. Frontend Pages

  /login     -> LoginPage       (sign in)
  /          -> DashboardPage   (stats cards + charts)
  /users     -> UsersPage       (admin users table)
  /inbounds  -> InboundsPage    (inbounds table + create)
  /plans     -> PlansPage       (tariff plans)
  /profiles  -> ProfilesPage    (clients + subscriptions)

## 6. Xray Interaction

### Current (config file):
  1. Panel reads /usr/local/etc/xray/config.json
  2. Modifies inbounds / clients
  3. Restarts Xray via systemctl restart xray

### Future (gRPC API - TODO):
  Xray gRPC API on 127.0.0.1:62789
  Services: HandlerService, StatsService, LoggerService
  Benefits: add/remove inbounds WITHOUT restart, get traffic stats,
            active connections, no connection drops

## 7. Infrastructure

### VPS: 45.134.39.18
  Ubuntu 24.04.4, 4 cores, 15 GB RAM, 145 GB disk

### Services:
  xray.service  - Xray-core v26.6.1 (:443 REALITY)
  caddy.service - Caddy v2 (:8443 -> :9090)
  hsvq.service  - HsVq Panel (:9090, Go backend)

### Ports:
  443   - Xray REALITY
  8443  - Caddy HTTPS -> :9090
  9090  - Go backend (internal)
  62789 - Xray gRPC API (localhost only)

## 8. Status

### Done:
  - [x] Go backend (models, migrations, handlers)
  - [x] JWT auth
  - [x] CRUD users, inbounds, plans, nodes
  - [x] Subscription profiles (create client with UUID)
  - [x] Xray config sync (add clients, restart)
  - [x] Embedded frontend (go:embed)
  - [x] Caddy + Let's Encrypt SSL
  - [x] Frontend: all pages
  - [x] Toggle for inbounds/nodes
  - [x] Edit client profile
  - [x] Xray updated to v26.6.1

### In Progress:
  - [ ] gRPC API for Xray (no restart)
  - [ ] Traffic stats via Xray StatsService
  - [ ] Dashboard charts (hour/day/week/month)
  - [ ] REALITY JSON config generation
  - [ ] QR codes for configs

### Planned:
  - [ ] Server stats (CPU, RAM, uptime)
  - [ ] Active connections (online users)
  - [ ] One-time links (temporary access N hours)
  - [ ] Telegram bot (/status, /bind UUID)
  - [ ] Backup export/import
  - [ ] Node Agent (remote node management)
  - [ ] User session history
  - [ ] Email/SMTP notifications
  - [ ] Telegram notifications (traffic limit, expiry)

## 9. REALITY Config

  Port: 443, Protocol: vless, Transport: tcp, Security: reality
  Dest: www.microsoft.com:443, ShortID: 19e72187
  PrivateKey: eJvaNmNJP-X6l2MprNA8KWPGV0aRgG_er6jArLrKZmY
  PublicKey: YI3mGqA_bSX9tIPEB5RL2pPxR8bmC1L_GQqz4v5CxlA
  ServerNames: [microsoft.com, bing.com, cloudflare.com, github.com]

## 10. Dependencies

### Backend:
  gin-gonic/gin, gorm.io/gorm, glebarez/sqlite, golang-jwt/jwt/v5,
  google/uuid, x/crypto/bcrypt

### Frontend:
  react, react-dom, react-router-dom, tailwindcss, vite
