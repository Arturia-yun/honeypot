package service

import (
    "honeypot/server/pkg/logger"
    "honeypot/server/pkg/service/ssh"
    "honeypot/server/pkg/service/MySQL"
    "honeypot/server/pkg/service/redisServer"
    "honeypot/server/pkg/service/web"
)

// StartSSHService 启动SSH服务
func StartSSHService(addr string, isProxy bool) error {
    logger.Log.Warningf("start ssh service on %v", addr)
    return ssh.StartSSH(addr, isProxy)
}

// StartMySQLService 启动MySQL服务
func StartMySQLService(addr string, isProxy bool) error {
    logger.Log.Warningf("start mysql service on %v", addr)
    return mysql.StartMySQL(addr, isProxy)
}

// StartRedisService 启动Redis服务
func StartRedisService(addr string, isProxy bool) error {
    logger.Log.Warningf("start redis service on %v", addr)
    return redisServer.StartRedis(addr, isProxy)  
}

// StartWebService 启动Web服务
func StartWebService(addr string, isProxy bool) error {
    logger.Log.Warningf("start web service on %v", addr)
    return web.StartWeb(addr, isProxy)
}