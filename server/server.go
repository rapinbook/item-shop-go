package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/rapinbook/item-shop-go/config"
	"github.com/rapinbook/item-shop-go/databases"
)

type echoServer struct {
	app  *echo.Echo
	db   databases.Database
	conf *config.Config
}

var (
	server *echoServer
	once   sync.Once
)

func NewEchoServer(db databases.Database, conf *config.Config) *echoServer {
	echoApp := echo.New()
	echoApp.Logger.SetLevel(log.DEBUG)

	once.Do(func() {
		server = &echoServer{
			app:  echoApp,
			db:   db,
			conf: conf,
		}
	})

	return server

}

func (s *echoServer) Start() {
	stat := NewStats()
	s.app.Use(stat.Process)
	// Server header
	s.app.Use(ServerHeader)
	// Middleware
	s.app.Use(middleware.Logger())
	// Prevent application from crashing
	s.app.Use(middleware.Recover())
	// CORS
	s.app.Use(getCORSMiddleware(s.conf.Server.AllowOrigins))
	s.app.Use(getTimeOutMiddleware(time.Duration(s.conf.Server.Timeout)))
	s.app.Use(getBodyLimitMiddleware(s.conf.Server.BodyLimit))

	s.app.GET("/v1/health", s.healthCheck)
	s.app.GET("/v1/stats", stat.Handle) // Endpoint to get stats
	// More modern way of shutting down without determine the type of signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go s.httpListening()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println("Waiting for ongoing work to finish...")
	time.Sleep(time.Second * 2) // Adjust this timeout based on your task
	s.app.Logger.Infof("Shutting down service...")
	time.Sleep(time.Second * 8) // Adjust this timeout based on your task
	fmt.Println("Server stopped")
	if err := s.app.Shutdown(ctx); err != nil {
		s.app.Logger.Fatal(err)
	}
}

func (s *echoServer) httpListening() {
	url := fmt.Sprintf(":%d", s.conf.Server.Port)

	if err := s.app.Start(url); err != nil && err != http.ErrServerClosed {
		s.app.Logger.Fatalf("Error: %v", err)
	}
}

func (s *echoServer) healthCheck(pctx echo.Context) error {
	return pctx.String(http.StatusOK, "OK")
}

func getTimeOutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "Error: Request timeout.",
		Timeout:      timeout * time.Second,
	})
}

func getCORSMiddleware(allowOrigins []string) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: allowOrigins,
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	})
}

func getBodyLimitMiddleware(bodyLimit string) echo.MiddlewareFunc {
	return middleware.BodyLimit(bodyLimit)
}
