package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/chauduongphattien/golang-chain/internal/handlers"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
	"google.golang.org/grpc"

	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"
	"github.com/chauduongphattien/golang-chain/internal/p2p/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	tcpPort := os.Getenv("TCP_PORT")
	if tcpPort == "" {
		tcpPort = "50050"
	}

	db := storage.NewStorage("./pkg/storage/data")
	defer db.Close()

	leaderHandler := handlers.NewLeaderHandler(db)
	http.HandleFunc("/hello", leaderHandler.Hello)
	http.HandleFunc("/leader/transaction", leaderHandler.HandleTransaction)
	http.HandleFunc("/mempool", leaderHandler.GetMemPoolHandler)
	http.HandleFunc("/leader/genBlock", leaderHandler.CreateBlockHandler)
	http.HandleFunc("/leader/proposal", leaderHandler.SendProposal)

	followerHandler := handlers.NewFollowerHandler(db)
	http.HandleFunc("/follower/sync", followerHandler.HandleSyncBlock)

	commonHandler := handlers.NewCommonHandler(db)
	http.HandleFunc("/wallet/new", commonHandler.CreateWalletHandler)
	http.HandleFunc("/wallet/get", commonHandler.GetWalletHandler)
	http.HandleFunc("/wallet/getAll", commonHandler.GetAllWalletsHandler)
	http.HandleFunc("/wallet/getLatesBlock", commonHandler.GetLastBlock)

	go func() {
		lis, err := net.Listen("tcp", ":"+tcpPort)
		if err != nil {
			log.Fatalf("Không thể lắng nghe: %v", err)
		}

		grpcServer := grpc.NewServer()
		proposalServer := service.NewProposalServer(db)
		pb.RegisterProposalServiceServer(grpcServer, proposalServer)

		log.Println("Follower đang lắng nghe ở :" + tcpPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Lỗi khi chạy gRPC server: %v", err)
		}
	}()

	fmt.Printf("Server running at http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
