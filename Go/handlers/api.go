package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

type API struct {
	cache *redis.Client
	db    *sql.DB
}

type APIResponse struct {
	Cache bool         `json:"cache"`
	Data  ExchangeRate `json:"data"`
}

type ExchangeRate struct {
	Code      string  `json:"code"`
	BasePrice float64 `json:"basePrice"`
}

func handleError(err error, w http.ResponseWriter, msg string, code int) {
	if err != nil {
		http.Error(w, msg, code)
	}
}

func (a *API) DB(w http.ResponseWriter, r *http.Request) {
	a.handleDB(w, r)
}

func (a *API) OpenAPI(w http.ResponseWriter, r *http.Request) {
	a.handleAPI(w, r)
}

func (a *API) fromCache(val string, w http.ResponseWriter, r *http.Request) {
	// cache hit
	var exchangeRate []ExchangeRate
	if err := json.Unmarshal([]byte(val), &exchangeRate); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(exchangeRate) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Println("cache hit")

	result := APIResponse{
		Cache: true,
		Data:  exchangeRate[0],
	}

	jerr := json.NewEncoder(w).Encode(result)
	if jerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (a *API) handleDB(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	code := r.URL.Query().Get("codes")

	if len(code) == 0 {
		http.Error(w, "bad request", http.StatusInternalServerError)
		return
	}

	val, err := a.cache.Get(r.Context(), code).Result()
	// cache miss
	if err != nil {
		log.Println("cache miss - fetch from postgres")
		rows, err := a.db.Query("SELECT * FROM exchange_rate WHERE id = 1")
		handleError(err, w, "database error", http.StatusInternalServerError)

		defer rows.Close()

		var id int
		var exchangeRate ExchangeRate

		rows.Next()
		err = rows.Scan(&id, &exchangeRate.Code, &exchangeRate.BasePrice)
		handleError(err, w, "database error", http.StatusInternalServerError)

		value, jerr := json.Marshal(exchangeRate)
		if jerr != nil {
			handleError(err, w, "json marshal error", http.StatusInternalServerError)
		}

		if rerr := a.cache.Set(r.Context(), code, string(value), time.Second*15).Err(); rerr != nil {
			handleError(rerr, w, "set cache error", http.StatusInternalServerError)
		}

		result := APIResponse{
			Cache: false,
			Data:  exchangeRate,
		}

		rerr := json.NewEncoder(w).Encode(result)
		handleError(rerr, w, "json encode error", http.StatusInternalServerError)
	} else {
		log.Println("cache hit")
		var exchangeRate ExchangeRate
		if err := json.Unmarshal([]byte(val), &exchangeRate); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		result := APIResponse{
			Cache: true,
			Data:  exchangeRate,
		}

		jerr := json.NewEncoder(w).Encode(result)
		if jerr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	log.Println(time.Since(start))
}

func (a *API) handleAPI(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	code := r.URL.Query().Get("codes")

	if len(code) == 0 {
		http.Error(w, "bad request", http.StatusInternalServerError)
	}

	val, err := a.cache.Get(r.Context(), code).Result()
	// cache miss
	if err != nil {
		log.Println("cache miss - fetch from api")
		data, gerr := a.getData(code, r)
		handleError(gerr, w, "bad request", http.StatusInternalServerError)

		result := APIResponse{
			Cache: false,
			Data:  data,
		}

		rerr := json.NewEncoder(w).Encode(result)
		handleError(rerr, w, "json encode error", http.StatusInternalServerError)
	} else {
		a.fromCache(val, w, r)
	}

	log.Println(time.Since(start))
}

func (a *API) getData(code string, r *http.Request) (ExchangeRate, error) {
	var result ExchangeRate

	url := "https://quotation-api-cdn.dunamu.com/v1/forex/recent?codes=FRX." + code
	res, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	if rerr := a.cache.Set(r.Context(), code, string(body), time.Second*15).Err(); rerr != nil {
		return result, err
	}

	var exchangeRate []ExchangeRate
	if err := json.Unmarshal(body, &exchangeRate); err != nil {
		return result, err
	}

	if err != nil {
		return result, nil
	}

	return exchangeRate[0], nil
}

func NewAPI() *API {
	return &API{
		cache: newRedis(),
		db:    newPostgres(),
	}
}

func newRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL") + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}

func newPostgres() *sql.DB {

	// postgres 컨테이너 연결 대기
	time.Sleep(time.Second * 5)

	var (
		HOST     = os.Getenv("POSTGRES_URL")
		PORT     = 5432
		DATABASE = "root"
		USER     = "root"
		PASSWORD = "test"
	)
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", HOST, PORT, USER, PASSWORD, DATABASE)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully created connection to database")

	return db
}
