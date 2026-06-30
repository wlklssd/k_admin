package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"time"

	// ① 引入框架适配器（自注册）
	_ "github.com/GoAdminGroup/go-admin/adapter/gin"
	// ② 引入数据库驱动（PostgreSQL）
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
	// ③ 引入主题（必须！否则 panic: wrong theme name）
	_ "github.com/GoAdminGroup/themes/adminlte"

	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/vbenapi"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	r := gin.New()
	e := engine.Default()
	addr := ":" + getenv("KADMIN_APP_PORT", "9033")

	// ③ 配置
	cfg := config.Config{
		Env: config.EnvLocal,
		Databases: config.DatabaseList{
			"default": {
				Host:            getenv("KADMIN_DB_HOST", "127.0.0.1"),
				Port:            getenv("KADMIN_DB_PORT", "15432"),
				User:            getenv("KADMIN_DB_USER", "postgres"),
				Pwd:             getenv("KADMIN_DB_PASSWORD", "kadmin_dev_pwd"),
				Name:            getenv("KADMIN_DB_NAME", "kadmin"),
				Driver:          config.DriverPostgresql,
				MaxIdleConns:    50,
				MaxOpenConns:    150,
				ConnMaxLifetime: time.Hour,
			},
		},
		UrlPrefix: "admin", // 所有后台路由在 /admin/ 下
		Store:     config.Store{Path: "./uploads", Prefix: "uploads"},
		Language:  language.CN, // 中文
		Debug:     true,
	}

	// ④ 初始化引擎
	if err := e.AddConfig(&cfg).Use(r); err != nil {
		panic(err)
	}

	// Vben 前后端分离接口。原 GoAdmin 服务端页面仍保留在 /admin 下。
	vbenapi.Register(r, e.DefaultConnection())

	// 访问根路径自动跳转到后台
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/admin")
	})

	r.Static("/uploads", "./uploads")

	go func() {
		if err := r.Run(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Print("closing database connection")
	e.PostgresqlConnection().Close()
}

func getenv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
