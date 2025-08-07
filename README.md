# Cache - 分布式缓存系统

这是一个基于Go语言实现的分布式缓存系统，支持LRU缓存淘汰策略、HTTP接口访问和一致性哈希节点选择等功能。

## 项目结构

```
.
├── cache
│   ├── byteview.go         // 缓存值抽象（只读字节视图）
│   ├── cache.go            // 并发安全缓存封装
│   ├── consistenthash      // 一致性哈希实现
│   │   ├── consistenthash.go
│   │   └── consistenthash_test.go
│   ├── group.go            // 缓存组核心逻辑
│   ├── group_test.go
│   ├── http.go             // HTTP服务端实现
│   └── lru                 // LRU缓存实现
│       ├── lru.go
│       └── lru_test.go
├── go.mod
└── main.go                 // 主程序入口
```

## 核心功能

### 1. LRU缓存
- 基于双向链表和哈希表实现
- 支持缓存淘汰策略
- 支持淘汰回调函数

### 2. 缓存组(Group)
- 缓存命名空间管理
- 回调函数机制（缓存未命中时从数据源加载）
- 并发安全访问

### 3. HTTP接口
- RESTful风格API
- URL格式：`/cache/<groupname>/<key>`
- 支持GET方法查询缓存值

### 4. 一致性哈希
- 虚拟节点扩展
- 平衡节点负载
- 减少节点变动带来的缓存迁移

## 使用示例

### 启动缓存服务器

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/LudensCS/Cache/cache"
)

var db = map[string]string{
	"jack": "663",
	"Tom":  "78515",
	"lucy": "125",
}

func main() {
	cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
	addr := "localhost:9999"
	peers := cache.NewHTTPPool(addr)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
```

### 访问缓存

通过HTTP请求访问缓存：
```
GET http://localhost:9999/cache/scores/Tom
```

响应：
```
78515
```

## 测试

运行所有测试：
```bash
go test ./...
```

## 依赖

- Go 1.24.4+

## 设计特点

1. **缓存分层设计**：
   - LRU底层缓存
   - 并发安全中间层
   - 缓存组管理层

2. **字节视图抽象**：
   - 只读缓存值保护
   - 字节切片安全克隆

3. **一致性哈希**：
   - 虚拟节点平衡负载
   - 哈希环快速查找

4. **HTTP接口**：
   - 简单RESTful API
   - 标准HTTP协议支持

## 性能考虑

- 读写锁优化并发性能
- LRU高效缓存淘汰
- 字节视图零拷贝优化
- 哈希环O(log n)查找效率

这个缓存系统适合作为中小型应用的分布式缓存解决方案，可扩展性强，易于集成到现有系统中。