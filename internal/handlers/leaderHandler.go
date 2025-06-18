package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"log"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/internal/network"
	"github.com/chauduongphattien/golang-chain/pkg/storage"

	"github.com/chauduongphattien/golang-chain/internal/p2p/grpcclient"
	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"
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
	followerAddrs []string
}

func NewLeaderHandler(storage *storage.Storage) *LeaderHandler {
	return &LeaderHandler{
		memPool:     []blockchain.Transaction{},
		storageInst: storage,
		voteCount:   0,
		followerAddrs: []string{"localhost:50051", "localhost:50052"},
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

func (h *LeaderHandler) SendProposal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	if h.pendingBlk == nil {
		return
	}

	h.GenProposeBlock(h.pendingBlk, h.followerAddrs)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Đã gửi proposal đến các follower"))
}


func (h *LeaderHandler) GenProposeBlock(b *blockchain.Block, followerAddrs []string) {
	protoBlock := convertToProtoBlock(b)

	req := &pb.ProposalRequest{
		Block:    protoBlock,
		LeaderID: "leader-1", // Có thể thay đổi theo cấu hình hoặc ID thật sự của leader
	}

	for _, addr := range followerAddrs {
		go func(address string) {
			resp, err := grpcclient.SendProposalToFollower(address, req)
			if err != nil {
				log.Printf("Gửi proposal đến %s thất bại: %v\n", address, err)
				return
			}
			log.Printf("Follower %s phản hồi: %s (accepted: %v)\n", address, resp.Message, resp.Accepted)
		}(addr)
	}
}

func convertToProtoBlock(b *blockchain.Block) *pb.Block {
	var txs []*pb.Transaction
	for _, t := range b.Transactions {
		txs = append(txs, &pb.Transaction{
			Sender:    t.Sender,
			Receiver:  t.Receiver,
			Amount:    t.Amount,
			Timestamp: t.Timestamp,
			Signature: t.Signature,
		})
	}

	return &pb.Block{
		Timestamp:    b.Timestamp,
		Transactions: txs,
		MerkleRoot:   b.MerkleRoot,
		PrevHash:     b.PrevHash,
		Nonce:        int32(b.Nonce),
		Hash:         b.Hash,
	}
}

func (lh *LeaderHandler) HandleVote(w http.ResponseWriter, r *http.Request) {

}

func (lh *LeaderHandler) CommitBlock() {

}
