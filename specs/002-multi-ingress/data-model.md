# Data Model: 多 Ingress Controller 支援

## 既有實體（無變更）

### system_settings

已有相關欄位，無需新增：

| Key | 類型 | 預設值 | 說明 |
|-----|------|--------|------|
| `default_ingress_class` | string | `"nginx"` | 全域預設 Ingress class |
| `ingress_annotations` | JSON string | `"{}"` | 使用者自訂額外 annotation |

## 新增概念實體（程式碼層面，非資料庫）

### IngressProfile

代表一種 Ingress Controller 的預設配置。以 Go map 實作，不儲存在資料庫。

| 欄位 | 類型 | 說明 |
|------|------|------|
| Name | string | Controller 名稱（"nginx", "traefik"） |
| TLSAnnotations | map[string]string | 啟用 TLS 時的預設 annotation |
| DefaultAnnotations | map[string]string | 所有 Ingress 的預設 annotation |

### 預設 Profile 定義

**nginx**:
```
TLSAnnotations:
  nginx.ingress.kubernetes.io/ssl-redirect: "true"
  nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
DefaultAnnotations: (空)
```

**traefik**:
```
TLSAnnotations:
  traefik.ingress.kubernetes.io/router.tls: "true"
  traefik.ingress.kubernetes.io/router.entrypoints: "websecure"
DefaultAnnotations: (空)
```

## IngressConfig 變更

既有 `k8s/ingress.go` 的 `IngressConfig` struct 新增 Annotations 欄位：

| 欄位 | 類型 | 說明 |
|------|------|------|
| Annotations | map[string]string | 合併後的最終 annotation（profile 預設 + 使用者自訂） |

## Annotation 合併順序

1. Controller profile 的預設 annotation
2. Controller profile 的 TLS annotation（如果 SSL 啟用）
3. 使用者在系統設定中自訂的 `ingress_annotations`
4. 後者覆蓋前者（使用者自訂優先）
