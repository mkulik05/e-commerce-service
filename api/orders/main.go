package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
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
	
	
		
		userID := -1
		if token := c.Request().Header.Get("Authorization"); token != "" {
			claims, err := VerifyJWT(token)
			if err == nil {
				if id, ok := claims["user_id"].(float64); ok {
					userID = int(id)
				}
			}
		}
	
		tx, err := dbpool.Begin(ctx)
		if err != nil {
			fmt.Println(1, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		var orderID int
		err = tx.QueryRow(ctx, "INSERT INTO orders (or_time, or_delivery_addr, or_user_id) VALUES ($1, $2, $3) RETURNING or_id",
			time.Now(), order.DeliveryAddr, userID).Scan(&orderID)
	
		if err != nil {
			fmt.Println(2, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		batch := &pgx.Batch{}
		for k, v := range order.Items {
			fmt.Println(v)
			tx.Exec(ctx, "INSERT INTO m2m_order_items (oi_order_id, oi_item_id, oi_item_amount) VALUES ($1, $2, $3)", orderID, k, v)
			batch.Queue("UPDATE items SET it_times_bought = it_times_bought + $2 WHERE it_id = $1", k, v)
		}



		if err = tx.Commit(ctx); err != nil {
			fmt.Println(4, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	
		tx, err = dbpool.Begin(ctx)
		if err != nil {
			fmt.Println(1, err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer tx.Rollback(ctx);

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			fmt.Println(err)
		}
		if err = tx.Commit(ctx); err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
	
		value, _ := json.Marshal(order)
		kafkaErr := writer.WriteMessages(context.Background(), kafka.Message{Key: []byte(strconv.Itoa(orderID)), Value: value})
		if kafkaErr != nil {
			fmt.Println(kafkaErr)
		}
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "order_id": orderID})
	})

	e.GET("/list", func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return c.JSON(http.StatusForbidden, map[string]string{"status": "error", "error": "Unauthorized"})
		}

		claims, err := VerifyJWT(token)
		if err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"status": "error", "error": "Invalid token"})
		}

		userID := int(claims["user_id"].(float64))
		rows, err := dbpool.Query(context.Background(), "SELECT or_id, or_delivery_addr FROM orders WHERE or_user_id=$1", userID)
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer rows.Close()

		var orders []ReturnOrder
		for rows.Next() {
			var order ReturnOrder
			if err := rows.Scan(&order.OrderID, &order.DeliveryAddr); err != nil {
				fmt.Println(err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}

			items_rows, err := dbpool.Query(context.Background(), "SELECT oi_item_id, oi_item_amount FROM m2m_order_items WHERE oi_order_id=$1", order.OrderID)
			if err != nil {
				fmt.Println(err)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}			
		
			order_items := make(map[int]int)
			for items_rows.Next() {
				var itemId, itemAmount int;
				if err := items_rows.Scan(&itemId, &itemAmount); err != nil {
					fmt.Println(err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				order_items[itemId] = itemAmount
			}
			items_rows.Close()
			order.Items = order_items
			orders = append(orders, order)
		}

		return c.JSON(http.StatusOK, orders)
	})

	e.Logger.Fatal(e.Start(":8082"))
}