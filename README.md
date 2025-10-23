# NetworkDev

一个基于 Go 语言开发的网络设备管理系统，提供应用程序管理、用户认证、卡片管理等功能的 Web 管理平台。

## 功能特性

### 🚀 核心功能
- **应用管理**: 支持应用的增删改查、版本管理、状态控制
- **更新管理**: 支持自动更新、手动下载、强制更新等多种更新方式
- **用户管理**: 完整的用户认证和权限管理系统
- **卡片管理**: 支持多种卡片类型的管理和统计
- **系统设置**: 灵活的系统配置和参数管理

### 🔧 技术特性
- **RESTful API**: 标准的 REST API 接口设计
- **JWT 认证**: 基于 JWT 的安全认证机制
- **数据库支持**: 支持 MySQL 和 SQLite 数据库
- **Redis 缓存**: 集成 Redis 缓存提升性能
- **日志系统**: 完整的日志记录和管理
- **配置管理**: 基于 Viper 的灵活配置系统

### 🎨 界面特性
- **响应式设计**: 支持多种设备和屏幕尺寸
- **现代化 UI**: 基于 LayUI 的现代化管理界面
- **主题支持**: 支持明暗主题切换
- **实时更新**: 支持数据的实时刷新和更新

## 技术栈

- **后端**: Go 1.24.1
- **Web 框架**: 原生 net/http + 自定义路由
- **数据库**: GORM + MySQL/SQLite
- **缓存**: Redis
- **认证**: JWT
- **日志**: Logrus
- **配置**: Viper
- **前端**: LayUI + JavaScript

## 项目结构

```
networkDev/
├── cmd/                    # 命令行工具
├── config/                 # 配置文件和配置管理
├── constants/              # 常量定义
├── controllers/            # 控制器层
│   ├── admin/             # 管理后台控制器
│   └── home/              # 前台控制器
├── database/              # 数据库相关
├── middleware/            # 中间件
├── models/                # 数据模型
├── server/                # 服务器路由配置
├── services/              # 业务逻辑层
├── utils/                 # 工具函数
│   ├── logger/           # 日志工具
│   ├── taskutil/         # 任务工具
│   └── timeutil/         # 时间工具
└── web/                   # Web 资源
    ├── static/           # 静态资源
    └── template/         # 模板文件
```

## 快速开始

### 环境要求

- Go 1.24.1 或更高版本
- MySQL 5.7+ 或 SQLite 3
- Redis (可选，用于缓存)

### 安装步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd networkDev
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置数据库**
   
   复制配置文件并修改数据库连接信息：
   ```bash
   cp config/config.json.example config/config.json
   ```
   
   编辑 `config/config.json` 文件，配置数据库连接参数。

4. **运行项目**
   ```bash
   go run main.go server
   ```

5. **访问系统**
   
   打开浏览器访问: `http://localhost:8080`

### 配置说明

主要配置文件位于 `config/config.json`，包含以下配置项：

- **服务器配置**: 端口、主机地址等
- **数据库配置**: 数据库类型、连接参数等
- **Redis 配置**: Redis 连接参数
- **JWT 配置**: JWT 密钥和过期时间
- **日志配置**: 日志级别和输出方式

### 命令行工具

项目提供了命令行工具支持：

```bash
# 启动服务器
go run main.go server

# 指定端口启动
go run main.go server -p 8080

# 指定主机和端口
go run main.go server -H 0.0.0.0 -p 8080
```

## API 文档

### 认证接口
- `POST /admin/api/auth/login` - 用户登录
- `POST /admin/api/auth/logout` - 用户登出

### 应用管理接口
- `GET /admin/api/apps/list` - 获取应用列表
- `POST /admin/api/apps/create` - 创建应用
- `PUT /admin/api/apps/update` - 更新应用
- `DELETE /admin/api/apps/delete` - 删除应用

### 用户管理接口
- `GET /admin/api/users/list` - 获取用户列表
- `POST /admin/api/users/create` - 创建用户
- `PUT /admin/api/users/update` - 更新用户
- `DELETE /admin/api/users/delete` - 删除用户

## 开发指南

### 代码规范

- 遵循 Go 官方代码规范
- 使用 gofmt 格式化代码
- 添加必要的注释和文档
- 遵循 RESTful API 设计原则

### 数据库迁移

项目使用 GORM 自动迁移功能，启动时会自动创建和更新数据库表结构。

### 日志系统

项目集成了完整的日志系统，支持：
- 不同级别的日志记录
- HTTP 请求日志
- 服务器状态日志
- 自定义日志格式

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t networkdev .

# 运行容器
docker run -d -p 8080:8080 networkdev
```

### 生产环境部署

1. 编译生产版本
   ```bash
   go build -o networkdev main.go
   ```

2. 配置生产环境配置文件

3. 使用进程管理工具（如 systemd）管理服务

## 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 创建 Pull Request

## 更新日志

### v1.0.0
- 初始版本发布
- 基础的应用管理功能
- 用户认证系统
- 管理后台界面

---

**注意**: 本项目仍在积极开发中，功能和 API 可能会发生变化。