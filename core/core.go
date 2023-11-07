package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ffhuo/go-kits/prometheus"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Option func(*option)

type option struct {
	name              string
	debug             bool
	enablePProf       bool
	disableSwagger    bool
	disablePrometheus bool
	enableCors        bool
	log               *zap.Logger
	withoutTracePaths map[string]bool
}

func WithName(name string) Option {
	return func(opt *option) {
		opt.name = name
	}
}

func WithLogger(log *zap.Logger) Option {
	return func(opt *option) {
		opt.log = log
	}
}

func WithDebug(debug bool) Option {
	return func(opt *option) {
		opt.debug = debug
	}
}

// WithDisablePProf 禁用 pprof
func WithPProf(debug bool) Option {
	return func(opt *option) {
		opt.enablePProf = debug
	}
}

// WithDisableSwagger 禁用 swagger
func WithDisableSwagger() Option {
	return func(opt *option) {
		opt.disableSwagger = true
	}
}

// WithDisablePrometheus 禁用prometheus
func WithDisablePrometheus() Option {
	return func(opt *option) {
		opt.disablePrometheus = true
	}
}

// WithEnableCors 设置支持跨域
func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
	}
}

func WithoutTracePaths(paths []string) Option {
	return func(o *option) {
		o.withoutTracePaths = make(map[string]bool, len(paths))
		for _, path := range paths {
			o.withoutTracePaths[path] = true
		}
	}
}

var _ Mux = (*mux)(nil)

// Mux http mux
type Mux interface {
	http.Handler
	Start(port string) error
	Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup
}

type mux struct {
	debug          bool
	disableSwagger bool
	engine         *gin.Engine
	log            *zap.Logger
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.engine.ServeHTTP(w, req)
}

func (m *mux) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return m.engine.Group(relativePath, handlers...)
}

func New(options ...Option) (Mux, error) {
	mux := &mux{
		engine: gin.New(),
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	// withoutTracePaths 这些请求，默认不记录日志
	if opt.withoutTracePaths == nil {
		opt.withoutTracePaths = make(map[string]bool)
	}
	for k, v := range map[string]bool{
		"/metrics": true,

		"/debug/pprof/":             true,
		"/debug/pprof/cmdline":      true,
		"/debug/pprof/profile":      true,
		"/debug/pprof/symbol":       true,
		"/debug/pprof/trace":        true,
		"/debug/pprof/allocs":       true,
		"/debug/pprof/block":        true,
		"/debug/pprof/goroutine":    true,
		"/debug/pprof/heap":         true,
		"/debug/pprof/mutex":        true,
		"/debug/pprof/threadcreate": true,

		"/favicon.ico": true,

		"/healthCheck": true,

		"/docs/index.html": true,
	} {
		opt.withoutTracePaths[k] = v
	}

	mux.debug = opt.debug
	mux.disableSwagger = opt.disableSwagger
	mux.log = opt.log

	if opt.log != nil {
		mux.engine.Use(mux.genRequestID(), mux.Logger(opt.log, opt.withoutTracePaths), gin.Recovery())
	} else {
		mux.engine.Use(mux.genRequestID(), gin.Logger(), gin.Recovery())
	}

	if !opt.debug {
		gin.SetMode(gin.ReleaseMode)
	}

	if opt.enablePProf {
		pprof.Register(mux.engine) // register pprof to gin
	}

	if !opt.disableSwagger {
		mux.engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // register swagger
	}

	if !opt.disablePrometheus {
		p := prometheus.Init(prometheus.Logger(opt.log), prometheus.Handler(mux.engine))
		mux.engine.Use(p.Middleware(opt.withoutTracePaths))
	}

	if opt.enableCors {
		mux.engine.Use(cors.New(cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowHeaders:     []string{"*"},
			AllowCredentials: true,
		}))
	}

	// recover两次，防止处理时发生panic，尤其是在OnPanicNotify中。
	mux.engine.Use(func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Fprintf(gin.DefaultWriter, "core::got panic: %+v\n, stack: %s", err, string(debug.Stack()))
			}
		}()

		ctx.Next()
	})

	return mux, nil
}

func (m *mux) Logger(log *zap.Logger, ignorePaths map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		if _, ok := ignorePaths[path]; ok {
			return
		}
		cost := time.Since(start)
		if cost > time.Minute {
			// Truncate in a golang < 1.8 safe way
			cost = cost - cost%time.Second
		}
		log.Info("gin logger: ",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("cost", cost),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
		// log.Info(c, "status: %d, cost: %v, path: %s, query: %s erors: %s, ip: %s, user-agent: %s",
		// 	c.Writer.Status(), cost, path, query, c.Errors.ByType(gin.ErrorTypePrivate).String(), c.ClientIP(), c.Request.UserAgent())
	}
}

func (m *mux) Start(port string) error {
	server := &http.Server{
		Addr:    port,
		Handler: m,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server start: %v", err))
		}
	}()
	if m.debug {
		if !m.disableSwagger {
			fmt.Fprint(gin.DefaultWriter, "\nstart swagger api: http://localhost"+port+"/docs/index.html\n\n")
		}
		fmt.Fprint(gin.DefaultWriter, "start server at listen"+port+"\n\n")
	}
	return nil
}

func (m *mux) traceOPRecord(withoutTracePaths map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if withoutTracePaths[c.Request.URL.Path] {
			c.Next()
			return
		}
		var body []byte
		if c.Request.Method != http.MethodGet {
			var err error
			body, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				if m.log != nil {
					m.log.Error("core:read body from request error", zap.Error(err))
				} else {
					fmt.Fprintf(gin.DefaultWriter, "core:read body from request error: %v\n", err)
				}
			} else {
				c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			}
		} else {
			body = []byte(c.Request.URL.RequestURI())
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()
	}
}

func (m *mux) genRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.Request.Header.Get("requestId")
		if requestId == "" {
			req := uuid.New()
			requestId = strings.ReplaceAll(req.String(), "-", "")
		}

		c.Set("requestId", requestId)
		c.Set("traceFields", []string{"requestId"})
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
