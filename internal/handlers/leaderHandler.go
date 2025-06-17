package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/internal/network"
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

type TransRequest struct {
	Sender   string `json:"sender"`   // Địa chỉ ví người gửi
	Receiver string `json:"receiver"` // Địa chỉ ví người nhận
	Amount   int    `json:"amount"`   // Số token gửi
}

func (h *LeaderHandler) GetMemPoolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Chỉ hỗ trợ GET", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.memPool)
}

func (h *LeaderHandler) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	var trans TransRequest
	if err := json.NewDecoder(r.Body).Decode(&trans); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Load ví
	walletData, err := h.storageInst.LoadWallet(trans.Sender)
	if err != nil {
		http.Error(w, "Không tìm thấy ví", http.StatusNotFound)
		return
	}

	if walletData.Token < trans.Amount {
		http.Error(w, "Số dư không đủ", http.StatusBadRequest)
		return
	}

	// Parse public key
	pubKey, err := network.ParsePublicKey(walletData.PublicKey)
	if err != nil {
		http.Error(w, "Public key không hợp lệ", http.StatusBadRequest)
		return
	}

	// Tạo transaction
	timestamp := time.Now().Unix()
	tx := &blockchain.Transaction{
		Sender:    trans.Sender,
		Receiver:  trans.Receiver,
		Amount:    float64(trans.Amount),
		Timestamp: timestamp,
	}

	// Parse private key và ký
	privateKey, err := network.ParsePrivateKey(walletData.PrivateKey)
	if err != nil {
		http.Error(w, "Private key không hợp lệ", http.StatusBadRequest)
		return
	}

	sign, err := network.GenerateSignature(tx, privateKey)
	if err != nil {
		http.Error(w, "Không thể tạo chữ ký", http.StatusInternalServerError)
		return
	}
	tx.Signature = sign

	// Xác thực chữ ký
	if !network.VerifyTransaction(tx, pubKey) {
		http.Error(w, "Giao dịch không hợp lệ (sai chữ ký)", http.StatusBadRequest)
		return
	}

	// Lưu vào mempool
	h.memPool = append(h.memPool, *tx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Giao dịch đã được nhận và đang chờ xử lý",
	})
}

type ProposalRequest struct {
	Block    *blockchain.Block `json:"block"`     // Block được đề xuất
	LeaderID string            `json:"leader_id"` // ID hoặc địa chỉ node leader gửi proposal
}

func (h *LeaderHandler) CreateBlockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	if len(h.memPool) == 0 {
		http.Error(w, "Không có giao dịch trong memPool", http.StatusBadRequest)
		return
	}

	// Lấy block cuối cùng từ DB
	lastBlock, err := h.storageInst.GetLatestBlock()
	var prevHash string
	if err == nil && lastBlock != nil {
		prevHash = lastBlock.Hash
	}

	// Tạo block mới
	timestamp := time.Now().Unix()
	newBlock := blockchain.NewBlock(h.memPool, prevHash, timestamp)

	// Cập nhật trạng thái
	h.pendingBlk = newBlock
	h.memPool = []blockchain.Transaction{} // Clear mempool

	// Trả về block đã tạo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newBlock)
}

func (h *LeaderHandler) CreateBlock() {

}

var validatorNodes = []string{
	"http://localhost:8081/proposal",
	"http://localhost:8082/proposal",
}

func (h *LeaderHandler) ProposeBlock() {
	proposal := ProposalRequest{
		Block:    h.pendingBlk,
		LeaderID: "leader-node-1",
	}

	for _, nodeURL := range validatorNodes {
		go func(url string) {
			data, err := json.Marshal(proposal)
			if err != nil {
				fmt.Println("Lỗi khi mã hóa proposal:", err)
				return
			}

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
			if err != nil {
				fmt.Printf("Gửi proposal tới %s thất bại: %v\n", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("Node %s từ chối proposal: %s\n", url, string(body))
			} else {
				fmt.Printf("Proposal gửi thành công tới %s\n", url)
			}
		}(nodeURL)
	}
}

func (lh *LeaderHandler) HandleVote(w http.ResponseWriter, r *http.Request) {

}

func (lh *LeaderHandler) CommitBlock() {

}
