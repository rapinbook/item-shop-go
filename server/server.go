package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	playerOauthConfig *oauth2.Config
	adminOauthConfig  *oauth2.Config
)

type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	// Add other relevant user details obtained from Google
}

func init() {
	conf := config.ConfigGetting()
	playerOauthConfig = &oauth2.Config{
		RedirectURL:  conf.OAuth2.PlayerRedirectUrl,
		ClientID:     conf.OAuth2.ClientId,
		ClientSecret: conf.OAuth2.ClientSecret,
		Scopes:       conf.OAuth2.Scopes,
		Endpoint:     google.Endpoint,
	}
	adminOauthConfig = &oauth2.Config{
		RedirectURL:  conf.OAuth2.PlayerRedirectUrl,
		ClientID:     conf.OAuth2.ClientId,
		ClientSecret: conf.OAuth2.ClientSecret,
		Scopes:       conf.OAuth2.Scopes,
		Endpoint:     google.Endpoint,
	}
}

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
	s.app.GET("/v1/oauth2/google/player/login", loginHandler)
	s.app.GET("/v1/oauth2/google/player/login/callback", callbackHandler)
	s.app.GET("/v1/oauth2/google/player/logout", logoutHandler)
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
	time.Sleep(time.Second * 2) // Adjust this timeout based on your task
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

func generateRandomString(length int) string {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func loginHandler(c echo.Context) error {
	state := generateRandomString(32)
	authURL := playerOauthConfig.AuthCodeURL(state)
	c.SetCookie(&http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Expires:  time.Now().AddDate(0, 0, 20),

		// Set other cookie attributes (secure, HttpOnly, etc.)
	})
	return c.Redirect(http.StatusFound, authURL)
}

func callbackHandler(c echo.Context) error {
	state := c.QueryParam("state")
	cookieState, err := c.Request().Cookie("oauthstate")
	if err != nil || state != cookieState.Value {
		return c.String(http.StatusForbidden, "Invalid state parameter")
	}

	code := c.QueryParam("code")
	ctx := context.Background()

	token, err := playerOauthConfig.Exchange(ctx, code)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to exchange code for token")
	}

	// Use the access token to retrieve user information
	client := oauth2.NewClient(ctx, playerOauthConfig.TokenSource(ctx, token))
	userInfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to retrieve user information")
	}

	defer userInfo.Body.Close()
	var user User
	err = json.NewDecoder(userInfo.Body).Decode(&user)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to decode user information")
	}
	c.SetCookie(&http.Cookie{
		Name:    "oauthstate",
		Value:   "true",
		MaxAge:  -1,
		Path:    "/",
		Expires: time.Now(),
		// Set other cookie attributes (secure, HttpOnly, etc.)
	})

	return c.String(http.StatusOK, fmt.Sprintf("Login successful for user: %s (Email: %s)", user.Name, user.Email))
}

func logoutHandler(c echo.Context) error {
	// Clear access token from client-side storage (e.g., localStorage)
	// You'll need to inject JavaScript code to achieve this
	// ... (client-side script injection)

	// Optionally, set a "loggedOut" cookie to indicate logout on server-side

	cookie, _ := c.Cookie("oauthstate")

	cookie = &http.Cookie{
		Name:    "oauthstate",
		Path:    "/",
		Expires: time.Now(),
		Value:   "",
		MaxAge:  -1,
	}
	c.SetCookie(cookie)
	return c.String(http.StatusFound, fmt.Sprintf("Logout successfully %v", cookie.Expires)) // Redirect to login page
}
