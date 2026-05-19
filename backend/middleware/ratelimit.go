package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// ipLimiter 存储每个 IP 的限流器
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*ipLimiter)
	mu       sync.RWMutex
)

// getVisitor 获取或创建 IP 对应的限流器
func getVisitor(ip string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(r, b)
		visitors[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// RateLimit 速率限制中间件
// r: 每秒请求数，b: 突发容量
func RateLimit(r float64, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := getVisitor(c.ClientIP(), rate.Limit(r), b)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimit 登录接口限流：20次/分钟
func LoginRateLimit() gin.HandlerFunc {
	return RateLimit(20.0/60.0, 10) // 每秒 20/60 个请求，突发 10
}

// RegisterRateLimit 注册接口限流：10次/分钟
func RegisterRateLimit() gin.HandlerFunc {
	return RateLimit(10.0/60.0, 5) // 每秒 10/60 个请求，突发 5
}

// SearchRateLimit 搜索接口限流：60次/分钟
func SearchRateLimit() gin.HandlerFunc {
	return RateLimit(60.0/60.0, 20) // 每秒 1 个请求，突发 20
}

// CleanupVisitors 定期清理过期的限流器条目
func CleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
