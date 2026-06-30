# 本地 Docker 数据库

本项目采用“桌面端运行 Go 服务，Docker 只运行项目数据库”的方式，方便和其他项目分离管理。

## 启动

```powershell
docker compose up -d postgres redis
go run .
```

Go 服务默认使用以下配置：

- 后台地址：`http://127.0.0.1:9033/admin`
- PostgreSQL 地址：`127.0.0.1`
- PostgreSQL 端口：`15432`
- 数据库：`kadmin`
- 用户名：`postgres`
- 密码：`kadmin_dev_pwd`
- Redis 地址：`127.0.0.1`
- Redis 端口：`16379`
- Redis 密码：`kadmin_redis_pwd`
- Redis DB：`0`

首次启动时，Docker 会把 `tests/data/admin_pg.sql` 导入到 `kadmin` 数据库中。该导入只会在数据库 volume 为空时执行。

## Navicat 连接 PostgreSQL

先确认 PostgreSQL 容器已启动：

```powershell
docker compose up -d postgres
docker compose ps
```

在 Navicat 中新建 PostgreSQL 连接，填写：

- 主机：`127.0.0.1`
- 端口：`15432`
- 初始数据库：`kadmin`
- 用户名：`postgres`
- 密码：`kadmin_dev_pwd`
- SSL：关闭或默认

这里的 `15432` 是宿主机端口，对应 `docker-compose.yml` 中的 `"${KADMIN_DB_PORT:-15432}:5432"`。如果本机 `15432` 被占用，可以在 `.env` 中改成其他端口，例如：

```dotenv
KADMIN_DB_PORT=25432
```

然后重启 PostgreSQL 容器，并在 Navicat 中使用新端口：

```powershell
docker compose up -d postgres
```

## 可选数据库管理界面

```powershell
docker compose --profile tools up -d adminer
```

打开 `http://127.0.0.1:18080`，然后使用：

- 系统：`PostgreSQL`
- 服务器：`postgres`
- 用户名：`postgres`
- 密码：`kadmin_dev_pwd`
- 数据库：`kadmin`

## 常用命令

```powershell
docker compose ps
docker compose logs -f postgres
docker compose logs -f redis
docker compose exec postgres psql -U postgres -d kadmin
docker compose exec redis redis-cli -a kadmin_redis_pwd ping
go build .
```

## 重置本地数据库

该命令只会删除本项目的 Docker 数据库 volume。

```powershell
docker compose down -v
docker compose up -d postgres redis
```
