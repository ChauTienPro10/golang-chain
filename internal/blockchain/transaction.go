package blockchain

import (
	"crypto/sha256"
	"fmt"
)

type Transaction struct {
	Sender    string
	Receiver  string
	Amount    float64
	Timestamp int64
	Signature []byte
}

func NewTransactionPtr(sender, receiver string, amount float64, timestamp int64, signature []byte) *Transaction {
	return &Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Timestamp: timestamp,
		Signature: signature,
	}
}

func (tx *Transaction) Hash() []byte {
	data := fmt.Sprintf("%s:%s:%f:%d", tx.Sender, tx.Receiver, tx.Amount, tx.Timestamp)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}
