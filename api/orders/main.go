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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	kafka "github.com/segmentio/kafka-go"
)

const topic = "new-order"
const partition = 0


type Order struct {
	Items_id []int `json:"items_id"`
	Delivery_addr string `json:"delivery_addr"`
}

func getFreeId(tx pgx.Tx) int {
	for {
		n := rand.N(999_999_999)
		row := tx.QueryRow(context.Background(), "SELECT order_id FROM orders WHERE order_id=$1 ", n)
		var id int
		err := row.Scan(&id)		
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
		BatchSize:    1,
    	BatchTimeout: 10 * time.Millisecond,
	}
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

		// TODO: add retrival of user_id from JWT token if present, -1 otherwise
		// TODO: switch to json in order (to support multiple number of one item)

		tx, err := dbpool.Begin(ctx)
		batch := &pgx.Batch{}
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer tx.Rollback(ctx)

		order_id := getFreeId(tx)
		user_id := -1;

		

		_, search_err := tx.Exec(ctx, "INSERT INTO orders (time, order_id, items_id, delivery_addr, user_id) VALUES ($1, $2, $3, $4, $5)", time.Now(), order_id, order.Items_id, order.Delivery_addr, user_id)
		
		if search_err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		
		for _, v := range order.Items_id {
			batch.Queue("UPDATE items SET times_bought = times_bought + 1 WHERE item_id = $1", v)
		}

		br := tx.SendBatch(ctx, batch)
		br.Close()
		if err = tx.Commit(ctx); err != nil {
			fmt.Println(err)
		}


		value, _ := json.Marshal(order)

		kafka_err := writer.WriteMessages(context.Background(), kafka.Message{Value: value}) 

		if kafka_err != nil {
			fmt.Println(kafka_err)
		}
		return c.JSON(http.StatusOK, echo.Map{"order_id": order_id})
		
	})

	e.Logger.Fatal(e.Start(":8082"))
}
