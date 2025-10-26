# NetworkDev（开发中）

一个基于 Go 语言开发的网络应用管理系统，提供应用程序管理、API接口管理、变量管理、用户认证等功能的 Web 管理平台。

## 功能特性

### 🚀 核心功能
- **应用管理**: 支持应用的增删改查、版本管理、状态控制、密钥管理
- **API接口管理**: 支持多种加密算法的API接口配置（RC4、RSA、易加密等）
- **变量管理**: 独立的变量系统，支持变量的增删改查和别名管理
- **用户管理**: 完整的用户认证和权限管理系统
- **系统设置**: 灵活的系统配置和参数管理
- **仪表盘**: 实时系统状态监控和统计数据展示

### 🔧 技术特性
- **RESTful API**: 标准的 REST API 接口设计
- **JWT 认证**: 基于 JWT 的安全认证机制
- **多种加密算法**: 支持 RC4、RSA、RSA动态、易加密等多种加密方式
- **数据库支持**: 支持 MySQL 和 SQLite 数据库
- **Redis 缓存**: 集成 Redis 缓存提升性能（可选）
- **日志系统**: 完整的日志记录和管理，支持日志切割
- **配置管理**: 基于 Viper 的灵活配置系统

### 🎨 界面特性
- **响应式设计**: 支持多种设备和屏幕尺寸
- **现代化 UI**: 基于 LayUI 的现代化管理界面
- **主题支持**: 支持明暗主题切换
- **实时更新**: 支持数据的实时刷新和更新
- **片段化加载**: 采用 AJAX 片段加载提升用户体验

## 技术栈

- **后端**: Go 1.24.1
- **Web 框架**: Gin + 自定义路由
- **数据库**: GORM + MySQL/SQLite
- **缓存**: Redis（可选）
- **认证**: JWT
- **日志**: Logrus + Lumberjack
- **配置**: Viper
- **前端**: LayUI + JavaScript
- **加密**: 自定义加密工具包

## 项目结构

```
networkDev/
├── cmd/                    # 命令行工具
│   ├── root.go            # 根命令定义
│   └── server.go          # 服务器启动命令
├── config/                 # 配置文件和配置管理
│   ├── config.go          # 配置加载和验证
│   ├── security.go        # 安全配置
│   └── validator.go       # 配置验证器
├── constants/              # 常量定义
│   └── status.go          # 状态常量
├── controllers/            # 控制器层
│   ├── admin/             # 管理后台控制器
│   │   ├── api.go         # API接口管理
│   │   ├── app.go         # 应用管理
│   │   ├── auth.go        # 认证管理
│   │   ├── handlers.go    # 通用处理器
│   │   ├── settings.go    # 系统设置
│   │   ├── user.go        # 用户管理
│   │   └── variable.go    # 变量管理
│   ├── base.go            # 基础控制器
│   └── home/              # 前台控制器
│       └── home.go        # 主页控制器
├── database/              # 数据库相关
│   ├── database.go        # 数据库连接
│   ├── migrate.go         # 数据库迁移
│   └── settings.go        # 默认设置初始化
├── middleware/            # 中间件
│   └── logging.go         # 日志中间件
├── models/                # 数据模型
│   ├── api.go             # API接口模型
│   ├── app.go             # 应用模型
│   ├── settings.go        # 系统设置模型
│   ├── user.go            # 用户模型
│   └── variable.go        # 变量模型
├── server/                # 服务器路由配置
│   ├── admin.go           # 管理后台路由
│   ├── home.go            # 前台路由
│   └── routes.go          # 路由注册
├── services/              # 业务逻辑层
│   ├── query.go           # 查询服务
│   └── settings.go        # 设置服务
├── utils/                 # 工具函数
│   ├── encrypt/           # 加密工具包
│   │   ├── easy.go        # 易加密
│   │   ├── rc4.go         # RC4加密
│   │   ├── rsa.go         # RSA加密
│   │   ├── rsa_dynamic.go # RSA动态加密
│   │   └── rsa_standard.go# RSA标准加密
│   ├── logger/            # 日志工具
│   │   ├── http.go        # HTTP日志
│   │   ├── logger.go      # 日志配置
│   │   └── server.go      # 服务器日志
│   ├── timeutil/          # 时间工具
│   │   └── server.go      # 服务器时间工具
│   ├── cookie.go          # Cookie工具
│   ├── crypto.go          # 加密工具
│   ├── csrf.go            # CSRF防护
│   ├── database.go        # 数据库工具
│   └── errors.go          # 错误处理
└── web/                   # Web 资源
    ├── assets/            # 资源文件
    │   ├── favicon.svg    # 网站图标
    │   ├── logo.svg       # 系统Logo
    │   └── themes.json    # 主题配置
    ├── static/            # 静态资源
    │   ├── css/           # 样式文件
    │   ├── js/            # JavaScript文件
    │   └── lib/           # 第三方库
    ├── template/          # 模板文件
    │   ├── admin/         # 管理后台模板
    │   └── index.html     # 主页模板
    └── public.go          # 静态资源处理
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

3. **配置系统**
   
   项目默认使用 SQLite 数据库，配置文件为 `config.json`。
   
   主要配置项：
   - **数据库配置**: 默认使用 SQLite，也可配置 MySQL
   - **服务器配置**: 默认监听 `0.0.0.0:8080`
   - **Redis配置**: 可选，用于缓存（连接失败时自动禁用）
   - **安全配置**: JWT密钥、加密密钥等

4. **编译项目**
   ```bash
   go build -o networkDev main.go
   ```

5. **运行项目**
   ```bash
   # 直接运行
   ./networkDev server
   
   # 或使用 go run
   go run main.go server
   
   # 指定端口
   ./networkDev server -p 8080
   
   # 指定主机和端口
   ./networkDev server -H 0.0.0.0 -p 8080
   ```

6. **访问系统**
   
   打开浏览器访问: `http://localhost:8080`
   
   默认管理员账号需要通过数据库初始化创建。

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
- `GET /admin/api/auth/captcha` - 获取验证码

### 应用管理接口
- `GET /admin/api/apps/list` - 获取应用列表
- `POST /admin/api/apps/create` - 创建应用
- `POST /admin/api/apps/update` - 更新应用
- `POST /admin/api/apps/delete` - 删除应用
- `POST /admin/api/apps/batch_delete` - 批量删除应用
- `GET /admin/api/apps/get_multi_config` - 获取多开配置
- `POST /admin/api/apps/update_multi_config` - 更新多开配置
- `GET /admin/api/apps/get_bind_config` - 获取绑定配置
- `POST /admin/api/apps/update_bind_config` - 更新绑定配置
- `GET /admin/api/apps/get_register_config` - 获取注册配置
- `POST /admin/api/apps/update_register_config` - 更新注册配置

### API接口管理
- `GET /admin/api/apis/list` - 获取API接口列表
- `POST /admin/api/apis/update` - 更新API接口配置
- `GET /admin/api/apis/apps` - 获取应用列表（用于接口关联）
- `GET /admin/api/apis/types` - 获取API类型列表
- `POST /admin/api/apis/generate_keys` - 生成加密密钥对

### 变量管理接口
- `GET /admin/variable/list` - 获取变量列表
- `POST /admin/variable/create` - 创建变量
- `POST /admin/variable/update` - 更新变量
- `POST /admin/variable/delete` - 删除变量
- `POST /admin/variable/batch_delete` - 批量删除变量

### 用户管理接口
- `GET /admin/api/user/profile` - 获取用户资料
- `POST /admin/api/user/profile/update` - 更新用户资料
- `POST /admin/api/user/password` - 修改密码

### 系统管理接口
- `GET /admin/api/settings` - 获取系统设置
- `POST /admin/api/settings/update` - 更新系统设置
- `GET /admin/api/system/info` - 获取系统信息
- `GET /admin/api/dashboard/stats` - 获取仪表盘统计数据

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

## 开发状态

### 🚧 当前开发进度

#### ✅ 已完成功能
- **基础架构**: 完整的 MVC 架构，支持模块化开发
- **用户认证**: JWT 认证系统，支持登录/登出/验证码
- **应用管理**: 完整的应用 CRUD 操作，支持配置管理
- **API接口管理**: 支持多种加密算法的接口配置
- **变量管理**: 独立的变量系统，支持别名管理
- **系统设置**: 灵活的配置管理系统
- **仪表盘**: 实时系统监控和统计
- **日志系统**: 完整的日志记录和切割功能
- **数据库**: 支持 SQLite 和 MySQL，自动迁移
- **前端界面**: 基于 LayUI 的现代化管理界面

#### 🔄 开发中功能
- **用户权限系统**: 多角色权限管理
- **API 文档**: 自动生成 API 文档
- **数据导入导出**: 支持配置和数据的导入导出
- **监控告警**: 系统状态监控和告警功能
- **插件系统**: 支持第三方插件扩展

#### 📋 计划功能
- **Docker 支持**: 容器化部署
- **集群支持**: 多节点部署和负载均衡
- **WebSocket**: 实时通信功能
- **国际化**: 多语言支持
- **移动端适配**: 响应式设计优化

### 📝 更新日志

#### v0.3.0 (开发中)
- ✅ 重构变量管理系统，移除应用依赖
- ✅ 完善 API 接口管理功能
- ✅ 优化前端用户体验
- ✅ 增强日志系统功能
- 🔄 开发用户权限管理

#### v0.2.0
- ✅ 实现应用管理完整功能
- ✅ 添加 API 接口管理
- ✅ 集成多种加密算法
- ✅ 完善系统设置功能

#### v0.1.0
- ✅ 项目基础架构搭建
- ✅ 用户认证系统
- ✅ 基础管理后台界面
- ✅ 数据库设计和迁移

---

**注意**: 本项目仍在积极开发中，功能和 API 可能会发生变化。建议在生产环境使用前进行充分测试。