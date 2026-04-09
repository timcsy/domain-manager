#!/bin/bash
# MCP Tools 測試腳本
# 用法: ./test_mcp_tools.sh [BASE_URL]

BASE_URL="${1:-http://localhost:8080}"
MCP_URL="$BASE_URL/mcp"
PASS=0
FAIL=0

call_mcp() {
  local id=$1
  local method=$2
  local params=$3
  curl -s -X POST "$MCP_URL" \
    -H "Content-Type: application/json" \
    -d "{\"jsonrpc\":\"2.0\",\"id\":$id,\"method\":\"$method\",\"params\":$params}"
}

check_result() {
  local name=$1
  local response=$2
  if echo "$response" | grep -q '"result"'; then
    echo "  PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $name"
    echo "    Response: $response"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== MCP Tools 測試 ==="
echo "Target: $MCP_URL"
echo ""

# 1. Initialize
echo "--- 初始化 ---"
resp=$(call_mcp 1 "initialize" '{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}')
check_result "initialize" "$resp"

# 2. Ping
echo "--- Ping ---"
resp=$(call_mcp 2 "ping" '{}')
check_result "ping" "$resp"

# 3. Tools list
echo "--- 工具列表 ---"
resp=$(call_mcp 3 "tools/list" '{}')
check_result "tools/list" "$resp"

# 4. Resources list
echo "--- 資源列表 ---"
resp=$(call_mcp 4 "resources/list" '{}')
check_result "resources/list" "$resp"

# 5. List domains
echo "--- 域名工具 ---"
resp=$(call_mcp 10 "tools/call" '{"name":"list_domains","arguments":{}}')
check_result "list_domains" "$resp"

# 6. List services
echo "--- 服務工具 ---"
resp=$(call_mcp 20 "tools/call" '{"name":"list_services","arguments":{}}')
check_result "list_services" "$resp"

# 7. List expiring certificates
echo "--- 憑證工具 ---"
resp=$(call_mcp 30 "tools/call" '{"name":"list_expiring_certificates","arguments":{"days":30}}')
check_result "list_expiring_certificates" "$resp"

# 8. Get diagnostics
echo "--- 診斷工具 ---"
resp=$(call_mcp 40 "tools/call" '{"name":"get_diagnostics","arguments":{"limit":10}}')
check_result "get_diagnostics" "$resp"

# 9. Get system health
resp=$(call_mcp 41 "tools/call" '{"name":"get_system_health","arguments":{}}')
check_result "get_system_health" "$resp"

# 10. Read domain resource
echo "--- 資源讀取 ---"
resp=$(call_mcp 50 "resources/read" '{"uri":"domain://list"}')
check_result "resources/read domain://list" "$resp"

resp=$(call_mcp 51 "resources/read" '{"uri":"service://list"}')
check_result "resources/read service://list" "$resp"

resp=$(call_mcp 52 "resources/read" '{"uri":"certificate://list"}')
check_result "resources/read certificate://list" "$resp"

resp=$(call_mcp 53 "resources/read" '{"uri":"diagnostics://logs"}')
check_result "resources/read diagnostics://logs" "$resp"

# 11. Method not found
echo "--- 錯誤處理 ---"
resp=$(call_mcp 90 "nonexistent/method" '{}')
if echo "$resp" | grep -q '"error"'; then
  echo "  PASS: method not found returns error"
  PASS=$((PASS + 1))
else
  echo "  FAIL: method not found should return error"
  FAIL=$((FAIL + 1))
fi

echo ""
echo "=== 結果 ==="
echo "通過: $PASS"
echo "失敗: $FAIL"
echo "總計: $((PASS + FAIL))"

if [ $FAIL -gt 0 ]; then
  exit 1
fi
