// gateway/main.go
package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/task-management/shared/middleware"
	"github.com/task-management/shared/utils"
	"go.uber.org/zap"
)

type Gateway struct {
	services map[string]*url.URL
}

func NewGateway() *Gateway {
	return &Gateway{
		services: map[string]*url.URL{
			"auth":          parseURL("http://localhost:8001"),
			"projects":      parseURL("http://localhost:8002"),
			"tasks":         parseURL("http://localhost:8003"),
			"files":         parseURL("http://localhost:8004"),
			"notifications": parseURL("http://localhost:8005"),
			"reports":       parseURL("http://localhost:8006"),
		},
	}
}

func (g *Gateway) setupRoutes() *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Auth routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.Any("/*path", g.proxyToService("auth"))
	}

	// Protected routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.Any("/projects/*path", g.proxyToService("projects"))
		api.Any("/tasks/*path", g.proxyToService("tasks"))
		api.Any("/files/*path", g.proxyToService("files"))
		api.Any("/reports/*path", g.proxyToService("reports"))
	}

	// WebSocket for notifications (with auth)
	router.GET("/ws", middleware.AuthMiddleware(), g.proxyWebSocket("notifications"))

	return router
}

func (g *Gateway) proxyToService(serviceName string) gin.HandlerFunc {
	target := g.services[serviceName]
	proxy := httputil.NewSingleHostReverseProxy(target)

	return func(c *gin.Context) {
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (g *Gateway) proxyWebSocket(serviceName string) gin.HandlerFunc {
	notificationURL := g.services[serviceName]
	return func(c *gin.Context) {
		// Upgrade the incoming request to a WebSocket
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			utils.GetLogger().Error("WebSocket upgrade failed", zap.Error(err))
			return
		}
		defer clientConn.Close()

		// Connect to the backend WebSocket service
		backendURL := "ws://" + notificationURL.Host + "/ws"
		backendConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
		if err != nil {
			utils.GetLogger().Error("Failed to connect to backend WebSocket", zap.Error(err))
			return
		}
		defer backendConn.Close()

		// Proxy messages between client and backend
		proxy := func(src, dst *websocket.Conn) {
			for {
				mt, message, err := src.ReadMessage()
				if err != nil {
					break
				}
				err = dst.WriteMessage(mt, message)
				if err != nil {
					break
				}
			}
		}
		// Run proxy in both directions
		go proxy(clientConn, backendConn)
		proxy(backendConn, clientConn)
	}
}

func parseURL(rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		utils.GetLogger().Fatal("Invalid service URL", zap.String("url", rawURL), zap.Error(err))
	}
	return url
}

func main() {
	utils.InitLogger()
	logger := utils.GetLogger()
	defer logger.Sync()

	gateway := NewGateway()
	router := gateway.setupRoutes()

	logger.Info("API Gateway starting on port 8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
