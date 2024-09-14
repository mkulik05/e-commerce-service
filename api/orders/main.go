package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	kafka "github.com/segmentio/kafka-go"
)

const topic = "new-order"
const partition = 0


type Order struct {
	Items_id []int `json:"items_id"`
	Delivery_addr string `json:"delivery_addr"`
}

func getFreeId(conn *pgx.Conn) int {
	for {
		n := rand.N(999_999_999)
		row := conn.QueryRow(context.Background(), "SELECT order_id FROM orders WHERE order_id=$1 ", n)
		var id int
		err := row.Scan(&id)
		fmt.Println(n, id, err)
		
		if err == pgx.ErrNoRows {
			return n
		}
	}
}

func newKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}


func main() {

	e := echo.New()
	e.POST("/order", func(c echo.Context) error {
		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URI"))
		if err != nil {
			panic("Failed to access db")
		}
		defer conn.Close(context.Background())
		fmt.Println(1)
		writer := newKafkaWriter("192.168.0.109:9094", topic)
		defer writer.Close()

		fmt.Println(2)
		order := new(Order)
		if err := c.Bind(order); err != nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
		fmt.Println(3)
		// TODO: add retrival of user_id from JWT token if present, -1 otherwise
		order_id := getFreeId(conn)
		user_id := -1;
		fmt.Println(33)
		_, search_err := conn.Query(context.Background(), "INSERT INTO orders (time, order_id, items_id, delivery_addr, user_id) VALUES ($1, $2, $3, $4, $5)", time.Now(), order_id, order.Items_id, order.Delivery_addr, user_id)
		fmt.Println(4)
		if search_err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		value, _ := json.Marshal(order)
		fmt.Println(4)
		kafka_err := writer.WriteMessages(context.Background(), kafka.Message{Value: value}) 
		fmt.Println(5)
		if kafka_err != nil {
			fmt.Println(kafka_err)
		}
		return c.JSON(http.StatusOK, echo.Map{"order_id": order_id})
		
	})

	e.Logger.Fatal(e.Start(":8082"))
}
