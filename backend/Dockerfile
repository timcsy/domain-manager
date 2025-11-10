# Build stage
FROM golang:1.25.4-alpine AS builder

# 安裝建置所需工具（SQLite 需要 CGO）
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./src/
COPY database/ ./database/

# Build the application with CGO enabled for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags='-w -s' \
    -a -installsuffix cgo \
    -o domain-manager ./src/main.go

# Final stage
FROM alpine:3.19

# 安裝執行時依賴
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

# 建立非 root 使用者
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/domain-manager .
COPY --from=builder /app/database ./database

# Create data directory and set permissions
RUN mkdir -p /app/data && \
    chown -R appuser:appuser /app

# 切換到非 root 使用者
USER appuser

EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/diagnostics/health || exit 1

CMD ["./domain-manager"]
