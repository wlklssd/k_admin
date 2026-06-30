# KAdmin Web

独立 Vue 管理端，使用 Vue 3、Vite 和 Ant Design Vue。默认通过 Vite 代理访问后端 `/api`，不会改动 GoAdmin 原生 `/admin` 后台。

## 启动

```powershell
cd admin-web
npm install
npm run dev
```

默认地址：

- Vue 管理端：`http://127.0.0.1:5173`
- Go 后端：`http://127.0.0.1:9033`
- 接口代理：`/api -> http://127.0.0.1:9033/api`

## 构建

```powershell
cd admin-web
npm run build
```

## 说明

- 登录页会优先请求现有 Go 后端 `POST /api/auth/login`。
- 后端未启动时可使用“演示登录”进入本地交互页面。
- 当前页面包含后台布局、搜索表单、权限按钮、数据表格、分页、抽屉表单、弹窗详情、树、穿梭框、上传选择和基础设置表单。
- 当前使用轻量 hash 路由，未知地址会显示 404 页面，例如 `/#/missing`。
