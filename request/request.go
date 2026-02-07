package request

import (
	"context"
	"errors"
	"fmt"
	"loginServer/config"
	"loginServer/src/log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// initCache 初始化缓存，从数据库加载服务器列表
func initCache() {
	servers, err := GetServerList()
	if err != nil {
		log.Error("initCache failed, err:%v", err)
		return
	}
	log.Info("缓存初始化成功，加载服务器数量: %d", len(servers))
}

// createHTTPServer 创建并配置 HTTP 服务器实例（针对登录游戏服务器优化）
func createHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,  // 读取请求体超时（登录请求通常很快）
		ReadHeaderTimeout: 3 * time.Second,  // 读取请求头超时（防止慢速攻击）
		WriteTimeout:      5 * time.Second,  // 写入响应超时
		IdleTimeout:       90 * time.Second, // Keep-Alive连接空闲超时（减少连接建立开销）
		MaxHeaderBytes:    1 << 16,          // 64KB（登录请求头通常很小，1MB过大）
	}
}

// Start 启动 HTTP 服务器
func Start() {
	initCache()
	// 初始化IP白名单（优先从数据库加载，失败则从配置文件加载）
	InitWhitelistFromDB()

	// 获取并设置 Gin 运行模式
	ginmod := config.Config.GetString("gin.mod")
	if !(ginmod == gin.ReleaseMode || ginmod == gin.DebugMode || ginmod == gin.TestMode) {
		ginmod = gin.DebugMode
	}
	gin.SetMode(ginmod)
	req := gin.Default()
	request(req)

	// 1. 获取配置中的 IP
	ip := config.Config.GetString("gin.ip")
	port := config.Config.GetString("gin.port")

	// 2. 判断是否需要获取真实局域网 IP
	if ip == "" || ip == "127.0.0.1" || ip == "localhost" {
		realIP, err := getLocalIPv4()
		if err != nil {
			log.Error("获取本地真实IPv4失败，回退到 127.0.0.1: %v", err)
			ip = "127.0.0.1"
		} else {
			ip = realIP
			log.Info("检测到本地 IP，服务将绑定到: %s", ip)
		}
	}

	// 3. 拼接地址 (注意：这里必须使用变量 ip，而不是再去读 config)
	addr := fmt.Sprintf("%s:%s", ip, port)

	server := createHTTPServer(addr, req)

	// 启动服务器
	serverErr := make(chan error, 1)
	go func() {
		log.Info("HTTP server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// 等待一小段时间确认服务器是否成功启动
	select {
	case err := <-serverErr:
		log.Error("HTTP server failed to start: %v", err)
		os.Exit(1)
	case <-time.After(100 * time.Millisecond):
		log.Info("HTTP server started successfully on %s", addr)
	}

	gracefulExitServer(server)
}

// Stop 停止服务（预留函数，可在关闭时执行清理操作）
func Stop() {
	// 预留清理操作
}

// gracefulExitServer 优雅关闭服务器，监听系统信号并安全关闭
func gracefulExitServer(server *http.Server) {
	// 创建信号通道，监听系统关闭信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	// 等待关闭信号
	sig := <-ch
	log.Info("Received shutdown signal: %v", sig)

	startTime := time.Now()
	Stop()

	// 设置关闭超时时间为 10 秒
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 优雅关闭服务器
	log.Info("Shutting down HTTP server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Error("HTTP server shutdown error: %v", err)
	} else {
		log.Info("HTTP server shutdown successfully, took: %v", time.Since(startTime))
	}
}

// getLocalIPv4 获取本机首个非回环的 IPv4 地址
func getLocalIPv4() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查 ip 地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			// 必须是 IPv4
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("cannot find local IP address")
}
