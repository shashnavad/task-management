// gateway/main.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/task-management/gateway/middleware"
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

func parseURL(rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal("Invalid service URL:", rawURL)
	}
	return url
}

func main() {
	gateway := NewGateway()
	router := gateway.setupRoutes()

	log.Println("API Gateway starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
