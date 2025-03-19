package api

import (
    "context"
    "honeypot/logServer/pkg/db"
    "honeypot/logServer/pkg/models"
    "github.com/gin-gonic/gin"
    "time"
    "fmt"

)

func HandlePacketLog(c *gin.Context) {
    var log models.PacketLog
    if err := c.ShouldBindJSON(&log); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 确保时间戳被设置
    if log.Time.IsZero() {
        log.Time = time.Now()
    }

    // 记录服务类型缺失的情况
    // 设置数据包捕获的服务类型
    if log.Service == "" {
        log.Service = "network_capture"
        // 使用 debug 级别记录，因为这是正常的数据包捕获
        fmt.Printf("Debug: Network packet captured. SrcIP: %s, DstIP: %s, DstPort: %s\n",
            log.SrcIP, log.DstIP, log.DstPort)
    }

    _, err := db.DB.Collection("packet_logs").InsertOne(context.Background(), log)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "success"})
}

func HandleServiceLog(c *gin.Context) {
    serviceType := c.Param("service")
    
    // 使用通用结构解析
    var logData map[string]interface{}
    if err := c.ShouldBindJSON(&logData); err != nil {
        c.JSON(400, gin.H{"error": "Invalid JSON format: " + err.Error()})
        return
    }

    // 添加服务类型标识
    logData["service_type"] = serviceType
    
    // 存储到对应集合
    collectionName := serviceType + "_logs"
    _, err := db.DB.Collection(collectionName).InsertOne(context.Background(), logData)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "success"})
}