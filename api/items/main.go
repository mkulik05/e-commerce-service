package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const MAX_RETURN_AMOUNT = 10

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_items_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_items_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

type ItemInfo struct {
	Item_id          int64  `json:"item_id"`
	Item_name        string `json:"item_name"`
	Item_amount      int32  `json:"item_amount"`
	Item_price       int32  `json:"item_price"`
	Item_description string `json:"item_description"`
	ItemBoughts      int    `json:"item_bought"`
}

type LItemInfo struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

type RequestParams struct {
	search   string
	page     int16
	sorting  string
	sort_asc bool
}

func getSQLQuery(params RequestParams, dbpool *pgxpool.Pool) (pgx.Rows, error) {
	sorting := false
	additional_params := ""

	switch params.sorting {
	case "price":
		additional_params += "ORDER BY it_price"
		sorting = true
	case "popularity":
		additional_params += "ORDER BY it_times_bought"
		sorting = true
	}

	if sorting && !params.sort_asc {
		additional_params += " DESC"
	}

	if params.search != "" {
		return dbpool.Query(context.Background(), "SELECT it_id, it_name, it_price FROM items WHERE it_name LIKE $2 "+additional_params+" OFFSET $1 LIMIT $3", MAX_RETURN_AMOUNT*params.page, "%"+params.search+"%", MAX_RETURN_AMOUNT)
	}

	return dbpool.Query(context.Background(), "SELECT it_id, it_name, it_price FROM items "+additional_params+" OFFSET $1 LIMIT $2", MAX_RETURN_AMOUNT*params.page, MAX_RETURN_AMOUNT)
}

func getSQLQueryRecsAmount(params RequestParams, dbpool *pgxpool.Pool) (int, error) {
	var row pgx.Row
	if params.search != "" {
		row = dbpool.QueryRow(context.Background(), "SELECT COUNT(*) FROM items WHERE it_name LIKE $1", "%"+params.search+"%")
	} else {
		row = dbpool.QueryRow(context.Background(), "SELECT COUNT(*) FROM items")
	}

	var n int
	err := row.Scan(&n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func prometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		duration := time.Since(start).Seconds()

		status := c.Response().Status
		method := c.Request().Method
		endpoint := c.Request().URL.Path // Only use the path, excluding query params

		requestCounter.WithLabelValues(method, endpoint, strconv.Itoa(status)).Inc()
		requestDuration.WithLabelValues(method, endpoint).Observe(duration)

		return err
	}
}

func main() {
	// Register metrics with Prometheus
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URI"))
	if err != nil {
		panic("Failed to access db")
	}
	defer dbpool.Close()

	e := echo.New()
	e.Use(prometheusMiddleware)

	e.GET("/list", func(c echo.Context) error {
		params := RequestParams{"", 1, "", true}
		page_index, err := strconv.Atoi(c.QueryParam("page"))

		if err != nil {
			params.page = 0
		} else {
			params.page = int16(page_index)
		}

		params.search = c.QueryParam("search")
		params.sorting = c.QueryParam("sort")
		params.sort_asc = c.QueryParam("sort_order") == "asc"

		items_n, err := getSQLQueryRecsAmount(params, dbpool)
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		rows, err := getSQLQuery(params, dbpool)
		if err != nil {
			fmt.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var item_id int64
		var item_name string
		var item_price int
		results := make([]LItemInfo, 0, MAX_RETURN_AMOUNT)

		for rows.Next() {
			rows.Scan(&item_id, &item_name, &item_price)
			results = append(results, LItemInfo{item_id, item_name, item_price})
		}

		res := items_n / MAX_RETURN_AMOUNT
		if items_n%MAX_RETURN_AMOUNT != 0 {
			res++
		}
		response := map[string]interface{}{
			"amount": res,
			"items":  results,
		}

		return c.JSON(http.StatusOK, response)
	})

	e.GET("/item", func(c echo.Context) error {
		item_id, err := strconv.Atoi(c.QueryParam("id"))
		if err == nil {
			row := dbpool.QueryRow(context.Background(), "SELECT (it_id, it_name, it_amount, it_price, it_item_desc, it_times_bought) FROM items WHERE it_id=$1", item_id)
			var info ItemInfo
			err := row.Scan(&info.Item_id, &info.Item_name, &info.Item_amount, &info.Item_price, &info.Item_description, &info.ItemBoughts)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest)
			}
			return c.JSON(http.StatusOK, &info)
		}
		return echo.NewHTTPError(http.StatusBadRequest)
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Logger.Fatal(e.Start(":8081"))
}