package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	

	// "fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var mySigningKey = []byte(os.Getenv("JWT_SIGN_KEY"))

// Initialize Prometheus metrics
var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_auth_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_auht_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

type RequestParams struct {
	Login string `json:"login"`
	Pwd   string `json:"pwd"`
}

type User struct {
	ID       int
	Username string
	Password string
}

func CreateJWT(username string, userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"user_id":  userID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func main() {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URI"))
	if err != nil {
		panic("Failed to access db")
	}
	defer dbpool.Close()

	e := echo.New()

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Middleware to measure request duration
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start).Seconds()
			requestDuration.WithLabelValues(c.Request().Method, c.Path()).Observe(duration)
			return err
		}
	})

	e.POST("/register", func(c echo.Context) (e error) {

		var params RequestParams
		if err := c.Bind(&params); err != nil {
			requestCounter.WithLabelValues("POST", "/register", "400").Inc()
			return c.JSON(http.StatusBadRequest, map[string]string{"status": "error", "error": "Invalid input"})
		}

		var existingUser User
		err := dbpool.QueryRow(context.Background(), "SELECT ad_user_id, ad_login FROM auth_data WHERE ad_login=$1", params.Login).Scan(&existingUser.ID, &existingUser.Username)
		if err == nil {
			requestCounter.WithLabelValues("POST", "/register", "409").Inc()
			return c.JSON(http.StatusConflict, map[string]string{"status": "error", "error": "Login is used"})
		}

		hashedPassword := hashPassword(params.Pwd)
		_, err = dbpool.Exec(context.Background(), "INSERT INTO auth_data (ad_login, ad_pwd_hash) VALUES ($1, $2)", params.Login, hashedPassword)
		if err != nil {
			requestCounter.WithLabelValues("POST", "/register", "500").Inc()
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "Could not register user"})
		}

		var newUserID int
		err = dbpool.QueryRow(context.Background(), "SELECT ad_user_id FROM auth_data WHERE ad_login=$1", params.Login).Scan(&newUserID)
		if err != nil {
			requestCounter.WithLabelValues("POST", "/register", "500").Inc()
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "Could not retrieve user ID"})
		}

		tokenString, err := CreateJWT(params.Login, newUserID)
		if err != nil {
			requestCounter.WithLabelValues("POST", "/register", "500").Inc()
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "Could not create token"})
		}

		c.SetCookie(&http.Cookie{
			Name:  "token",
			Value: tokenString,
			Path:  "/",
		})

		requestCounter.WithLabelValues("POST", "/register", "200").Inc()
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "token": tokenString})
	})

	e.POST("/login", func(c echo.Context) error {
		var params RequestParams
		if err := c.Bind(&params); err != nil {
			requestCounter.WithLabelValues("POST", "/login", "400").Inc()
			return c.JSON(http.StatusBadRequest, map[string]string{"status": "error", "error": "Invalid input"})
		}

		var user User
		err := dbpool.QueryRow(context.Background(), "SELECT ad_user_id, ad_login, ad_pwd_hash FROM auth_data WHERE ad_login=$1", params.Login).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			requestCounter.WithLabelValues("POST", "/login", "401").Inc()
			return c.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "error": "Invalid login credentials"})
		}
		if hashPassword(params.Pwd) != user.Password {
			requestCounter.WithLabelValues("POST", "/login", "401").Inc()
			return c.JSON(http.StatusUnauthorized, map[string]string{"status": "error", "error": "Invalid login credentials"})
		}

		tokenString, err := CreateJWT(user.Username, user.ID)
		if err != nil {
			requestCounter.WithLabelValues("POST", "/login", "500").Inc()
			return c.JSON(http.StatusInternalServerError, map[string]string{"status": "error", "error": "Could not create token"})
		}

		c.SetCookie(&http.Cookie{
			Name:  "token",
			Value: tokenString,
			Path:  "/",
		})

		requestCounter.WithLabelValues("POST", "/login", "200").Inc()
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "token": tokenString})
	})

	e.POST("/logout", func(c echo.Context) error {
		c.SetCookie(&http.Cookie{
			Name:   "token",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		requestCounter.WithLabelValues("POST", "/logout", "200").Inc()
		return c.Redirect(http.StatusSeeOther, "/")
	})

	e.Logger.Fatal(e.Start(":8083"))
}