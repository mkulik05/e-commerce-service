package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"github.com/golang-jwt/jwt/v5"
)

const topic = "new-order"
// const partition = 0

type Order struct {
	Items       map[int]int `json:"items"`
	DeliveryAddr  string `json:"delivery_addr"`
}

type ReturnOrder struct {
	OrderID      int   `json:"order_id"`
	Items      map[int]int `json:"items"`
	DeliveryAddr string `json:"delivery_addr"`
}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:       kafka.TCP(kafkaURL),
		Topic:      topic,
		Balancer:   &kafka.LeastBytes{},
		BatchSize:  1,
		BatchTimeout: 10 * time.Millisecond,
	}
}

func VerifyJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SIGN_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func main() {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URI"))
	if err != nil {
		panic("Failed to access db")
	}
	defer dbpool.Close()

	writer := newKafkaWriter(os.Getenv("KAFKA_URL"), topic)
	defer writer.Close()

	e := echo.New()

	e.POST("/order", func(c echo.Context) error {
		ctx := context.Background()
		order := new(Order)
		if err := c.Bind(order); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
	
		tx, err := dbpool.Begin(ctx)
		if err != nil {
			fmt.Println(1, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer tx.Rollback(ctx)
	
		
		userID := -1
		fmt.Println("lalala")
		if token := c.Request().Header.Get("Authorization"); token != "" {
			claims, err := VerifyJWT(token)
			fmt.Println("========", err)
			if err == nil {
				if id, ok := claims["user_id"].(float64); ok {
					userID = int(id)
				}
			}
		}
	
		// before inserting order into db convert it to JSON
		// so as a key there will be item id, as value - item amount
		items, _ := json.Marshal(order.Items)
		var orderID int
		err = tx.QueryRow(ctx, "INSERT INTO orders (time, items, delivery_addr, user_id) VALUES ($1, $2, $3, $4) RETURNING order_id",
			time.Now(), items, order.DeliveryAddr, userID).Scan(&orderID)
	
		if err != nil {
			fmt.Println(2, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	
		batch := &pgx.Batch{}
		
		for _, v := range order.Items {
			batch.Queue("UPDATE items SET times_bought = times_bought + 1 WHERE item_id = $1", v)
		}
	
		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			fmt.Println(err)
		}
		if err = tx.Commit(ctx); err != nil {
			fmt.Println(err)
		}
	
		value, _ := json.Marshal(order)
		kafkaErr := writer.WriteMessages(context.Background(), kafka.Message{Value: value})
		if kafkaErr != nil {
			fmt.Println(kafkaErr)
		}
		return c.JSON(http.StatusOK, echo.Map{"order_id": orderID})
	})

	e.GET("/list", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
		}

		claims, err := VerifyJWT(token)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid token"})
		}

		userID := int(claims["user_id"].(float64))
		rows, err := dbpool.Query(context.Background(), "SELECT order_id, items, delivery_addr FROM orders WHERE user_id=$1", userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer rows.Close()

		var orders []ReturnOrder
		for rows.Next() {
			var order ReturnOrder
			var items_str string;
			if err := rows.Scan(&order.OrderID, &items_str, &order.DeliveryAddr); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
			json.Unmarshal([]byte(items_str), &order.Items)
			orders = append(orders, order)
		}

		return c.JSON(http.StatusOK, orders)
	})

	e.Logger.Fatal(e.Start(":8082"))
}