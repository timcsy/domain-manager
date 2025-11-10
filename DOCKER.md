# Docker 使用指南

## 快速開始

### 建置並啟動
```bash
docker-compose up --build
```

### 啟動（使用已建置的 image）
```bash
docker-compose up -d
```

### 停止
```bash
docker-compose down
```

### 完全清理（恢復到測試前狀態）
```bash
./docker-cleanup.sh
```

或手動執行：
```bash
docker-compose down -v --rmi all
```

---

## 測試流程

### 1. 建置 Docker Image
```bash
docker-compose build
```

### 2. 啟動容器
```bash
docker-compose up -d
```

### 3. 查看日誌
```bash
docker logs domain-manager -f
```

### 4. 測試 API
```bash
# 登入
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# 取得域名列表
curl http://localhost:8080/api/v1/domains \
  -H "X-Session-Token: YOUR_TOKEN"
```

### 5. 瀏覽器測試
開啟 http://localhost:8080

### 6. 完全清理
```bash
./docker-cleanup.sh
```

---

## 常用指令

### 查看容器狀態
```bash
docker ps | grep domain-manager
```

### 查看容器日誌
```bash
docker logs domain-manager --tail 100 -f
```

### 進入容器
```bash
docker exec -it domain-manager sh
```

### 重啟容器
```bash
docker-compose restart
```

### 查看資源使用
```bash
docker stats domain-manager
```

---

## 清理指令詳解

### 方法 1: 使用清理腳本（推薦）
```bash
./docker-cleanup.sh
```

這會自動執行：
1. 停止並刪除容器
2. 刪除 volumes（資料庫資料）
3. 刪除 Docker images
4. 顯示剩餘資源確認

### 方法 2: Docker Compose 一鍵清理
```bash
docker-compose down -v --rmi all
```

### 方法 3: 手動清理
```bash
# 停止容器
docker stop domain-manager

# 刪除容器
docker rm domain-manager

# 刪除 image
docker rmi domain-manager:latest

# 刪除 volume
docker volume rm domain-manager_domain-manager-data

# 刪除 network
docker network rm domain-manager_domain-manager-network
```

---

## 確認完全清理

執行以下指令確認所有資源已刪除：

```bash
# 檢查容器
docker ps -a | grep domain-manager

# 檢查 images
docker images | grep domain-manager

# 檢查 volumes
docker volume ls | grep domain-manager

# 檢查 networks
docker network ls | grep domain-manager
```

如果所有指令都沒有輸出，表示已完全清理！

---

## 環境變數

可以在 `docker-compose.yml` 中修改環境變數：

```yaml
environment:
  - K8S_MOCK=true          # Mock 模式（不需要真實 K8s）
  - K8S_IN_CLUSTER=false   # 是否在 K8s 叢集內
  - DB_PATH=/app/data/database.db  # 資料庫路徑
  - LOG_LEVEL=info         # 日誌等級
```

---

## 疑難排解

### Port 8080 已被佔用
```bash
# 查看佔用 port 的程式
lsof -i :8080

# 停止佔用的程式
lsof -ti:8080 | xargs kill -9
```

### 容器無法啟動
```bash
# 查看詳細日誌
docker logs domain-manager

# 檢查容器狀態
docker inspect domain-manager
```

### 資料庫損壞
```bash
# 刪除 volume 重新建立
docker-compose down -v
docker-compose up -d
```

---

## 進階使用

### 使用特定版本
```bash
# 建置特定 tag
docker build -t domain-manager:v1.0.0 .

# 啟動特定版本
docker run -p 8080:8080 domain-manager:v1.0.0
```

### 匯出 Image
```bash
docker save domain-manager:latest > domain-manager.tar
```

### 匯入 Image
```bash
docker load < domain-manager.tar
```

### 推送到 Registry
```bash
docker tag domain-manager:latest your-registry/domain-manager:latest
docker push your-registry/domain-manager:latest
```
