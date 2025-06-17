package storage

import (
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"encoding/json"

	"github.com/chauduongphattien/golang-chain/internal/blockchain"
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
	return s.db.Put([]byte("last_hash"), []byte(block.Hash), nil)
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

// 0f9f3825ca6bd9baf32253e4e70aa991b346290373857f47e11b00bcf69221c1