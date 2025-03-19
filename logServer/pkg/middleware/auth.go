package middleware

import (
    "honeypot/logServer/pkg/config"
    "github.com/gin-gonic/gin"
)

func APIKeyAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.GetHeader("X-API-Key")
        if key != config.GlobalConfig.APIKey {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}