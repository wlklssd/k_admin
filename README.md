# KAdmin

KAdmin 是一个前后端分离的后台管理项目。后端复用 [GoAdmin](https://github.com/GoAdminGroup/go-admin) 的公共组件，并在 `internal/kadmin` 中提供独立的应用接口；前端基于 [Vben Admin](https://github.com/vbenjs/vue-vben-admin) 的 Vue 3 与 Ant Design Vue 技术栈开发。

> 本项目是独立维护的二次开发项目，不是 GoAdminGroup 或 Vben 团队的官方发行版，也不代表上述项目对 KAdmin 的认可或背书。

## 当前能力

- JWT 登录、刷新令牌、退出登录和 Redis 令牌状态管理
- 用户管理、状态管理、密码重置、头像上传和用户导入
- 部门、角色、菜单及用户授权管理
- 服务端动态菜单和前端权限路由
- 字典类型与字典数据管理
- 系统配置管理
- Cron 定时任务、运行时暂停/恢复、立即执行与执行日志
- 可启停的系统监控、CPU/内存占用、服务器信息与运行时间
- 自动生成的 Swagger API 文档与在线调试界面
- PostgreSQL 数据存储、Redis 鉴权状态存储
- 可选的 MinIO 对象存储和 Adminer 数据库管理工具
- 保留 GoAdmin 原有后台入口，供兼容和公共能力复用

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 后端 | Go、Gin、GoAdmin 公共组件 |
| 前端 | Vue 3、Vben Admin、Ant Design Vue、TypeScript、Vite |
| 数据库 | PostgreSQL |
| 鉴权状态 | Redis |
| 可选存储 | MinIO |
| 本地依赖 | Docker Compose |

## 项目结构

```text
.
├─ admin-web/          # 独立 Vue 前端，主要应用位于 apps/web-antd
├─ internal/kadmin/    # KAdmin 应用后端，统一挂载在 /api
│  ├─ bootstrap/       # 权限和菜单种子定义
│  ├─ modules/files/   # 文件业务的完整 HTTP、服务和仓储链路
│  ├─ modules/jobs/    # 定时任务、内置处理器与执行日志
│  ├─ modules/monitor/ # 可启停的系统指标采样与监控接口
│  ├─ platform/storage/ # 本地与 MinIO 对象存储实现
│  └─ transport/httpx/ # HTTP 响应和中间件基础能力
├─ adapter/            # GoAdmin Web 框架适配器
├─ engine/             # GoAdmin 核心引擎
├─ modules/            # GoAdmin 公共模块
├─ plugins/            # GoAdmin 公共插件
├─ template/           # GoAdmin 原后台模板
├─ tests/data/         # 本地 PostgreSQL 初始化数据
├─ main.go             # KAdmin 后端入口
└─ docker-compose.yml  # PostgreSQL、Redis、MinIO、Adminer
```

项目边界：

- GoAdmin 原后台保留在 `/admin`，不作为新页面的主要开发入口。
- 新的后端接口集中维护在 `internal/kadmin`，默认前缀为 `/api`；新业务优先在 `internal/kadmin/modules` 下按业务域聚合。
- 新的后台页面集中维护在 `admin-web`，沿用 Vben Admin 的组件和工程约定。

## 本地开发

### 环境要求

- Go 1.22.10 或更高版本
- Node.js 22.18+ 或 24+
- pnpm 11+
- Docker 与 Docker Compose

### 1. 启动依赖服务

项目的本地默认值已经与 `docker-compose.yml` 对齐，可以直接启动 PostgreSQL 和 Redis：

```powershell
docker compose up -d postgres redis
```

如需调整端口或账号，可先创建 Docker Compose 使用的 `.env`：

```powershell
Copy-Item .env.example .env
```

后端读取的是进程环境变量；未设置时使用与 `.env.example` 一致的本地开发默认值。生产环境必须替换数据库密码、Redis 密码、`KADMIN_JWT_SECRET` 等敏感配置。

### 2. 启动后端

```powershell
go run .
```

默认地址：

- 原 GoAdmin 后台：`http://127.0.0.1:9033/admin`
- KAdmin API：`http://127.0.0.1:9033/api`
- Swagger 文档：`http://127.0.0.1:9033/swagger/index.html`
- 根地址：`http://127.0.0.1:9033/`，会跳转到 `/admin`

Swagger 默认在调试模式下启用，可通过进程环境变量显式控制：

```powershell
$env:KADMIN_SWAGGER_ENABLED = 'true'
go run .
```

生产环境建议设置 `KADMIN_SWAGGER_ENABLED=false`。修改接口注释后，在项目根目录重新生成文档：

```powershell
go generate .
```

接口注释集中维护在 `internal/kadmin/swagger_operations.go`。生成文件位于 `internal/kadmin/docs`，包含可供 Swagger UI 使用的 Go、JSON 和 YAML 文档。

首次创建 PostgreSQL volume 时，Docker Compose 会自动导入 `tests/data/admin_pg.sql`。

### 3. 启动前端

```powershell
Set-Location admin-web
corepack enable
pnpm install
pnpm dev
```

访问 `http://127.0.0.1:5666`。开发服务器会把 `/api` 请求代理到 `http://127.0.0.1:9033`。

初始化数据中的本地账号为 `admin` / `admin`，仅用于本地开发；首次登录后应立即修改密码。

## 可选服务

启动 MinIO：

```powershell
docker compose --profile storage up -d minio minio-init
```

启动 Adminer：

```powershell
docker compose --profile tools up -d adminer
```

端口、默认账号和数据重置方式见 [DOCKER_LOCAL.md](./DOCKER_LOCAL.md)。

## 构建与检查

后端：

```powershell
go test ./internal/kadmin/...
go build .
```

前端：

```powershell
Set-Location admin-web
pnpm test:unit
pnpm build
```

## 开源来源与许可

本仓库包含来自不同上游项目的代码，使用和再分发时请同时遵守对应许可证：

- 后端基于 GoAdmin 修改并复用其公共组件。GoAdmin 由 GoAdminGroup 及其贡献者维护，采用 [Apache License 2.0](./LICENSE)。本仓库包含对上游代码的修改，相关修改由 KAdmin 项目维护者负责。
- `admin-web` 基于 Vben Admin 修改。Vben Admin 由 Vben 及其贡献者维护，采用 [MIT License](./admin-web/LICENSE)。
- GoAdmin、Vben Admin 及相关名称和标识的权利归各自权利人所有；本项目仅为说明软件来源而引用这些名称。

上游项目：

- GoAdmin：<https://github.com/GoAdminGroup/go-admin>
- Vben Admin：<https://github.com/vbenjs/vue-vben-admin>

请勿删除或替换仓库中的许可证、版权及归属说明。上游项目的问题请反馈到各自社区；KAdmin 修改产生的问题应在本项目中处理。
