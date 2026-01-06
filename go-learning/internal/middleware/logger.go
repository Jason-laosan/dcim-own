package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
// 记录每个 HTTP 请求的方法、路径、状态码和响应时间
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 计算执行时间
		latency := endTime.Sub(startTime)

		// 获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		// 输出日志
		log.Printf("[%s] %s %s %d %v %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			latency,
			clientIP,
		)
	}
}
