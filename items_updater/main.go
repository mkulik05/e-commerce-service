	package main

	import (
		"context"
		"encoding/json"
		"fmt"
		"log"
		"os"

		"github.com/jackc/pgx/v5/pgxpool"
		"github.com/segmentio/kafka-go"
	)

	type ItemInfo struct {
		ItemID          int64  `json:"item_id"`
		ItemName        string `json:"item_name"`
		ItemAmount      int32  `json:"item_amount"`
		ItemPrice       int64  `json:"item_price"`
		ItemDescription string `json:"item_description"`
		TimesBought     int    `json:"times_bought"`
	}

	type ItemMessage struct {
		Action string   `json:"action"` // "add", "modify", "delete"
		Item   ItemInfo `json:"item"`
	}

	func handleMessage(dbpool *pgxpool.Pool, msg ItemMessage) error {
		switch msg.Action {
		case "add":
			_, err := dbpool.Exec(context.Background(), "INSERT INTO items (it_name, it_amount, it_price, it_desc, it_times_bought) VALUES ($1, $2, $3, $4, $5)",
				msg.Item.ItemName, msg.Item.ItemAmount, msg.Item.ItemPrice, msg.Item.ItemDescription, msg.Item.TimesBought)
			return err
		case "modify":
			_, err := dbpool.Exec(context.Background(), "UPDATE items SET it_name=$1, it_amount=$2, it_price=$3, it_desc=$4, it_times_bought=$5 WHERE item_id=$6",
				msg.Item.ItemName, msg.Item.ItemAmount, msg.Item.ItemPrice, msg.Item.ItemDescription, msg.Item.TimesBought, msg.Item.ItemID)
			return err
		case "delete":
			_, err := dbpool.Exec(context.Background(), "DELETE FROM items WHERE it_id=$1", msg.Item.ItemID)
			return err
		default:
			return fmt.Errorf("unknown action: %s", msg.Action)
		}
	}

	func main() {

		dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URI"))
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer dbpool.Close()


		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{os.Getenv("KAFKA_URL")},
			Topic:     "items-updates",
			Partition: 0,
			MaxBytes:  10e6, 
		})

		defer r.Close()

		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message: %v", err)
				break
			}

			var itemMsg ItemMessage
			if err := json.Unmarshal(m.Value, &itemMsg); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}

			if err := handleMessage(dbpool, itemMsg); err != nil {
				log.Printf("Error handling message: %v", err)
			} else {
				log.Printf("Successfully processed message: %s", string(m.Value))
			}
		}
	}