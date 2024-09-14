package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

const MAX_RETURN_AMOUNT = 10

type ItemInfo struct {
	Item_id   int64 `json:"item_id"`
	Item_name string `json:"item_name"`
	Item_amount int32 `json:"item_amount"`
	Item_price int32 `json:"item_price"`
	Item_description string `json:"item_description"`
}


type LItemInfo struct {
	Id   int64 `json:"id"`
	Name string `json:"name"`
}

type RequestParams struct {
	search string
	page     int16
	sorting  string
	sort_asc bool
}

// TODO: fix sql injection vulnerability (search)
func getSQLQuery(params RequestParams) string {
	sorting := false
	additional_params := ""

	if params.search != "" {
		additional_params += "WHERE item_name LIKE %" + params.search + "%"
	}

	switch params.sorting {
	case "price":
		additional_params += " SORT BY price"
		sorting = true
	case "popularity":
		additional_params += " SORT BY times_bought"
		sorting = true
	}

	if sorting && !params.sort_asc {
		additional_params += " DESC"
	}
	return fmt.Sprintf("SELECT item_id, item_name FROM items LIMIT %d OFFSET %d %s", MAX_RETURN_AMOUNT, MAX_RETURN_AMOUNT * params.page, additional_params)
}

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URI"))
	if err != nil {
		panic("Failed to access db")
	}
	defer conn.Close(context.Background())

	e := echo.New()
	e.GET("/items/list", func(c echo.Context) error {
		params := RequestParams{"", 1, "", true};
		page_index, err := strconv.Atoi(c.QueryParam("page"))
		
		if err != nil {
			params.page = 0
		} else {
			params.page = int16(page_index)
		}

		params.search = c.QueryParam("search")
		params.sorting = c.QueryParam("sort")
		params.sort_asc = c.QueryParam("sort_order") == "asc"

		rows, err := conn.Query(context.Background(), getSQLQuery(params))
		fmt.Println(getSQLQuery(params))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		var item_id int64
		var item_name string
		results := make([]LItemInfo, 0, MAX_RETURN_AMOUNT)
		for rows.Next() {
			rows.Scan(&item_id, &item_name)
			results = append(results, LItemInfo{item_id, item_name})
		}

		return c.JSON(http.StatusOK, &results)

	})

	e.GET("/item", func(c echo.Context) error {
		order_id, err := strconv.Atoi(c.QueryParam("id"))
		if err == nil {
			row := conn.QueryRow(context.Background(), "SELECT * FROM items WHERE item_id = " + strconv.Itoa(order_id))
			var info ItemInfo
			row.Scan(&info.Item_id, &info.Item_name, &info.Item_amount, &info.Item_price, &info.Item_description)
			return c.JSON(http.StatusOK, &info)
		}
		return echo.NewHTTPError(http.StatusBadRequest)
	})

	e.Logger.Fatal(e.Start(":8081"))
}
