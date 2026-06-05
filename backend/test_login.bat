@echo off
curl -s -X POST http://127.0.0.1:9090/api/auth/login -H "Content-Type: application/json" -d "{\"email\":\"admin\",\"password\":\"admin\"}"
