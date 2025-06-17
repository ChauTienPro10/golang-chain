package handlers

import (
	"net/http"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
)

type VoteRequest struct {
	BlockHash string `json:"block_hash"`
	Accepted  bool   `json:"accepted"`
}

type LeaderHandler struct {
	memPool     []blockchain.Transaction
	storageInst *storage.Storage
	voteCount   int
	voteMu      sync.Mutex
	totalVotes  int
	pendingBlk  *blockchain.Block
}

func NewLeaderHandler(storage *storage.Storage) *LeaderHandler {
	return &LeaderHandler{
		memPool:     []blockchain.Transaction{},
		storageInst: storage,
		voteCount:   0,
	}
}

func (h *LeaderHandler) Hello(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from LeaderHandler!"))
}

// Handle nhận transaction từ client
func (h *LeaderHandler) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	
}

func (h *LeaderHandler) CreateBlock() {
	
}


func (h *LeaderHandler) CreateBlock() {
	
}


func (h *LeaderHandler) ProposeBlock() {

}

func (lh *LeaderHandler) HandleVote(w http.ResponseWriter, r *http.Request) {

}

func (lh *LeaderHandler) CommitBlock() {
	
}