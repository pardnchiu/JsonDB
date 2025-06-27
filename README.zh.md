# JsonDB - JSON Indexed Cache with Query
> JsonDB 是一個實現的高性能 JSON 數據庫系統，結合了仿 Redis 風格的鍵值操做和仿 MongoDB 風格的文檔查詢功能<br>
> 結合多層記憶體快取系統，提供極致的讀寫性能<br>
> 會先以 Go 開發，驗證完架構後會提供 Rust 版本
>
> 這是 [go-redis-fallback](https://github.com/pardnchiu/go-redis-fallback) 的延伸專案，作一個以 JSON 為主的資料庫<br>
> 使用標準庫套件探討資料庫架設，與後續用於熟悉 Rust 的語法

[![license](https://img.shields.io/github/license/pardnchiu/JsonDB)](LICENSE)
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## 設計目標
- 設計: 多層記憶體快取 + 磁碟持久化
- 風格: 仿 Redis 鍵值操作 + 仿 MongoDB 文檔查詢
- 快取: LRU / Memory 快取架構 + 快取預熱機制
- 儲存: AOF 日誌確保數據安全 + 實體本地三層檔案夾儲存 JSON
- TTL: 惰性刪除 + 自動過期清理 + 快取淘汰策略


### 檔案系統結構
```
go/
├── cmd/
│   ├── cli/main.go          # CLI 客戶端入口
│   └── server/main.go       # 伺服器入口
├── internal/
│   ├── command/             # 指令解析與類型
│   │   ├── parser.go        # 指令解析器
│   │   └── types.go         # 指令類型定義
│   ├── server/              # 伺服器核心
│   │   ├── server.go        # 伺服器主體
│   │   ├── client.go        # 客戶端處理
│   │   ├── clientKV.go      # KV 操作實作
│   │   ├── clientDoc.go     # 文檔操作實作
│   │   └── clientTTL.go     # TTL 操作實作
│   ├── storage/             # 存儲層
│   │   ├── config.go        # 配置與路徑管理
│   │   ├── aofReader.go     # AOF 讀取器
│   │   └── aofWriter.go     # AOF 寫入器
│   └── util/
│       └── util.go          # 工具函數
└── data/                    # 資料存儲目錄
    ├── aof/                 # AOF 日誌檔案
    │   ├── db_0.aof
    │   └── db_1.aof
    └── 0/                   # 資料庫 0 JSON 檔案
        └── 09/8f/6b/        # 三層目錄結構
            └── hash.json
```

### 核心系統
- [x] 多資料庫支援 (0-15)
- [x] 三層目錄結構檔案存儲 (MD5 雜湊分層)
- [x] AOF 持久化機制 (追加式檔案日誌)
- [x] 自動過期清理機制 (每分鐘清理一次)
- [x] 客戶端 CLI 介面
- [x] 支持單次動作指令 `-c "SET <key>"` 
- [ ] LRU 快取機制
- [ ] 快取預熱功能
- [ ] 連線池管理

### KV 操作
- [x] `SELECT <db:int>` - 指定資料庫（0-15）
- [x] `GET <key>` - 取得指定 KEY 的 VALUE
- [x] `SET <key> <value> [ttl_second|expire_time]` - 設定 KV，可選過期時間
- [x] `DEL <key1> [key2] ...` - 刪除一個或多個 KEY
- [x] `EXISTS <key>` - 檢查 KEY 是否存在
- [x] `KEYS <pattern>` - 搜尋符合 [開頭*] 的 KEY
- [x] `TYPE <key>` - 取得 KEY 的資料型別

### DOC 操作
- [x] `FIND <key> <filters> [page:int] [offset:int]` - 查詢符合條件的 DOC `{filters:[]}` 風格，支持分頁查詢結果（已完成 KV 查找）
- [x] `ADD <key> <value>` - 新增 DOC 到 COLLECTION
- [ ] `SORT <key> <filters> <sort_by> [page:int] [offset:int]` - 對查詢結果進行排序，使用 `{sort:[]}` 風格
- [ ] `UPDATE <key> <filters> <set>` - 更新符合條件的 DOC，使用 `{set:{}}` 風格
- [ ] `REMOVE <key> <filters>` - 刪除符合條件的 DOC

### TTL 操作
- [x] `TTL <key> [filters]` - 查看 KEY 的剩餘時間
- [x] `EXPIRE <key> <ttl_second|expire_time> [filters]` - 設定 KEY 的過期時間
- [x] `PERSIST <key> [filters]` - 移除 KEY 的過期設定

### 	其他操作
- [x] `PING` - 連線測試
- [x] `HELP` - 說明資訊
- [ ] 效能測試
- [ ] 錯誤處理
- [ ] 單元測試

### 查詢語法規劃

#### 基本查詢
```bash
# 基本查詢
FIND users {"name":"John"}

# 複合查詢
FIND products {"category":"electronics","price":{"$lt": 1000}}

# 正規表達式查詢
FIND users {"email":{"$regex":".*@gmail.com"}}

# 範圍查詢
FIND orders {"date":{"$gte": "2024-01-01","$lte": "2024-12-31"}}
```

#### 排序與分頁
```bash
# 分頁查詢
FIND users {"status": "active"} 0 10

# 排序查詢
SORT users {"name": "John"} {"age": 1, "name": -1}

# 排序 + 分頁查詢
SORT users {"name": "John"} {"age": 1, "name": -1} 0 10
```

#### 更新資料
```bash
# 單一更新
UPDATE users {"name": "John"} {"$set": {"age": 30}}

# 批量更新
UPDATE products {"category": "electronics"} {"$inc": {"stock": -1}}

# 條件更新
UPDATE orders {"status": "pending"} {"$set": {"status": "processing"}}
```

#### 刪除資料
```bash
# 條件刪除
REMOVE users {"status": "inactive"}

# 批量刪除
REMOVE logs {"date": {"$lt": "2024-01-01"}}
```

## 授權條款

此專案採用 [MIT](LICENSE) 授權條款。

## 作者

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)
