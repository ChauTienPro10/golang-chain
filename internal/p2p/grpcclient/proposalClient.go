package grpcclient

import (
	"context"
	"log"
	"time"

	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"
	"google.golang.org/grpc"

	"github.com/chauduongphattien/golang-chain/internal/p2p/utils"
	"github.com/chauduongphattien/golang-chain/internal/blockchain"
)


// SendProposalToFollower gửi proposal từ Leader đến một Follower cụ thể qua gRPC.
func SendProposalToFollower(followerAddr string, proposal *pb.ProposalRequest) (*pb.ProposalResponse, error) {
	// Thiết lập kết nối đến follower (ví dụ: "localhost:50051")
	conn, err := grpc.Dial(followerAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(3*time.Second))
	if err != nil {
		log.Printf("Không thể kết nối đến follower %s: %v", followerAddr, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewProposalServiceClient(conn)

	// Gửi proposal với context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SendProposal(ctx, proposal)
	if err != nil {
		log.Printf("Gửi proposal đến follower %s thất bại: %v", followerAddr, err)
		return nil, err
	}

	log.Printf("Proposal gửi đến %s thành công. Phản hồi: %s (Accepted: %v)", followerAddr, resp.Message, resp.Accepted)
	return resp, nil
}

func SendCommitBlockToFollower(address string, req *pb.CommitBlockRequest) (*pb.CommitBlockResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	if err != nil {
		log.Printf("Không thể kết nối đến %s: %v", address, err)
		return nil, err
	}
	defer conn.Close()

	client := pb.NewProposalServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return client.CommitBlock(ctx, req)
}

func SyncFromLeader(leaderAddr string, lastHash string) ([]*blockchain.Block, error) {
	conn, err := grpc.Dial(leaderAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewProposalServiceClient(conn)

	req := &pb.SyncBlocksRequest{FromHash: lastHash}
	resp, err := client.SyncMissingBlocks(context.Background(), req)
	if err != nil {
		return nil, err
	}

	var blocks []*blockchain.Block
	for _, protoBlk := range resp.Blocks {
		block := utils.ConvertFromProtoBlock(protoBlk)
		blocks = append(blocks, block)
	}

	return blocks, nil
}