// mini_chain.go
// go run . 3000 localhost:3001,localhost:3002
// go run . 3001 localhost:3000,localhost:3002
// go run . 3002 localhost:3000,localhost:3001
// 区块链节点主程序，支持通过命令行参数配置端口和邻居节点

package main

import (
	"bufio"         // 用于读取网络连接和标准输入的数据
	"crypto/ecdsa"  // 椭圆曲线数字签名算法，用于钱包密钥
	"crypto/elliptic" // 椭圆曲线加密相关
	"crypto/rand"   // 加密安全的随机数生成器
	"crypto/sha256" // SHA256哈希函数
	"encoding/hex"  // 十六进制编码解码
	"encoding/json" // JSON序列化反序列化
	"fmt"           // 格式化输入输出
	"io"            // IO操作接口
	"log"           // 日志记录
	"net"           // 网络编程相关
	"os"            // 系统操作
	"strconv"       // 字符串与数值转换
	"strings"       // 字符串处理
	"sync"          // 同步原语，如互斥锁
	"time"          // 时间处理
)

// ===== 数据结构 =====
// Transaction 交易结构体，表示一笔转账交易
type Transaction struct {
	From      string `json:"from"`      // 发送方地址
	To        string `json:"to"`        // 接收方地址
	Amount    int    `json:"amount"`    // 转账金额
	Signature string `json:"signature"` // 交易签名，十六进制ASN.1编码格式
}

// Block 区块结构体，包含区块的所有信息
type Block struct {
	Index        int           `json:"index"`         // 区块索引（高度）
	Timestamp    int64         `json:"timestamp"`     // 时间戳
	Transactions []Transaction `json:"transactions"`  // 包含的交易列表
	PrevHash     string        `json:"prev_hash"`     // 前一个区块的哈希值
	Nonce        int64         `json:"nonce"`         // 工作量证明的随机数
	Hash         string        `json:"hash"`          // 当前区块的哈希值
}

// Message 网络消息结构体，用于节点间通信
type Message struct {
	Type string          `json:"type"` // 消息类型: "TX"(交易), "BLOCK"(区块), "GETCHAIN"(获取区块链), "CHAIN"(返回区块链)
	Data json.RawMessage `json:"data"` // 消息数据内容
}

// ===== 全局链与内存池 =====
// 区块链核心数据结构和相关同步控制
var (
	blockchain  []Block             // 区块链，按顺序存储所有区块
	txPool      []Transaction       // 交易池，存储待打包的交易
	chainMutex  sync.Mutex          // 区块链访问互斥锁，保证并发安全
	txPoolMutex sync.Mutex          // 交易池访问互斥锁，保证并发安全
	peers       []string            // 邻居节点地址列表
	addr        string              // 本节点地址，格式如"localhost:3000"
	difficulty  = 3                 // PoW挖矿难度：要求哈希前difficulty个0(十六进制字符串)
)

// ===== 钱包与签名工具 =====
// NewKeyPair 生成新的椭圆曲线密钥对，用于创建钱包地址
func NewKeyPair() (*ecdsa.PrivateKey, string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pubBytes := elliptic.Marshal(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	return priv, hex.EncodeToString(pubBytes)
}

// HashTransaction 计算交易的哈希值，用于签名和验证
func HashTransaction(tx Transaction) []byte {
	data := tx.From + "|" + tx.To + "|" + strconv.Itoa(tx.Amount)
	h := sha256.Sum256([]byte(data))
	return h[:]
}

// SignTransaction 使用私钥对交易进行签名
func SignTransaction(priv *ecdsa.PrivateKey, tx Transaction) (string, error) {
	h := HashTransaction(tx)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, h)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// VerifyTransaction 验证交易签名的有效性
func VerifyTransaction(tx Transaction) bool {
	// 从tx.From恢复公钥（这里假设From是hex.Marshal(pub)）
	pubBytes, err := hex.DecodeString(tx.From)
	if err != nil {
		return false
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil {
		return false
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	sigBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}
	h := HashTransaction(tx)
	return ecdsa.VerifyASN1(&pub, h, sigBytes)
}

// ===== 区块与 PoW =====
// CalculateHash 计算区块的哈希值
func CalculateHash(b Block) string {
	// 注意：不把Hash字段本身参与哈希
	txBytes, _ := json.Marshal(b.Transactions)
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) + string(txBytes) + b.PrevHash + strconv.FormatInt(b.Nonce, 10)
	h := sha256.Sum256([]byte(record))
	return fmt.Sprintf("%x", h)
}

// MineBlock 挖掘新区块，通过工作量证明找到满足难度要求的哈希值
func MineBlock(transactions []Transaction, prev Block) Block {
	newBlock := Block{
		Index:        prev.Index + 1,        // 新区块索引为前一区块索引+1
		Timestamp:    time.Now().Unix(),     // 设置当前时间戳
		Transactions: transactions,          // 设置要打包的交易
		PrevHash:     prev.Hash,             // 设置前一区块哈希
		Nonce:        0,                     // 初始化随机数为0
		Hash:         "",                    // 初始化哈希为空
	}
	// 不断尝试不同的Nonce值直到找到满足难度要求的哈希
	for {
		newBlock.Hash = CalculateHash(newBlock)
		// 检查哈希是否满足难度要求（以指定数量的0开头）
		if strings.HasPrefix(newBlock.Hash, strings.Repeat("0", difficulty)) {
			break
		}
		newBlock.Nonce++ // 增加Nonce值继续尝试
	}
	return newBlock
}

// ===== 链操作 =====
// InitGenesis 初始化创世区块
func InitGenesis() {
	genesis := Block{
		Index:        0,                     // 创世区块索引为0
		Timestamp:    time.Now().Unix(),     // 设置当前时间戳
		Transactions: []Transaction{},       // 创世区块不包含交易
		PrevHash:     "0",                   // 前一区块哈希为"0"
		Nonce:        0,                     // 随机数初始化为0
	}
	genesis.Hash = CalculateHash(genesis)    // 计算创世区块哈希
	blockchain = []Block{genesis}           // 将创世区块加入区块链
}

// AddBlock 向区块链添加新区块
func AddBlock(b Block) bool {
	chainMutex.Lock()                       // 加锁保护区块链数据
	defer chainMutex.Unlock()               // 函数结束时解锁

	last := blockchain[len(blockchain)-1]   // 获取最后一个区块
	// 验证区块有效性：
	// 1. 前一区块哈希必须匹配
	if b.PrevHash != last.Hash {
		return false
	}
	// 2. 区块哈希必须正确
	if CalculateHash(b) != b.Hash {
		return false
	}
	// 3. 区块哈希必须满足难度要求（简单验证PoW）
	if !strings.HasPrefix(b.Hash, strings.Repeat("0", difficulty)) {
		return false
	}
	blockchain = append(blockchain, b)      // 将新区块添加到区块链末尾
	return true
}

// ReplaceChain 用更长的链替换当前链（共识机制的一部分）
func ReplaceChain(newChain []Block) {
	chainMutex.Lock()                       // 加锁保护区块链数据
	defer chainMutex.Unlock()               // 函数结束时解锁
	// 只有当新链比当前链更长时才替换
	if len(newChain) > len(blockchain) {
		blockchain = newChain
	}
}

// ===== P2P 简单实现（基于 TCP） =====
// startServer 启动TCP服务器监听指定地址
func startServer(listenAddr string) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	log.Printf("Listening on %s", listenAddr)
	// 循环接受客户端连接
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		// 为每个连接启动一个goroutine处理
		go handleConn(conn)
	}
}

// handleConn 处理网络连接上的消息
func handleConn(conn net.Conn) {
	defer conn.Close()                      // 函数结束时关闭连接
	reader := bufio.NewReader(conn)
	// 循环读取并处理消息
	for {
		raw, err := reader.ReadBytes('\n')   // 读取一行数据（以\n结尾）
		if err != nil {
			if err != io.EOF {
				log.Println("read error:", err)
			}
			return
		}
		var msg Message
		// 反序列化消息
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Println("invalid message:", err)
			continue
		}
		// 根据消息类型进行不同处理
		switch msg.Type {
		case "TX":
			var tx Transaction
			// 反序列化交易数据
			if err := json.Unmarshal(msg.Data, &tx); err == nil {
				handleTx(tx)                // 处理接收到的交易
			}
		case "BLOCK":
			var b Block
			// 反序列化区块数据
			if err := json.Unmarshal(msg.Data, &b); err == nil {
				// 尝试添加区块
				if AddBlock(b) {
					log.Println("Added block from peer:", b.Index)
					// 清除已包含的交易
					removeTxs(b.Transactions)
				} else {
					log.Println("Received invalid block")
				}
			}
		case "GETCHAIN":
			sendChain(conn)                 // 发送本地区块链数据
		case "CHAIN":
			var chain []Block
			// 反序列化区块链数据
			if err := json.Unmarshal(msg.Data, &chain); err == nil {
				ReplaceChain(chain)         // 替换本地区块链（如果更长）
			}
		default:
			// 忽略未知类型的消息
		}
	}
}

// sendChain 向连接发送本地区块链数据
func sendChain(w io.Writer) {
	chainMutex.Lock()                       // 加锁保护区块链数据
	defer chainMutex.Unlock()               // 函数结束时解锁
	data, _ := json.Marshal(blockchain)     // 序列化本地区块链
	msg := Message{Type: "CHAIN", Data: data} // 构造响应消息
	out, _ := json.Marshal(msg)
	out = append(out, '\n')                 // 添加换行符作为分隔符
	w.Write(out)                            // 发送数据
}

// broadcastMessage 广播消息给所有邻居节点
func broadcastMessage(msg Message) {
	out, _ := json.Marshal(msg)             // 序列化消息
	out = append(out, '\n')                 // 添加换行符作为分隔符
	// 遍历所有邻居节点
	for _, p := range peers {
		// 为每个节点启动一个goroutine进行异步发送
		go func(peer string) {
			conn, err := net.Dial("tcp", peer) // 连接到节点
			if err != nil {
				// 连接失败时记录日志但不中断其他节点的发送
				// log.Printf("could not connect %s: %v", peer, err)
				return
			}
			defer conn.Close()              // 函数结束时关闭连接
			conn.Write(out)                 // 发送消息
		}(p)
	}
}

// ===== 交易处理 =====
// handleTx 处理接收到的交易
func handleTx(tx Transaction) {
	// 首先验证交易签名的有效性
	if !VerifyTransaction(tx) {
		log.Println("Invalid tx signature")
		return
	}
	txPoolMutex.Lock()                      // 加锁保护交易池
	defer txPoolMutex.Unlock()              // 函数结束时解锁
	// 简单去重：用signature判断交易是否已在交易池中
	for _, t := range txPool {
		if t.Signature == tx.Signature {
			return
		}
	}
	txPool = append(txPool, tx)             // 将交易添加到交易池
	log.Println("Accepted tx into pool. Pool size:", len(txPool))
	// 广播给其他节点
	data, _ := json.Marshal(tx)
	broadcastMessage(Message{Type: "TX", Data: data})
}

// removeTxs 从交易池中移除已打包的交易
func removeTxs(txs []Transaction) {
	txPoolMutex.Lock()                      // 加锁保护交易池
	defer txPoolMutex.Unlock()              // 函数结束时解锁
	newPool := []Transaction{}              // 创建新的交易池

	// 遍历当前交易池中的所有交易
	for _, p := range txPool {
		found := false
		// 检查该交易是否在要移除的交易列表中
		for _, t := range txs {
			if p.Signature == t.Signature {
				found = true
				break
			}
		}
		// 如果不在要移除的列表中，则保留在新交易池中
		if !found {
			newPool = append(newPool, p)
		}
	}
	txPool = newPool                        // 更新交易池
}

// ===== 客户端命令行（交互） =====
// printChain 打印当前区块链信息
func printChain() {
	chainMutex.Lock()                       // 加锁保护区块链数据
	defer chainMutex.Unlock()               // 函数结束时解锁
	fmt.Println("=== Blockchain ===")
	// 遍历并打印每个区块的信息
	for _, b := range blockchain {
		fmt.Printf("Index:%d Time:%d PrevHash:%s Hash:%s TxCount:%d Nonce:%d\n", b.Index, b.Timestamp, b.PrevHash[:8], b.Hash[:8], len(b.Transactions), b.Nonce)
	}
	fmt.Println("==================")
}

// mineRoutine 挖矿例程，持续挖掘新区块
func mineRoutine(priv *ecdsa.PrivateKey) {
	for {
		// 取出当前交易池
		txPoolMutex.Lock()                  // 加锁访问交易池
		// 如果交易池为空则等待
		if len(txPool) == 0 {
			txPoolMutex.Unlock()
			time.Sleep(2 * time.Second)     // 等待2秒后重试
			continue
		}
		// 复制当前交易池中的所有交易
		txs := make([]Transaction, len(txPool))
		copy(txs, txPool)
		txPoolMutex.Unlock()                // 解锁交易池

		chainMutex.Lock()                   // 加锁访问区块链
		last := blockchain[len(blockchain)-1] // 获取最新的区块
		chainMutex.Unlock()                 // 解锁区块链

		log.Println("Start mining block with", len(txs), "txs...")
		// 挖掘包含这些交易的新区块
		newB := MineBlock(txs, last)
		// 尝试将新区块添加到区块链
		if AddBlock(newB) {
			log.Println("Mined new block:", newB.Index, newB.Hash[:10])
			// 从交易池中移除已打包的交易并广播新区块
			removeTxs(txs)
			data, _ := json.Marshal(newB)
			broadcastMessage(Message{Type: "BLOCK", Data: data})
		}
	}
}

// ===== 启动 & 辅助 =====
// requestChainsFromPeers 向所有邻居节点请求区块链数据以同步
func requestChainsFromPeers() {
	// 遍历所有邻居节点
	for _, p := range peers {
		// 为每个节点启动一个goroutine进行异步请求
		go func(peer string) {
			conn, err := net.Dial("tcp", peer) // 连接到节点
			if err != nil {
				return
			}
			defer conn.Close()              // 函数结束时关闭连接
			// 构造GETCHAIN请求消息
			msg := Message{Type: "GETCHAIN", Data: nil}
			out, _ := json.Marshal(msg)
			out = append(out, '\n')         // 添加换行符作为分隔符
			conn.Write(out)                 // 发送请求
		}(p)
	}
}

// main 主函数，程序入口点
func main() {
	// 检查命令行参数，至少需要提供端口号
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run mini_chain.go <port> [peer1,peer2,..]")
		return
	}
	port := os.Args[1]                      // 获取端口号
	addr = "localhost:" + port              // 构造本节点地址
	// 如果提供了邻居节点参数，则解析它们
	if len(os.Args) >= 3 {
		peers = strings.Split(os.Args[2], ",")
	}
	InitGenesis()                           // 初始化创世区块

	// 生成本地钱包（用于演示）
	priv, pubAddr := NewKeyPair()
	fmt.Println("Your wallet address (From):", pubAddr)

	// 启动服务器监听连接
	go startServer(addr)

	// 启动矿工线程（后台自动挖矿）
	go mineRoutine(priv)

	// 向邻居节点请求区块链数据以同步
	time.Sleep(500 * time.Millisecond)
	requestChainsFromPeers()

	// 启动命令行交互界面
	reader := bufio.NewReader(os.Stdin)
	for {
		// 显示可用命令
		fmt.Println("Commands: tx <to> <amount> | chain | pool | peers | addpeer <host:port> | exit")
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')  // 读取用户输入
		line = strings.TrimSpace(line)      // 去除首尾空格
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")   // 分割命令和参数
		// 根据命令执行相应操作
		switch parts[0] {
		case "tx":
			// 发起交易命令：tx <接收地址> <金额>
			if len(parts) < 3 {
				fmt.Println("usage: tx <to> <amount>")
				continue
			}
			to := parts[1]                  // 接收方地址
			amt, _ := strconv.Atoi(parts[2]) // 转账金额
			// 创建交易
			tx := Transaction{From: pubAddr, To: to, Amount: amt}
			// 对交易签名
			sig, err := SignTransaction(priv, tx)
			if err != nil {
				fmt.Println("sign err:", err)
				continue
			}
			tx.Signature = sig              // 设置签名
			handleTx(tx)                    // 处理该交易（本地处理并广播）
			fmt.Println("Broadcasted tx")
		case "chain":
			printChain()                    // 打印区块链信息
		case "pool":
			// 打印交易池信息
			txPoolMutex.Lock()
			fmt.Println("Pending txs:", len(txPool))
			for i, t := range txPool {
				fmt.Printf("%d: %s -> %s : %d sig:%s\n", i, t.From[:10], t.To[:10], t.Amount, t.Signature[:10])
			}
			txPoolMutex.Unlock()
		case "peers":
			fmt.Println("Peers:", peers)    // 打印邻居节点列表
		case "addpeer":
			// 添加新的邻居节点：addpeer <地址>
			if len(parts) < 2 {
				fmt.Println("usage: addpeer host:port")
				continue
			}
			peers = append(peers, parts[1]) // 添加到邻居节点列表
			fmt.Println("Added peer", parts[1])
		case "exit":
			return                          // 退出程序
		default:
			fmt.Println("unknown command")
		}
	}
}