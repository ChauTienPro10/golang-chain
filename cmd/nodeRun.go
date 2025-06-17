package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/chauduongphattien/golang-chain/internal/handlers"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := storage.NewStorage("./pkg/storage/data")
	defer db.Close()

	leaderHandler := handlers.NewLeaderHandler(db)
	http.HandleFunc("/hello", leaderHandler.Hello)
	http.HandleFunc("/leader/transaction", leaderHandler.HandleTransaction)
	http.HandleFunc("/mempool", leaderHandler.GetMemPoolHandler)
	http.HandleFunc("/leader/genBlock", leaderHandler.CreateBlockHandler)

	commonHandler := handlers.NewCommonHandler(db)
	http.HandleFunc("/wallet/new", commonHandler.CreateWalletHandler)
	http.HandleFunc("/wallet/get", commonHandler.GetWalletHandler)
	http.HandleFunc("/wallet/getAll", commonHandler.GetAllWalletsHandler)

	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
