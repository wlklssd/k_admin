# 本地 Docker 数据库

本项目采用“桌面端运行 Go 服务，Docker 只运行项目数据库”的方式，方便和其他项目分离管理。

## 启动

```powershell
docker compose up -d postgres
go run .
```

Go 服务默认使用以下配置：

- 后台地址：`http://127.0.0.1:9033/admin`
- PostgreSQL 地址：`127.0.0.1`
- PostgreSQL 端口：`15432`
- 数据库：`pezmax`
- 用户名：`postgres`
- 密码：`pezmax_dev_pwd`

首次启动时，Docker 会把 `tests/data/admin_pg.sql` 导入到 `pezmax` 数据库中。该导入只会在数据库 volume 为空时执行。

## 可选数据库管理界面

```powershell
docker compose --profile tools up -d adminer
```

打开 `http://127.0.0.1:18080`，然后使用：

- 系统：`PostgreSQL`
- 服务器：`postgres`
- 用户名：`postgres`
- 密码：`pezmax_dev_pwd`
- 数据库：`pezmax`

## 常用命令

```powershell
docker compose ps
docker compose logs -f postgres
docker compose exec postgres psql -U postgres -d pezmax
go build .
```

## 重置本地数据库

该命令只会删除本项目的 Docker 数据库 volume。

```powershell
docker compose down -v
docker compose up -d postgres
```
