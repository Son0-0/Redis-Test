package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/go-redis/redis/v8"
)

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	a.handle(w, r)
}

func (a *API) handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	code := r.URL.Query().Get("codes")
	log.Println("I'm Handler code is: ", code, reflect.TypeOf(code))

	val, err := a.cache.Get(r.Context(), code).Result()
	// cache fault
	if err != nil {
		log.Println("cache fault")
		data, gerr := a.getData(code, r)
		if gerr != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
		}

		result := APIResponse{
			Cache: false,
			Data:  data,
		}

		rerr := json.NewEncoder(w).Encode(result)
		if rerr != nil {
			fmt.Printf("error encoding response: %v\n", rerr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else { // cache hit
		log.Println("cache hit")
		var exchangeRate []ExchangeRate
		if err := json.Unmarshal([]byte(val), &exchangeRate); err != nil {
			log.Println("ERROR:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		result := APIResponse{
			Cache: true,
			Data:  exchangeRate[0],
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			log.Printf("error encoding response: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
		log.Println("ERROR:", err)
		return result, err
	}

	if err != nil {
		log.Println("Set Cache Error")
		return result, nil
	}

	return exchangeRate[0], nil
}

func NewAPI() *API {
	return &API{
		cache: newRedis(),
	}
}

func newRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return rdb
}

type API struct {
	cache *redis.Client
}

type APIResponse struct {
	Cache bool         `json:"cache"`
	Data  ExchangeRate `json:"data"`
}

type ExchangeRate struct {
	Code              string  `json:"code"`
	CurrencyCode      string  `json:"currencyCode"`
	CurrencyName      string  `json:"currencyName"`
	Country           string  `json:"country"`
	Name              string  `json:"name"`
	Date              string  `json:"date"`
	Time              string  `json:"time"`
	RecurrenceCount   int     `json:"recurrenceCount"`
	BasePrice         float64 `json:"basePrice"`
	High52WPrice      float64 `json:"high52wPrice"`
	High52WDate       string  `json:"high52wDate"`
	Low52WPrice       float64 `json:"low52wPrice"`
	Low52WDate        string  `json:"low52wDate"`
	Provider          string  `json:"provider"`
	Timestamp         int64   `json:"timestamp"`
	ID                int     `json:"id"`
	CreatedAt         string  `json:"createdAt"`
	ModifiedAt        string  `json:"modifiedAt"`
	SignedChangePrice float64 `json:"signedChangePrice"`
	SignedChangeRate  float64 `json:"signedChangeRate"`
	ChangeRate        float64 `json:"changeRate"`
}
