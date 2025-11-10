#!/bin/bash

echo "========================================" 
echo "Docker 完全清理腳本"
echo "========================================"
echo ""

echo "🛑 步驟 1: 停止並刪除容器..."
docker-compose down

echo ""
echo "🗑️ 步驟 2: 刪除 volumes（資料庫資料）..."
docker-compose down -v

echo ""
echo "🗑️ 步驟 3: 刪除 Docker images..."
docker rmi domain-manager:latest 2>/dev/null || echo "   Image already removed"

echo ""
echo "📊 步驟 4: 顯示剩餘資源..."
echo "   Containers:"
docker ps -a | grep domain-manager || echo "   ✓ 無 domain-manager 容器"
echo ""
echo "   Images:"
docker images | grep domain-manager || echo "   ✓ 無 domain-manager images"
echo ""
echo "   Volumes:"
docker volume ls | grep domain-manager || echo "   ✓ 無 domain-manager volumes"
echo ""
echo "   Networks:"
docker network ls | grep domain-manager || echo "   ✓ 無 domain-manager networks"

echo ""
echo "========================================" 
echo "✅ 清理完成！系統已恢復到測試前狀態"
echo "========================================"
