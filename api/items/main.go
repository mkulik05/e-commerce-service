package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

const MAX_RETURN_AMOUNT = 10

type ItemInfo struct {
	Item_id   int64 `json:"item_id"`
	Item_name string `json:"item_name"`
	Item_amount int32 `json:"item_amount"`
	Item_price int32 `json:"item_price"`
	Item_description string `json:"item_description"`
	ItemBoughts int `json:"item_bought"`
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

func getSQLQuery(params RequestParams, dbpool *pgxpool.Pool) (pgx.Rows, error) {
	sorting := false
	additional_params := ""

	switch params.sorting {
	case "price":
		additional_params += "ORDER BY item_price"
		sorting = true
	case "popularity":
		additional_params += "ORDER BY times_bought"
		sorting = true
	}

	if sorting && !params.sort_asc {
		additional_params += " DESC"
	}

	if params.search != "" {
		return dbpool.Query(context.Background(), "SELECT item_id, item_name FROM items WHERE item_name LIKE $3 "+additional_params+" LIMIT $1 OFFSET $2", MAX_RETURN_AMOUNT, MAX_RETURN_AMOUNT * params.page, "%" + params.search + "%")
	} 
	
	return dbpool.Query(context.Background(), "SELECT item_id, item_name FROM items "+additional_params+" LIMIT $1 OFFSET $2", MAX_RETURN_AMOUNT, MAX_RETURN_AMOUNT * params.page)
}

func main() {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URI"))
	if err != nil {
		panic("Failed to access db")
	}
	defer dbpool.Close()

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

		rows, err := getSQLQuery(params, dbpool)
		if err != nil {
			fmt.Println(err)
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
		item_id, err := strconv.Atoi(c.QueryParam("id"))
		if err == nil {
			row := dbpool.QueryRow(context.Background(), "SELECT * FROM items WHERE item_id=$1", item_id)
			var info ItemInfo
			err := row.Scan(&info.Item_id, &info.Item_name, &info.Item_amount, &info.Item_price, &info.Item_description, &info.ItemBoughts)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest)
			}
			return c.JSON(http.StatusOK, &info)
		}
		return echo.NewHTTPError(http.StatusBadRequest)
	})

	e.Logger.Fatal(e.Start(":8081"))
}
