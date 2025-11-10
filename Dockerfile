# 多階段構建 - 完整應用程式（Frontend + Backend）

# Stage 1: Build Frontend Assets
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend

# Copy frontend package files
COPY frontend/package*.json ./
RUN npm ci

# Copy frontend source
COPY frontend/ ./

# Build CSS with TailwindCSS
RUN npm run build:css

# Stage 2: Build Backend
FROM golang:1.25.4-alpine AS backend-builder

# 安裝建置所需工具（SQLite 需要 CGO）
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/src/ ./src/
COPY backend/database/ ./database/

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s' \
    -a -installsuffix cgo \
    -o domain-manager ./src/main.go

# Stage 3: Final Runtime Image
FROM alpine:3.19

# 安裝執行時依賴
RUN apk --no-cache add ca-certificates tzdata sqlite-libs wget

# 建立非 root 使用者
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/domain-manager .
COPY --from=backend-builder /app/database ./database

# Copy frontend assets
COPY --from=frontend-builder /frontend/src ./frontend/src
COPY --from=frontend-builder /frontend/dist ./frontend/dist

# Create data directory and set permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# 切換到非 root 使用者
USER appuser

# 暴露服務埠
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/diagnostics/health || exit 1

# 設定環境變數
ENV FRONTEND_PATH=/app/frontend

# 啟動應用程式
CMD ["./domain-manager"]
