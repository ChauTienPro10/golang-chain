package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/chauduongphattien/golang-chain/internal/p2p/grpcclient"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
)

type FollowerHandler struct {
	storageInst *storage.Storage
	leaderAddr  string
}

func NewFollowerHandler(storage *storage.Storage) *FollowerHandler {
	return &FollowerHandler{
		storageInst: storage,
		leaderAddr:  getLeaderAddr(),
	}
}

func getLeaderAddr() string {
	raw := os.Getenv("LEADER")
	if raw == "" {
		raw = "leader:50050"
	}
	return raw
}

func (h *FollowerHandler) HandleSyncBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	lastBlock, err := h.storageInst.GetLatestBlock()
	if err != nil {
		http.Error(w, fmt.Sprintf("Lỗi đồng bộ từ leader, get Block cuoi that bai: %v", err), http.StatusInternalServerError)
		return
	}

	blocks, err := grpcclient.SyncFromLeader(h.leaderAddr, lastBlock.Hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Lỗi đồng bộ từ leader: %v", err), http.StatusInternalServerError)
		return
	}

	for _, block := range blocks {
		err := h.storageInst.SaveBlock(block)
		if err != nil {
			http.Error(w, fmt.Sprintf("Lỗi lưu block về local: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(blocks)
	if err != nil {
		http.Error(w, "Lỗi encode kết quả", http.StatusInternalServerError)
		return
	}
}
