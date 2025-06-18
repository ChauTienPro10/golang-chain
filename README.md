# 🧱 Mini Blockchain System with Leader-Follower Consensus

> Châu Dương Phát Tiến.
> chauduongphattien2201@gmail.com
> 0812788212

## 📦 Modules Overview

### 1. `blockchain` – Core Blockchain Logic

> **Path**: `internal/blockchain`

* `block.go`: Định nghĩa một `Block` và các hàm tính **hash**, **Merkle Root**.
* `transaction.go`: Định nghĩa và xử lý các giao dịch.

---

### 2. `storage` – LevelDB Storage Layer

> **Path**: `pkg/storage`

* `db.go`: Cung cấp các phương thức lưu trữ và truy vấn block, ví... từ **LevelDB**.

---

### 3. `network` – Wallet and Cryptography

> **Path**: `internal/network`

* `wallet.go`: Quản lý các ví, địa chỉ mạng, khóa công khai và riêeng.

---

### 4. `handlers` – HTTP & gRPC Endpoints

> **Files**:

* `leaderHandler.go`, `followerHandler.go`, `commonHandler.go`: Xử lý các HTTP request như tạo ví, giao dịch, tạo block, v.v.
* `grpcclient`, `grpcserver`: Gửi và nhận các gói tin **proposal block** qua gRPC.

---

### 5. `ProposeBlock.proto` – gRPC Interface

> Định nghĩa toàn bộ các RPC để gửi proposal block giữa leader và followers.

---

## 🚀 Getting Started

### 1. ✅ Prerequisites

* Cài đặt **Docker**.

### 2. ▶️ Run the System

```bash (run từ project_root)
docker-compose up -d
```

* **Leader**:

  * HTTP: `localhost:8080`
  * gRPC: `50050`

* **Follower1**:

  * HTTP: `localhost:8081`
  * gRPC: `50051`

* **Follower2**:

  * HTTP: `localhost:8082`
  * gRPC: `50052`

---

## 🔍 Usage Guide

Bạn có thể sử dụng **Postman** hoặc **curl**.

### A. Postman API Endpoints

| Method | Endpoint                                        | Description              |
| ------ | ----------------------------------------------- | ------------------------ |
| POST   | `localhost:8080/wallet/new`                     | Tạo ví mới               |
| GET    | `localhost:8080/wallet/getAll`                  | Xem danh sách ví         |
| POST   | `localhost:8080/leader/transaction`             | Gửi giao dịch            |
| POST   | `localhost:8080/leader/genBlock`                | Tạo block mới            |
| POST   | `localhost:8080/leader/proposal`                | Gửi proposal và bỏ phiếu |
| POST   | `/follower/sync` (port 8081/8082)               | Follower đồng bộ block   |
| GET    | `/wallet/getLatesBlock` (port 8081/8082/8080)   | Xem block cuối cùng      |

---

### B. Curl Commands (nên dùng PowerShell trên Windows)

* **Tạo ví**:

```bash
curl -X POST http://localhost:8080/wallet/new \
     -H "Content-Type: application/json" \
     -d "{\"name\": \"Alice\", \"token\": 100}"

> or

curl -X POST http://localhost:8080/wallet/new -H "Content-Type: application/json" -d "{\"name\": \"Alice\", \"token\": 100}"
```

* **Xem danh sách ví**:

```bash
curl http://localhost:8080/wallet/getAll | python -m json.tool
```

* **Thực hiện giao dịch**:

```bash
curl -X POST http://localhost:8080/leader/transaction \
     -H "Content-Type: application/json" \
     -d "{\"sender\":\"<sender_addr>\", \"receiver\":\"<receiver_addr>\", \"amount\":10}"

> or

curl -X POST http://localhost:8080/leader/transaction -H "Content-Type: application/json" -d "{\"sender\":\"bd21bcd2190b5f5626c04b90e3e9e8e1eb87be46ea9b549ef5fd74f1f62962fa\", \"receiver\":\"e078dd883f05bd0590c9c55a31f3afb218bc43d3448aaa1d0f7c59efcdfecafe\", \"amount\":10}"

```

* **Tạo block**:

```bash
curl -X POST http://localhost:8080/leader/genBlock | python -m json.tool
```

* **Gửi proposal và vote**:

```bash
curl -X POST http://localhost:8080/leader/proposal
```

* **Đồng bộ follower**:

```bash
curl -X POST http://localhost:8081/follower/sync | python -m json.tool
```
=> sẽ trả về null nếu tất cả đã đông bộ


* **Xem block cuối**:

```bash
curl localhost:8080/wallet/getLatesBlock | python -m json.tool
```

*(có thể thay 8080 bằng 8081 hoặc 8082)*

---

## ✨ Tips

* Theo dõi log của container leader để quan sát quá trình proposal:

```bash
docker logs -f leader
```

* Dử đoán block chain consistency bằng cách so sánh `latest block` giữa các node.

---

Happy ⚡️!
