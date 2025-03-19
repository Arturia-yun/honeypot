package main

import (
    "honeypot/logServer/pkg/api"
    "honeypot/logServer/pkg/config"
    "honeypot/logServer/pkg/db"
    "honeypot/logServer/pkg/middleware"
    "github.com/gin-gonic/gin"
)

func main() {
    // 初始化配置
    config.Init()

    // 连接数据库
    if err := db.Connect(); err != nil {
        panic(err)
    }
    defer db.Disconnect()

    // 创建 Gin 路由
    r := gin.Default()

    // 创建 API 路由组并应用中间件
    apiGroup := r.Group("/api")
    apiGroup.Use(middleware.APIKeyAuth())
    {
        apiGroup.POST("/packet/", api.HandlePacketLog)
        // 统一添加服务日志路由
        apiGroup.POST("/:service/", api.HandleServiceLog)  // 改为参数化路由
    }

    // 启动服务器
    r.Run(":8083")
}