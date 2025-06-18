package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
	"log"
	"os"
	"strings"
	"fmt"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/internal/network"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
	"github.com/chauduongphattien/golang-chain/internal/p2p/utils"

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
		followerAddrs: getFollowerAddrs(),
	}
}

func getFollowerAddrs() []string {
	raw := os.Getenv("FOLLOWERS")
	if raw == "" {
		return []string{} 
	}
	return strings.Split(raw, ",")
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
	Sender   string `json:"sender"`   
	Receiver string `json:"receiver"` 
	Amount   int    `json:"amount"`   
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

	walletData, err := h.storageInst.LoadWallet(trans.Sender)
	if err != nil {
		http.Error(w, "Không tìm thấy ví", http.StatusNotFound)
		return
	}

	if walletData.Token < trans.Amount {
		http.Error(w, "Số dư không đủ", http.StatusBadRequest)
		return
	}

	pubKey, err := network.ParsePublicKey(walletData.PublicKey)
	if err != nil {
		http.Error(w, "Public key không hợp lệ", http.StatusBadRequest)
		return
	}

	timestamp := time.Now().Unix()
	tx := &blockchain.Transaction{
		Sender:    trans.Sender,
		Receiver:  trans.Receiver,
		Amount:    float64(trans.Amount),
		Timestamp: timestamp,
	}

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

	if !network.VerifyTransaction(tx, pubKey) {
		http.Error(w, "Giao dịch không hợp lệ (sai chữ ký)", http.StatusBadRequest)
		return
	}

	h.memPool = append(h.memPool, *tx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Giao dịch đã được nhận và đang chờ xử lý",
	})
}

type ProposalRequest struct {
	Block    *blockchain.Block `json:"block"`     
	LeaderID string            `json:"leader_id"`
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

	lastBlock, err := h.storageInst.GetLatestBlock()
	var prevHash string
	if err == nil && lastBlock != nil {
		prevHash = lastBlock.Hash
	}

	timestamp := time.Now().Unix()
	newBlock := blockchain.NewBlock(h.memPool, prevHash, timestamp)

	h.pendingBlk = newBlock
	h.memPool = []blockchain.Transaction{} // Clear mempool

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
	protoBlock := utils.ConvertToProtoBlock(b)

	req := &pb.ProposalRequest{
		Block:    protoBlock,
		LeaderID: "leader-1", 
	}
	var wg sync.WaitGroup 
	for _, addr := range followerAddrs {
		wg.Add(1)

		go func(address string) {
			defer wg.Done()
			resp, err := grpcclient.SendProposalToFollower(address, req)
			if err != nil {
				log.Printf("Gửi proposal đến %s thất bại: %v\n", address, err)
				return
			}
			log.Printf("Follower %s phản hồi: %s (accepted: %v)\n", address, resp.Message, resp.Accepted)
			if resp.Accepted {
				h.voteMu.Lock()
				h.voteCount++
				h.voteMu.Unlock()
			}	
		}(addr)
	}
	
	wg.Wait()

	log.Printf("Số phiếu đồng thuận nhận được: %d\n", h.voteCount)

	if h.voteCount >= 1 {
		log.Println("Đủ phiếu, tiến hành gửi commit đến follower")

		commitReq := &pb.CommitBlockRequest{
			Block: protoBlock,
		}

		for _, addr := range followerAddrs {
			wg.Add(1)
			go func(address string) {
				defer wg.Done()
				resp, err := grpcclient.SendCommitBlockToFollower(address, commitReq)
				if err != nil {
					log.Printf("Gửi commit đến %s thất bại: %v\n", address, err)
					return
				}
				log.Printf("Commit xác nhận từ %s: %s (success: %v)\n", address, resp.Message, resp.Success)
			}(addr)
		}
		wg.Wait()
	} else {
		log.Println("Không đủ phiếu, không gửi commit")
	}
	
}

func (h *LeaderHandler) HandleSyncBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		LeaderAddr string `json:"leaderAddr"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Lỗi đọc JSON request body", http.StatusBadRequest)
		return
	}

	lastBlock, err := h.storageInst.GetLatestBlock()

	blocks, err := grpcclient.SyncFromLeader(req.LeaderAddr, lastBlock.Hash)
	if err != nil {
		http.Error(w, fmt.Sprintf("Lỗi đồng bộ từ leader: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(blocks)
	if err != nil {
		http.Error(w, "Lỗi encode kết quả", http.StatusInternalServerError)
		return
	}
}
