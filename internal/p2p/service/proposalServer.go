package service

import (
	"context"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"

	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/internal/p2p/utils"
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

	block := utils.ConvertFromProtoBlock(req.Block)

	log.Printf("Thông tin block nhận được:\n"+
		"  Timestamp: %d\n"+
		"  PrevHash: %s\n"+
		"  Hash: %s\n"+
		"  MerkleRoot: %s\n"+
		"  Transactions:\n", block.Timestamp, block.PrevHash, block.Hash, block.MerkleRoot)

	for i, tx := range block.Transactions {
		log.Printf("    Tx #%d - Sender: %s, Receiver: %s, Amount: %.6f, Timestamp: %d, Signature: %s", i+1, tx.Sender, tx.Receiver, tx.Amount, tx.Timestamp, tx.Signature)
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

// cu ly commit block
func (s *ProposalServer) CommitBlock(ctx context.Context, req *pb.CommitBlockRequest) (*pb.CommitBlockResponse, error) {
	block := utils.ConvertFromProtoBlock(req.Block)

	err := s.Storage.SaveBlock(block)
	if err != nil {
		log.Println("Lỗi khi commit block:", err)
		return &pb.CommitBlockResponse{
			Message: "Commit thất bại",
			Success: false,
		}, nil
	}

	log.Println("Block đã được lưu vào local storage (Commit)")
	return &pb.CommitBlockResponse{
		Message: "Commit thành công",
		Success: true,
	}, nil
}

// xu ly dong bo Block
func (s *ProposalServer) SyncMissingBlocks(ctx context.Context, req *pb.SyncBlocksRequest) (*pb.SyncBlocksResponse, error) {
	knownHash := req.FromHash
	var blocks []*pb.Block

	current, err := s.Storage.GetLatestBlock()
	if err != nil {
		log.Printf("Lỗi lấy block cuối: %v", err)
		return nil, err
	}

	log.Printf("cur block: %x\n", current.Hash)
	log.Printf("knownHash: %x\n", knownHash)

	for current != nil && current.Hash != knownHash {
		protoBlk := utils.ConvertToProtoBlock(current)
		blocks = append([]*pb.Block{protoBlk}, blocks...) // prepend

		current, err = s.Storage.LoadBlock(current.PrevHash)
		if err != nil {
			log.Printf("Lỗi load block theo prevHash: %v", err)
			break
		}
	}

	if current == nil || current.Hash != knownHash {
		return nil, fmt.Errorf("Không tìm thấy block có hash %x", knownHash)
	}

	return &pb.SyncBlocksResponse{Blocks: blocks}, nil
}
