package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
	"github.com/chauduongphattien/golang-chain/internal/network"
	"github.com/syndtr/goleveldb/leveldb"
)

type Storage struct {
	db *leveldb.DB
}

func NewStorage(path string) *Storage {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalf("Không thể mở LevelDB: %v", err)
	}
	return &Storage{db: db}
}

func (s *Storage) Put(key string, value []byte) error {
	return s.db.Put([]byte(key), value, nil)
}

func (s *Storage) Get(key string) ([]byte, error) {
	return s.db.Get([]byte(key), nil)
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) SaveBlock(block *blockchain.Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}
	blockKey := "block_" + block.Hash
	if err := s.db.Put([]byte(blockKey), data, nil); err != nil {
		return err
	}
	return s.db.Put([]byte("last_block_hash"), []byte(block.Hash), nil)
}

func (s *Storage) LoadBlock(hash string) (*blockchain.Block, error) {
	data, err := s.db.Get([]byte("block_"+hash), nil)
	if err != nil {
		return nil, err
	}
	var block blockchain.Block
	err = json.Unmarshal(data, &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// vi
func (s *Storage) SaveWallet(address string, wallet *network.Wallet) error {
	data, err := json.Marshal(wallet)
	if err != nil {
		return err
	}
	fmt.Println(">> Lưu ví:", address)
	return s.db.Put([]byte("wallet:"+address), data, nil)
}

func (s *Storage) LoadWallet(address string) (*network.Wallet, error) {
	fmt.Println(">> Truy vấn ví:", address)
	data, err := s.db.Get([]byte("wallet:"+address), nil)
	if err != nil {
		return nil, err
	}

	var w network.Wallet
	if err := json.Unmarshal(data, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Storage) LoadAllWallets() ([]*network.Wallet, error) {
	var wallets []*network.Wallet

	iter := s.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		if !strings.HasPrefix(string(key), "wallet:") {
			continue
		}

		var w network.Wallet
		if err := json.Unmarshal(iter.Value(), &w); err != nil {
			continue // bỏ qua nếu lỗi
		}
		wallets = append(wallets, &w)
	}
	iter.Release()
	return wallets, nil
}

func (s *Storage) GetLatestBlock() (*blockchain.Block, error) {
	hashBytes, err := s.db.Get([]byte("last_block_hash"), nil)
	if err != nil {
		return nil, err
	}

	blockData, err := s.db.Get([]byte("block_"+string(hashBytes)), nil)
	if err != nil {
		return nil, err
	}

	var block blockchain.Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

// 0f9f3825ca6bd9baf32253e4e70aa991b346290373857f47e11b00bcf69221c1
