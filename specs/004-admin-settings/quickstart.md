# Quickstart: 管理員設定整合

## 修改密碼

### 透過 UI

1. 登入 domain-manager
2. 進入「系統設定」頁面
3. 在「帳號管理」區塊輸入舊密碼和新密碼
4. 點擊「修改密碼」
5. 自動登出，使用新密碼重新登入

### 透過 API

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/password \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"old_password": "admin", "new_password": "new-secure-password"}'
```

## 修改 Email

```bash
curl -X PATCH http://localhost:8080/api/v1/admin/email \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"email": "devops@example.com"}'
```

## 更新 Cloudflare Token

在已啟用狀態下直接輸入新 token，點擊「更新」即可覆蓋。
