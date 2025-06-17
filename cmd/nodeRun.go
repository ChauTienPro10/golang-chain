package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chauduongphattien/golang-chain/internal/handlers"
)


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // mặc định nếu không có biến môi trường
	}

	leaderHandler := handlers.NewLeaderHandler()
	http.HandleFunc("/hello", leaderHandler.Hello)

	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}