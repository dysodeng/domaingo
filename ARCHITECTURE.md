# 架构设计文档

## 分层概览

```
internal/
├── domain/                        # 领域层（最内层，无任何外部依赖）
│   └── user/
│       ├── entity.go              # 领域实体 / 聚合根
│       ├── value_object.go        # 值对象（按需创建）
│       ├── repository.go          # 仓储接口（由此层定义）
│       ├── service.go             # 领域服务（跨实体业务逻辑，按需创建）
│       └── event.go               # 领域事件（跨领域通信，按需创建）
│
├── infrastructure/                # 基础设施层
│   ├── persistence/               # 数据持久化（实现 domain 的仓储接口）
│   │   └── user_repository.go
│   ├── rdb/                       # 关系型数据库客户端（主从封装）
│   ├── gateway/                   # 网关客户端
│   └── registry/                  # 注册中心客户端
│
├── application/                   # 应用层（编排业务流程）
│   └── user/
│       ├── dto.go                 # 请求/响应数据结构
│       └── service.go             # 应用服务
│
├── interfaces/                    # 接口层（处理 HTTP 协议）
│   └── http/
│       └── user/
│           └── handler.go         # 控制器
│
└── modular/                       # 组合根（组装各层，不含业务逻辑）
    ├── module.go                  # Module 接口 + Deps 定义
    └── user/
        └── module.go              # user 领域的组装代码

cmd/
├── app.go                         # 应用生命周期管理
└── modules.go                     # 领域清单（新增领域只改这里）
```

---

## 各层职责

### 领域层 `domain/<领域名>`

系统核心，只表达业务概念，**不依赖任何框架和基础设施**。包含以下几类概念：

#### 实体（Entity）与聚合根（Aggregate Root）

实体是有唯一身份标识的对象，承载业务规则和状态。聚合根是一组实体和值对象的边界，**仓储只操作聚合根，不直接操作内部实体**。

```go
// domain/user/entity.go
type User struct {   // User 是聚合根
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
}
```

> 例如 `Order`（订单）是聚合根，`OrderItem`（订单项）是其内部实体。外部代码只能通过 `Order` 操作 `OrderItem`，不能绕过聚合根直接修改内部状态。这保证了事务边界的一致性。

#### 值对象（Value Object）

没有唯一标识、由属性定义、不可变的概念建模为值对象。相同属性即视为相等。

```go
// domain/order/value_object.go
type Money struct {
    Amount   int64
    Currency string
}
// 没有 ID，不可变，两个 Money{100, "CNY"} 完全相等
```

> 适合建模为值对象的概念：金额、地址、手机号、坐标、颜色等。

#### 仓储接口（Repository）

由领域层定义，声明"我需要什么数据能力"，由基础设施层实现。

```go
// domain/user/repository.go
type Repository interface {
    FindByID(ctx context.Context, id int64) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}
```

> **接口为什么定义在领域层？** 依赖倒置原则：领域层不依赖基础设施，而是由基础设施反过来实现领域定义的接口。换数据库只改 `persistence`，业务代码完全不动。

#### 领域服务（Domain Service）

业务逻辑不总是属于某个实体。当一个操作横跨多个实体、且不适合放在任何单个实体上时，使用领域服务。

```go
// domain/account/service.go
type TransferService struct{}

func (s *TransferService) Transfer(ctx context.Context, from, to *Account, amount Money) error {
    // 跨实体的转账业务规则，不属于任何单个 Account
}
```

#### 领域事件（Domain Event）

领域中发生的重要事情，用于跨领域的解耦通信。`order` 领域不直接调用 `inventory` 领域的代码，而是发布事件。

```go
// domain/order/event.go
type OrderCreated struct {
    OrderID   int64
    UserID    int64
    CreatedAt time.Time
}
```

> 领域事件在业务初期可以暂不实现，待跨领域通信需求出现时再引入，避免过度设计。

---

### 基础设施层 `infrastructure/persistence`

实现领域层定义的仓储接口，是唯一引用 gorm 的地方。

```go
// infrastructure/persistence/user_repository.go

// userModel 是数据库映射模型，与领域实体分离
type userModel struct {
    ID        int64  `gorm:"primaryKey"`
    Name      string
    Email     string
    CreatedAt time.Time
}

type UserRepository struct {
    writer *gorm.DB
    reader *gorm.DB
}

func NewUserRepository(writer, reader *gorm.DB) *UserRepository {
    return &UserRepository{writer: writer, reader: reader}
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*user.User, error) {
    var m userModel
    if err := r.reader.WithContext(ctx).First(&m, id).Error; err != nil {
        return nil, err
    }
    return &user.User{ID: m.ID, Name: m.Name, Email: m.Email}, nil
}
```

> **为什么要有独立的 `userModel`，不直接用领域实体？** 领域实体是业务模型，不应被 gorm tag 污染。表结构和领域模型可以独立演进，互不影响。

> **为什么同时注入 `writer` 和 `reader`？** Repository 内部按操作类型选择连接：查询走从库，写入走主库。调用方无需感知主从细节，由 Repository 在内部处理。

---

### 应用层 `application/<领域名>`

编排业务流程，依赖领域层的接口，**不直接接触数据库和 HTTP**。

- `dto.go`：与外部通信的数据结构，与领域实体分离，避免内部模型泄露
- `service.go`：调用仓储接口和领域服务完成业务用例

```go
// application/user/service.go
type Service struct {
    repo user.Repository  // 依赖接口，不依赖具体实现
}

func NewService(repo user.Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetUser(ctx context.Context, req GetUserRequest) (*GetUserResponse, error) {
    u, err := s.repo.FindByID(ctx, req.ID)
    if err != nil {
        return nil, err
    }
    return &GetUserResponse{ID: u.ID, Name: u.Name, Email: u.Email}, nil
}
```

> **为什么不直接用 `persistence.UserRepository`？** 通过接口依赖，测试时可注入 mock，不需要真实数据库。真实实现只在组合根（modular）注入。

---

### 接口层 `interfaces/http/<领域名>`

只负责 HTTP 协议处理，将请求翻译为应用层调用，将结果序列化为响应。**不包含任何业务逻辑。**

```go
// interfaces/http/user/handler.go
type Handler struct {
    svc *usersvc.Service
}

func NewHandler(svc *usersvc.Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
    mux.HandleFunc("GET /users/{id}", h.GetUser)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    // 解析请求 → 调用应用层 → 序列化响应
}
```

---

### 组合根 `modular/<领域名>`

唯一知道所有层的地方，职责只有一个：把各层连起来。**不含任何业务逻辑。**

```go
// modular/module.go
type Deps struct {
    Rdb    *infraRdb.Rdb
    Logger tilesLogger.Logger
    // 未来新增基础设施（如 Redis）只在此处扩展，所有 Module.Build 签名无需改变
}

type Module interface {
    Build(deps Deps, mux *http.ServeMux) error
}
```

```go
// modular/user/module.go
type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Build(deps modular.Deps, mux *http.ServeMux) error {
    repo    := persistence.NewUserRepository(deps.Rdb.Writer(), deps.Rdb.Reader())
    svc     := usersvc.NewService(repo)
    handler := userhttp.NewHandler(svc)
    handler.Register(mux)
    return nil
}
```

```go
// cmd/modules.go —— 新增领域只改这一个文件
func registerModules() []modular.Module {
    return []modular.Module{
        user.New(),
        // order.New(),
    }
}
```

---

## 依赖方向

```
interfaces/http
      ↓
application              ← 只依赖领域接口
      ↓
domain                   ← 无任何依赖（最稳定）
      ↑
persistence              ← 实现 domain 的接口

modular                  ← 引用以上所有层（仅用于组装，不含逻辑）
      ↑
cmd/modules.go           ← 领域清单
```

**核心原则：依赖只能向内。越靠内的层越稳定，越不依赖具体技术。**

---

## 循环引用规避

组装代码（`modular/<领域名>/module.go`）**不放在 domain 层内**，而是独立存在于 `modular/` 下。

| 包 | 可以引用 | 不能引用 |
|---|---|---|
| `domain/user` | 无 | 其他任何层 |
| `persistence` | `domain/user` | `application`, `interfaces`, `modular` |
| `application/user` | `domain/user` | `persistence`, `interfaces`, `modular` |
| `interfaces/http/user` | `application/user` | `persistence`, `domain`, `modular` |
| `modular/user` | 以上全部 | 无限制（组合根） |

没有任何包引用 `modular`，因此不存在循环。

---

## 新增一个领域的完整步骤

以新增 `order` 领域为例：

1. `internal/domain/order/entity.go` — 定义聚合根和内部实体
2. `internal/domain/order/value_object.go` — 定义值对象（如有）
3. `internal/domain/order/repository.go` — 定义仓储接口
4. `internal/infrastructure/persistence/order_repository.go` — 实现接口
5. `internal/application/order/dto.go` — 定义 DTO
6. `internal/application/order/service.go` — 实现应用服务
7. `internal/interfaces/http/order/handler.go` — 实现控制器
8. `internal/modular/order/module.go` — 组装以上各层
9. `cmd/modules.go` 加一行 `order.New()` — **唯一需要改动的现有文件**

---

## 已知局限与后续演进

### 跨仓储事务

当一个应用服务需要同时操作多个 Repository 且要求原子性时（如下单同时扣库存），当前设计无法直接支持。

**后续方案：** 引入 Unit of Work 模式，由应用层统一管理事务边界，将同一个 `*gorm.DB`（已开启事务）传入多个 Repository。

### 跨领域通信

`order` 领域不应该直接 import `inventory` 领域的代码，否则领域间形成强耦合。

**后续方案：** 引入领域事件机制，`order` 发布 `OrderCreated` 事件，`inventory` 订阅处理库存扣减，两个领域通过事件总线解耦。

### 以上两点均无需现在实现

在业务真正出现对应复杂度之前引入这些模式属于过度设计。当前架构已足够支撑业务初期的开发，遇到具体场景时再针对性演进即可。
