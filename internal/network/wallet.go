package network

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
)

type ECDSASignature struct {
	R, S *big.Int
}

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

type Wallet struct {
	Address    string
	PrivateKey []byte
	PublicKey  []byte
	Token      int
}

func NewWallet(token int) *Wallet {
	privKey, err := GenerateKeyPair()
	if err != nil {
		panic(err)
	}
	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)
	address := GetAddressFromPubKey(pubKey)

	return &Wallet{
		PrivateKey: privKey.D.Bytes(),
		PublicKey:  pubKey,
		Token:      token,
		Address:    address,
	}
}

func GetAddressFromPubKey(pubKey []byte) string {
	hash := sha256.Sum256(pubKey)
	return hex.EncodeToString(hash[:])
}

func GenerateSignature(tx *blockchain.Transaction, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	hash := tx.Hash()
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}
	return asn1.Marshal(ECDSASignature{R: r, S: s})
}

func VerifyTransaction(tx *blockchain.Transaction, publicKey *ecdsa.PublicKey) bool {
	var sig ECDSASignature
	_, err := asn1.Unmarshal(tx.Signature, &sig)
	if err != nil {
		return false
	}
	hash := tx.Hash()
	return ecdsa.Verify(publicKey, hash, sig.R, sig.S)
}

func ParsePublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	keyLen := len(pubKeyBytes) / 2
	if keyLen == 0 || len(pubKeyBytes)%2 != 0 {
		return nil, errors.New("public key không hợp lệ")
	}

	x := new(big.Int).SetBytes(pubKeyBytes[:keyLen])
	y := new(big.Int).SetBytes(pubKeyBytes[keyLen:])

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func ParsePrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	if len(data) == 0 {
		return nil, errors.New("dữ liệu private key rỗng")
	}

	curve := elliptic.P256()
	D := new(big.Int).SetBytes(data)

	priv := new(ecdsa.PrivateKey)
	priv.D = D
	priv.PublicKey.Curve = curve
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(D.Bytes())

	return priv, nil
}
