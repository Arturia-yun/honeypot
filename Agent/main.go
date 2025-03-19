package main

import (
    "log"
    "honeypot/Agent/pkg/capture"
    "honeypot/Agent/pkg/policy"
    "honeypot/Agent/pkg/forward"
    "honeypot/Agent/pkg/logger"
    "honeypot/Agent/pkg/config"
    "honeypot/Agent/pkg/vars"
)

func main() {
    // 加载配置
    if err := config.LoadConfig("config/app.ini"); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 初始化日志
    logger.InitLogger(vars.GlobalConfig.Client.ManagerURL)

    // 加载策略配置
    if err := policy.LoadPolicy("config/policy.yaml"); err != nil {
        log.Fatalf("Failed to load policy: %v", err)
    }

    // 创建并启动数据包捕获
    pc, err := capture.NewPacketCapture(vars.GlobalConfig.Client.Interface)
    if err != nil {
        log.Fatalf("Failed to create packet capture: %v", err)
    }
    defer pc.Stop()

    // 创建并启动转发服务
    fs := forward.NewForwardServer()
    if err := fs.Start(); err != nil {
        log.Fatalf("Failed to start forward server: %v", err)
    }
    defer fs.Stop()

    // 启动数据包捕获
    if err := pc.Start(); err != nil {
        log.Fatalf("Failed to start packet capture: %v", err)
    }

    // 保持程序运行
    select {}
}