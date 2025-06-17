package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
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

type ECDSASignature struct {
	R, S *big.Int
}

func GenerateSignature(tx *Transaction, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := tx.Hash()
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(ECDSASignature{R: r, S: s})
}

func VerifyTransaction(tx *Transaction, publicKey *ecdsa.PublicKey) bool {
	var sig ECDSASignature
	_, err := asn1.Unmarshal(tx.Signature, &sig)
	if err != nil {
		return false
	}
	hash := tx.Hash()
	return ecdsa.Verify(publicKey, hash, sig.R, sig.S)
}

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
