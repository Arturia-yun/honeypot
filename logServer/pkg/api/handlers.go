package api

import (
	"context"
	"encoding/json"
	"fmt"
	"honeypot/logServer/pkg/db"
	"honeypot/logServer/pkg/models"
	"time"

	"github.com/gin-gonic/gin"
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

    // 创建一个新的文档，用于存储到MongoDB
    newDocument := make(map[string]interface{})
    
    // 复制service_type字段
    newDocument["service_type"] = serviceType
    
    // 如果有message字段且是字符串，尝试解析为JSON
    if message, ok := logData["message"]; ok {
        if msgStr, isStr := message.(string); isStr {
            // 打印接收到的消息，用于调试
            fmt.Printf("接收到的消息: %s\n", msgStr)
            
            // 尝试将message解析为JSON对象
            var msgData map[string]interface{}
            if err := json.Unmarshal([]byte(msgStr), &msgData); err == nil {
                // 成功解析，将所有字段添加到新文档
                for k, v := range msgData {
                    newDocument[k] = v
                }
                fmt.Printf("成功解析JSON: %+v\n", msgData)
            } else {
                fmt.Printf("JSON解析错误: %v\n", err)
                // 解析失败，保留原始message字段
                newDocument["message"] = msgStr
                // 添加解析错误信息，方便调试
                newDocument["parse_error"] = err.Error()
            }
        } else {
            // 如果message不是字符串，直接复制
            newDocument["message"] = message
        }
    }
    
    // 复制其他字段
    for k, v := range logData {
        if k != "message" && k != "service_type" {
            newDocument[k] = v
        }
    }
    
    // 存储到对应集合
    collectionName := serviceType + "_logs"
    _, err := db.DB.Collection(collectionName).InsertOne(context.Background(), newDocument)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "success"})
}