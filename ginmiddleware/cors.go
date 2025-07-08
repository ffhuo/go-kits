package ginmiddleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string // 允许的源
	AllowMethods     []string // 允许的方法
	AllowHeaders     []string // 允许的头部
	ExposeHeaders    []string // 暴露的头部
	AllowCredentials bool     // 是否允许凭证
	MaxAge           int      // 预检请求的缓存时间（秒）
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24小时
	}
}

// CORS 跨域中间件
func CORS(config ...*CORSConfig) gin.HandlerFunc {
	var cfg *CORSConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查是否允许该源
		if len(cfg.AllowOrigins) > 0 {
			allowed := false
			for _, allowedOrigin := range cfg.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			if !allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// 设置CORS头部
		if origin != "" {
			if contains(cfg.AllowOrigins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
			} else if contains(cfg.AllowOrigins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		}

		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		}

		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		}

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Header("Access-Control-Max-Age", string(rune(cfg.MaxAge)))
		}

		// 处理预检请求
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
