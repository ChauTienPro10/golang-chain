package service

import (
	"context"
	"log"
	"github.com/syndtr/goleveldb/leveldb"

	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/pkg/storage"

)

type ProposalServer struct {
	pb.UnimplementedProposalServiceServer
	Storage *storage.Storage
}

func NewProposalServer(store *storage.Storage) *ProposalServer {
	return &ProposalServer{Storage: store}
}


func (s *ProposalServer) SendProposal(ctx context.Context, req *pb.ProposalRequest) (*pb.ProposalResponse, error) {
	log.Println("Nhận đề xuất block từ leader:", req.LeaderID)
	log.Println("Block hash:", req.Block.Hash)

	block := convertFromProtoBlock(req.Block)

	log.Printf("Thông tin block nhận được:\n"+
	"  Timestamp: %d\n"+
	"  PrevHash: %s\n"+
	"  Hash: %s\n"+
	"  MerkleRoot: %s\n"+
	"  Transactions:\n", block.Timestamp, block.PrevHash, block.Hash, block.MerkleRoot)

	for i, tx := range block.Transactions {
		log.Printf("    Tx #%d - Sender: %s, Receiver: %s, Amount: %.6f, Timestamp: %d, Signature: %s", i+1, tx.Sender, tx.Receiver, tx.Amount,tx.Timestamp, tx.Signature)
	}

	calculatedRoot := blockchain.CalculateMerkleRoot(block.Transactions)
	if calculatedRoot != block.MerkleRoot {
		return &pb.ProposalResponse{
			Message:  "Merkle Root không khớp",
			Accepted: false,
		}, nil
	}

	lastBlock, errLoadLatesBl := s.Storage.GetLatestBlock()
	
	if errLoadLatesBl != nil && errLoadLatesBl != leveldb.ErrNotFound {
		log.Println("Lỗi khi load block cuối cùng:", errLoadLatesBl)

		return &pb.ProposalResponse{
			Message:  "Khong the load block cuoi",
			Accepted: false,
		}, nil
	}
	
	if lastBlock != nil {
		if block.PrevHash != lastBlock.Hash {
				return &pb.ProposalResponse{
					Message:  "Block không nối tiếp đúng",
					Accepted: false,
				}, nil
			}
	}

	return &pb.ProposalResponse{
		Message:  "Block hợp lệ, đã nhận",
		Accepted: true,
	}, nil
}


func convertFromProtoBlock(pbBlock *pb.Block) *blockchain.Block {
	txs := make([]blockchain.Transaction, 0)
	for _, pbTx := range pbBlock.Transactions {
		txs = append(txs, blockchain.Transaction{
			Sender:    pbTx.Sender,
			Receiver:  pbTx.Receiver,
			Amount:    pbTx.Amount,
			Timestamp: pbTx.Timestamp,
			Signature: pbTx.Signature,
		})
	}

	return &blockchain.Block{
		Timestamp:    pbBlock.Timestamp,
		Transactions: txs, 
		PrevHash:     pbBlock.PrevHash,
		Hash:         pbBlock.Hash,
		MerkleRoot:   pbBlock.MerkleRoot,
	}
}