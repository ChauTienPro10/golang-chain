package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Block struct {
	Index        int           
	Timestamp    int64         
	Transactions []Transaction 
	MerkleRoot   string         
	PrevHash     string        
	Nonce        int            
	Hash         string         
}

func NewBlock(index int, timestamp int64, transactions []Transaction, prevHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    timestamp,
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
	}

	block.MerkleRoot = CalculateMerkleRoot(transactions)
	block.Hash = block.CalculateHash()

	return block
}


func CalculateMerkleRoot(txs []Transaction) string {
	if len(txs) == 0 {
		return ""
	}
	var hashes [][]byte
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	for len(hashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				concat := append(hashes[i], hashes[i+1]...)
				newHash := sha256.Sum256(concat)
				newLevel = append(newLevel, newHash[:])
			} else {
				newLevel = append(newLevel, hashes[i]) 
			}
		}
		hashes = newLevel
	}
	return hex.EncodeToString(hashes[0])
}

func (b *Block) CalculateHash() string {
	data := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.MerkleRoot, b.PrevHash, b.Nonce)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}