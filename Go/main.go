package main

import (
	"fmt"
	"net/http"

	"github.com/Son0-0/redis-test/handlers"
)

func main() {

	api := handlers.NewAPI()

	http.HandleFunc("/api", api.OpenAPI)
	http.HandleFunc("/db", api.DB)

	fmt.Println("Server Running :9090")
	http.ListenAndServe(":9090", nil)
}
