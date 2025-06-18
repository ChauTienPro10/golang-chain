package utils
import (
	"github.com/chauduongphattien/golang-chain/internal/blockchain"

	pb "github.com/chauduongphattien/golang-chain/blockchain/proposalpb"

)

func ConvertFromProtoBlock(pbBlock *pb.Block) *blockchain.Block {
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

func ConvertToProtoBlock(b *blockchain.Block) *pb.Block {
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