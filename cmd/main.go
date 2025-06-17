package main

import (
	"fmt"
	"log"
	"time"
	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/pkg/storage"
)

func main() {
    privKeyAlice, err := blockchain.GenerateKeyPair()
    if err != nil {
        log.Fatal(err)
    }
    pubKeyAlice := &privKeyAlice.PublicKey

    timestamp := time.Now().Unix()

	tx := &blockchain.Transaction{
		Sender:    "Alice",
		Receiver:  "Bob",
		Amount:    10.0,
		Timestamp: timestamp,
	}	

    sig, err := blockchain.GenerateSignature(tx, privKeyAlice)
    if err != nil {
        log.Fatal(err)
    }
    tx.Signature = sig

    isValid := blockchain.VerifyTransaction(tx, pubKeyAlice)
    fmt.Println("Chữ ký hợp lệ?", isValid)

	// them mot block 
	block := blockchain.NewBlock(0, timestamp, []blockchain.Transaction{*tx}, "")
	fmt.Println("Block vừa tạo:")
	fmt.Printf("Index: %d\n", block.Index)
	fmt.Printf("Merkle Root: %s\n", block.MerkleRoot)
	fmt.Printf("Block Hash: %s\n", block.Hash)

	db := storage.NewStorage("./pkg/storage/data")
	defer db.Close()

	// saveBlockErr := db.SaveBlock(block)
	// if saveBlockErr != nil {
	// log.Fatalf("Lưu block thất bại: %v", saveBlockErr)
	// }
	// fmt.Println("Block đã được lưu vào DB thành công.")

	loadedBlock, err := db.LoadBlock("0f9f3825ca6bd9baf32253e4e70aa991b346290373857f47e11b00bcf69221c1")
	if err != nil {
		log.Fatalf("Không thể load block: %v", err)
	}

	// In thông tin block đã load
	fmt.Println("Block đã load từ DB:")
	fmt.Printf("Index: %d\n", loadedBlock.Index)
	fmt.Printf("Hash: %s\n", loadedBlock.Hash)
	fmt.Printf("Merkle Root: %s\n", loadedBlock.MerkleRoot)
	fmt.Printf("PrevHash: %s\n", loadedBlock.PrevHash)
	fmt.Printf("Số giao dịch: %d\n", len(loadedBlock.Transactions))

}