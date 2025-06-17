package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/chauduongphattien/golang-chain/internal/network"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
)

type CommonHandler struct {
	storageInst *storage.Storage
}

func NewCommonHandler(storage *storage.Storage) *CommonHandler {
	return &CommonHandler{
		storageInst: storage,
	}
}

func (h *CommonHandler) HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from CommonHandler!"))
}

type WalletResponse struct {
	Address string `json:"address"`
	Token   int    `json:"token"`
}

type CreateWalletRequest struct {
	Name  string `json:"name"`
	Token int    `json:"token"`
}

func (h *CommonHandler) CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		// Xử lý preflight request từ browser (CORS)
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Chỉ hỗ trợ POST", http.StatusMethodNotAllowed)
		return
	}

	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	newWallet := network.NewWallet(req.Token)
	if err := h.storageInst.SaveWallet(newWallet.Address, newWallet); err != nil {
		http.Error(w, "Không thể lưu ví", http.StatusInternalServerError)
		return
	}
	resp := WalletResponse{
		Address: newWallet.Address,
		Token:   newWallet.Token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CommonHandler) GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodGet {
		http.Error(w, "Chỉ hỗ trợ GET", http.StatusMethodNotAllowed)
		return
	}

	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Thiếu address", http.StatusBadRequest)
		return
	}

	walletData, err := h.storageInst.LoadWallet(address)
	if err != nil {
		http.Error(w, "Không tìm thấy ví", http.StatusNotFound)
		return
	}

	resp := WalletResponse{
		Address: walletData.Address,
		Token:   walletData.Token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *CommonHandler) GetAllWalletsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Chỉ hỗ trợ GET", http.StatusMethodNotAllowed)
		return
	}

	wallets, err := h.storageInst.LoadAllWallets()
	if err != nil {
		http.Error(w, "Lỗi khi lấy danh sách ví", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallets)
}
