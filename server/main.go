package main

import (
	"honeypot/server/pkg/config-load"
	"honeypot/server/pkg/logger"
	"honeypot/server/pkg/proxy"
	"honeypot/server/pkg/service"
	"honeypot/server/pkg/vars"
	"log"
)

func main() {
    // 加载配置
    if err := config.LoadConfig("config/config.yaml"); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 初始化日志
    logger.InitLogger()

    // 启动代理服务
    proxy.StartProxy()

    // 启动各个服务
    for serviceName, serviceConfig := range vars.GlobalConfig.Services {
        go func(name string, cfg vars.ServiceConfig) {
            var err error
            switch name {
            case "ssh":
                err = service.StartSSHService(cfg.ListenAddr, true)
            case "mysql":
                err = service.StartMySQLService(cfg.ListenAddr, true)
            case "redis":
                err = service.StartRedisService(cfg.ListenAddr, true)
            case "web":
                err = service.StartWebService(cfg.ListenAddr, true)
            }
            if err != nil {
                logger.Log.Fatalf("Failed to start %s service: %v", name, err)
            }
        }(serviceName, serviceConfig)
    }

    // 保持程序运行
    select {}
}