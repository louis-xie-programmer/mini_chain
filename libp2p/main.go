// 实现了一个基于libp2p的简单区块链网络
// 使用方法:
// go run . 3000
// go run . 3001
// go run . 3002
package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	ma "github.com/multiformats/go-multiaddr"
)

// ProtocolID 定义了本地区块链网络使用的协议标识符
const ProtocolID = "/mini-chain/1.0.0"

// RendezvousString 用于本地网络发现的标识字符串
const RendezvousString = "mini-chain-mdns"

// Transaction 表示一笔交易
type Transaction struct {
	From      string `json:"from"`      // 发送方地址
	To        string `json:"to"`        // 接收方地址
	Amount    int    `json:"amount"`    // 交易金额
	Signature string `json:"signature"` // 交易签名
}

// Block 表示一个区块
type Block struct {
	Index        int           `json:"index"`        // 区块索引
	Timestamp    int64         `json:"timestamp"`    // 时间戳
	Transactions []Transaction `json:"transactions"` // 包含的交易列表
	PrevHash     string        `json:"prev_hash"`    // 前一区块哈希值
	Nonce        int64         `json:"nonce"`        // 工作量证明随机数
	Hash         string        `json:"hash"`         // 当前区块哈希值
}

// Message 是节点间通信的消息结构
type Message struct {
	Type string          `json:"type"` // 消息类型: "TX"(交易), "BLOCK"(区块), "GETCHAIN"(获取链), "CHAIN"(链数据)
	Data json.RawMessage `json:"data"` // 消息数据
}

// Global chain & mempool 全局区块链和交易池变量
var (
	blockchain  []Block       // 区块链
	txPool      []Transaction // 交易池（内存池）
	chainMutex  sync.Mutex    // 区块链访问互斥锁
	txPoolMutex sync.Mutex    // 交易池访问互斥锁
	difficulty  = 3           // 挖矿难度（哈希值前缀0的个数）
)

// libp2p related libp2p相关变量
var (
	h           host.Host                    // libp2p主机实例
	ctx         context.Context              // 上下文
	cancel      context.CancelFunc           // 取消函数
	knownPeers  = make(map[peer.ID]struct{}) // 已知节点集合
	knownPeersM sync.Mutex                   // 已知节点集合访问互斥锁
)

// ===== Wallet & Signature Utils 钱包与签名工具函数 =====

// NewKeyPair 生成新的ECDSA密钥对，并返回私钥和公钥地址
func NewKeyPair() (*ecdsa.PrivateKey, string) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pubBytes := elliptic.Marshal(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	return priv, hex.EncodeToString(pubBytes)
}

// HashTransaction 计算交易的哈希值
func HashTransaction(tx Transaction) []byte {
	data := tx.From + "|" + tx.To + "|" + strconv.Itoa(tx.Amount)
	h := sha256.Sum256([]byte(data))
	return h[:]
}

// SignTransaction 对交易进行签名
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

// ===== Block & PoW 区块与工作量证明相关函数 =====

// CalculateHash 计算区块的哈希值
func CalculateHash(b Block) string {
	txBytes, _ := json.Marshal(b.Transactions)
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) + string(txBytes) + b.PrevHash + strconv.FormatInt(b.Nonce, 10)
	h := sha256.Sum256([]byte(record))
	return fmt.Sprintf("%x", h)
}

// MineBlock 挖掘新区块（执行工作量证明）
func MineBlock(transactions []Transaction, prev Block) Block {
	newBlock := Block{
		Index:        prev.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prev.Hash,
		Nonce:        0,
		Hash:         "",
	}
	// 不断尝试不同的nonce值直到找到满足难度要求的哈希值
	for {
		newBlock.Hash = CalculateHash(newBlock)
		// 检查哈希值是否满足难度要求（以指定数量的0开头）
		if strings.HasPrefix(newBlock.Hash, strings.Repeat("0", difficulty)) {
			break
		}
		newBlock.Nonce++
	}
	return newBlock
}

// ===== Chain operations 区块链操作函数 =====

// InitGenesis 初始化创世区块
func InitGenesis() {
	genesis := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     "0",
		Nonce:        0,
	}
	genesis.Hash = CalculateHash(genesis)
	blockchain = []Block{genesis}
}

// AddBlock 向区块链中添加新区块
func AddBlock(b Block) bool {
	chainMutex.Lock()
	defer chainMutex.Unlock()

	last := blockchain[len(blockchain)-1]
	// 验证区块的前哈希是否正确
	if b.PrevHash != last.Hash {
		return false
	}
	// 验证区块哈希值是否正确
	if CalculateHash(b) != b.Hash {
		return false
	}
	// 验证工作量证明是否有效
	if !strings.HasPrefix(b.Hash, strings.Repeat("0", difficulty)) {
		return false
	}
	blockchain = append(blockchain, b)
	return true
}

// ReplaceChain 替换本地区块链（当收到更长的有效链时）
func ReplaceChain(newChain []Block) {
	chainMutex.Lock()
	defer chainMutex.Unlock()
	// 只有当新链比当前链更长时才替换
	if len(newChain) > len(blockchain) {
		blockchain = newChain
		log.Println("Replaced chain with longer chain length:", len(blockchain))
	}
}

// ===== tx pool handling 交易池处理函数 =====

// handleTx 处理接收到的交易
func handleTx(tx Transaction) {
	// 首先验证交易签名
	if !VerifyTransaction(tx) {
		log.Println("Invalid tx signature")
		return
	}
	txPoolMutex.Lock()
	defer txPoolMutex.Unlock()
	// 检查交易是否已在交易池中（通过签名去重）
	for _, t := range txPool {
		if t.Signature == tx.Signature {
			return
		}
	}
	// 将交易添加到交易池
	txPool = append(txPool, tx)
	log.Println("Accepted tx into pool. Pool size:", len(txPool))
	// 广播该交易给其他节点
	broadcastMessage(Message{Type: "TX", Data: mustMarshal(tx)})
}

// removeTxs 从交易池中移除指定的交易
func removeTxs(txs []Transaction) {
	txPoolMutex.Lock()
	defer txPoolMutex.Unlock()
	newPool := []Transaction{}
	// 遍历当前交易池，只保留不在待移除列表中的交易
	for _, p := range txPool {
		found := false
		for _, t := range txs {
			if p.Signature == t.Signature {
				found = true
				break
			}
		}
		if !found {
			newPool = append(newPool, p)
		}
	}
	txPool = newPool
}

// ===== libp2p broadcast & send/receive libp2p广播与发送接收函数 =====

// mustMarshal 将接口对象序列化为JSON RawMessage
func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

// sendToPeer 向指定节点发送消息
func sendToPeer(pid peer.ID, msg Message) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctxWithCancel(), 6*time.Second)
	defer cancel()
	// 创建到目标节点的新流
	s, err := h.NewStream(ctx, pid, ProtocolID)
	if err != nil {
		return err
	}
	defer s.Close()
	// 序列化消息并发送
	out, _ := json.Marshal(msg)
	out = append(out, '\n')
	_, err = s.Write(out)
	return err
}

// broadcastMessage 向所有已知节点广播消息
func broadcastMessage(msg Message) {
	knownPeersM.Lock()
	// 获取所有已知节点的副本
	peers := make([]peer.ID, 0, len(knownPeers))
	for p := range knownPeers {
		peers = append(peers, p)
	}
	knownPeersM.Unlock()

	// 并发向每个节点发送消息
	for _, pid := range peers {
		go func(p peer.ID) {
			if err := sendToPeer(p, msg); err != nil {
				// 如果某些节点发送失败可以接受，不影响整体
			}
		}(pid)
	}
}

// sendChainToStreamWriter 将本地区块链数据写入指定的io.Writer（用于响应GETCHAIN请求）
func sendChainToStreamWriter(w io.Writer) {
	chainMutex.Lock()
	defer chainMutex.Unlock()
	// 序列化本地区块链数据
	data, _ := json.Marshal(blockchain)
	msg := Message{Type: "CHAIN", Data: data}
	out, _ := json.Marshal(msg)
	out = append(out, '\n')
	w.Write(out)
}

// requestChainsFromPeers 向所有已知节点请求它们的区块链数据
func requestChainsFromPeers() {
	knownPeersM.Lock()
	// 获取所有已知节点的副本
	peers := make([]peer.ID, 0, len(knownPeers))
	for p := range knownPeers {
		peers = append(peers, p)
	}
	knownPeersM.Unlock()

	// 并发向每个节点发送GETCHAIN请求
	for _, pid := range peers {
		go func(p peer.ID) {
			_ = sendToPeer(p, Message{Type: "GETCHAIN", Data: nil})
		}(pid)
	}
}

// ===== stream handler 流处理器 =====

// setStreamHandler 设置协议流处理器
func setStreamHandler() {
	h.SetStreamHandler(ProtocolID, func(s network.Stream) {
		defer s.Close()
		// 获取远程节点ID并添加到已知节点列表
		remote := s.Conn().RemotePeer()
		addKnownPeer(remote)
		// 创建读取器来读取流数据
		r := bufio.NewReader(s)
		// 循环读取消息
		for {
			raw, err := r.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					// log.Println("stream read err:", err)
				}
				return
			}
			// 解析消息
			var msg Message
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Println("invalid message:", err)
				continue
			}
			// 根据消息类型处理不同逻辑
			switch msg.Type {
			case "TX":
				// 处理交易消息
				var tx Transaction
				if err := json.Unmarshal(msg.Data, &tx); err == nil {
					handleTx(tx)
				}
			case "BLOCK":
				// 处理区块消息
				var b Block
				if err := json.Unmarshal(msg.Data, &b); err == nil {
					if AddBlock(b) {
						log.Println("Added block from peer:", b.Index)
						// 移除已被包含在区块中的交易
						removeTxs(b.Transactions)
					} else {
						log.Println("Received invalid block")
					}
				}
			case "GETCHAIN":
				// 处理获取区块链请求
				sendChainToStreamWriter(s)
			case "CHAIN":
				// 处理区块链数据
				var chain []Block
				if err := json.Unmarshal(msg.Data, &chain); err == nil {
					ReplaceChain(chain)
				}
			default:
				// 忽略未知类型的消息
			}
		}
	})
}

// ===== known peers helpers 已知节点辅助函数 =====

// addKnownPeer 添加已知节点
func addKnownPeer(pid peer.ID) {
	knownPeersM.Lock()
	defer knownPeersM.Unlock()
	// 检查节点是否已经存在于已知节点列表中
	if _, ok := knownPeers[pid]; !ok {
		knownPeers[pid] = struct{}{}
		log.Println("Known peer added:", pid.String())
	}
}

// removeKnownPeer 移除已知节点
func removeKnownPeer(pid peer.ID) {
	knownPeersM.Lock()
	defer knownPeersM.Unlock()
	delete(knownPeers, pid)
}

// ===== mDNS discovery (local LAN discovery) mDNS发现（局域网发现） =====

// mdnsNotifee 实现了mDNS发现的通知接口
type mdnsNotifee struct {
	h host.Host
}

// HandlePeerFound 处理发现的新节点
func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("mDNS found peer:", pi.ID.String(), pi.Addrs)
	// 将发现的节点地址添加到Peerstore
	m.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
	// 创建连接上下文
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	// 连接到发现的节点
	if err := m.h.Connect(ctx2, pi); err != nil {
		// log.Println("mDNS connect err:", err)
	} else {
		// 连接成功后添加到已知节点列表
		addKnownPeer(pi.ID)
	}
}

func setupMdns() {
	mdnsSvc := mdns.NewMdnsService(h, RendezvousString, &mdnsNotifee{h: h})
	_ = mdnsSvc
	log.Println("mDNS service started")
}

// ===== miner routine 挖矿例程 =====

// mineRoutine 挖矿协程函数
func mineRoutine(priv *ecdsa.PrivateKey) {
	for {
		txPoolMutex.Lock()
		// 如果交易池为空，则等待
		if len(txPool) == 0 {
			txPoolMutex.Unlock()
			time.Sleep(2 * time.Second)
			continue
		}
		// 获取交易池中的所有交易副本
		txs := make([]Transaction, len(txPool))
		copy(txs, txPool)
		txPoolMutex.Unlock()

		chainMutex.Lock()
		// 获取最新的区块作为前一个区块
		last := blockchain[len(blockchain)-1]
		chainMutex.Unlock()

		log.Println("Start mining block with", len(txs), "txs...")
		// 开始挖矿（工作量证明）
		newB := MineBlock(txs, last)
		// 如果成功添加新区块
		if AddBlock(newB) {
			log.Println("Mined new block:", newB.Index, newB.Hash[:10])
			// 从交易池中移除已被包含在区块中的交易
			removeTxs(txs)
			// 广播新区块给其他节点
			broadcastMessage(Message{Type: "BLOCK", Data: mustMarshal(newB)})
		}
	}
}

// ===== helper to create a context accessible to sendToPeer 创建可被sendToPeer访问的上下文的辅助函数 =====

// ctxWithCancel 返回全局上下文
func ctxWithCancel() context.Context {
	return ctx
}

// ===== CLI helpers CLI辅助函数 =====

// printChain 打印当前区块链信息
func printChain() {
	chainMutex.Lock()
	defer chainMutex.Unlock()
	fmt.Println("=== Blockchain ===")
	for _, b := range blockchain {
		fmt.Printf("Index:%d Time:%d PrevHash:%s Hash:%s TxCount:%d Nonce:%d\n",
			b.Index, b.Timestamp, shorten(b.PrevHash, 8), shorten(b.Hash, 8), len(b.Transactions), b.Nonce)
	}
	fmt.Println("==================")
}

// shorten 缩短字符串显示长度的辅助函数
func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// printPeers 打印已知节点信息
func printPeers() {
	fmt.Println("Known peers:")
	knownPeersM.Lock()
	defer knownPeersM.Unlock()
	// 打印已知节点列表
	for p := range knownPeers {
		fmt.Println(" -", p.String())
	}
	fmt.Println("Peerstore addrs (sample):")
	// 打印Peerstore中的节点地址信息（示例）
	for _, pid := range h.Peerstore().Peers() {
		addrs := h.Peerstore().Addrs(pid)
		if len(addrs) > 0 {
			fmt.Println(pid.String(), "->", addrs[0])
		}
	}
}

// ===== main 主函数 =====
func main() {
	// 初始化上下文
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run mini_chain_libp2p.go <port> (e.g. 3000)")
		fmt.Println("Optional: you can also connect to a remote peer multiaddr using 'addpeer <multiaddr>' CLI command")
	}

	// 初始化创世区块
	InitGenesis()

	// 创建钱包
	priv, pubAddr := NewKeyPair()
	fmt.Println("Your wallet address (From):", pubAddr)

	// 创建libp2p主机
	var lpHost host.Host
	var err error

	// 根据命令行参数决定监听端口
	if len(os.Args) >= 2 {
		p := os.Args[1]
		maddrStr := "/ip4/0.0.0.0/tcp/" + p
		lpHost, err = libp2p.New(
			libp2p.ListenAddrStrings(maddrStr),
			libp2p.NATPortMap(),
			libp2p.EnableRelay(),
		)
	} else {
		lpHost, err = libp2p.New(
			libp2p.NATPortMap(),
			libp2p.EnableRelay(),
		)
	}
	if err != nil {
		log.Fatal("Failed to create libp2p host:", err)
	}
	h = lpHost

	// 打印主机信息和监听地址
	fmt.Println("Host ID:", h.ID().String())
	fmt.Println("Listening on:")
	for _, a := range h.Addrs() {
		fmt.Println("  ", a.String()+"/p2p/"+h.ID().String())
	}

	// 设置协议流处理器
	setStreamHandler()

	setupMdns()

	// 启动挖矿协程
	go mineRoutine(priv)

	// 延迟一段时间后向已知节点请求区块链数据
	time.AfterFunc(1*time.Second, func() {
		requestChainsFromPeers()
	})

	// 启动交互式命令行界面
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Commands: tx <to> <amount> | chain | pool | peers | addpeer <multiaddr> | exit")
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		// 根据用户输入执行相应命令
		switch parts[0] {
		case "tx":
			// 发起交易命令
			if len(parts) < 3 {
				fmt.Println("usage: tx <to> <amount>")
				continue
			}
			to := parts[1]
			amt, _ := strconv.Atoi(parts[2])
			tx := Transaction{From: pubAddr, To: to, Amount: amt}
			sig, err := SignTransaction(priv, tx)
			if err != nil {
				fmt.Println("sign err:", err)
				continue
			}
			tx.Signature = sig
			handleTx(tx)
			fmt.Println("Broadcasted tx")
		case "chain":
			// 显示区块链命令
			printChain()
		case "pool":
			// 显示交易池命令
			txPoolMutex.Lock()
			fmt.Println("Pending txs:", len(txPool))
			for i, t := range txPool {
				fmt.Printf("%d: %s -> %s : %d sig:%s\n", i, shorten(t.From, 10), shorten(t.To, 10), t.Amount, shorten(t.Signature, 10))
			}
			txPoolMutex.Unlock()
		case "peers":
			// 显示节点信息命令
			printPeers()
		case "addpeer":
			// 手动添加节点命令
			if len(parts) < 2 {
				fmt.Println("usage: addpeer <multiaddr>")
				continue
			}
			addrStr := parts[1]
			maddr, err := ma.NewMultiaddr(addrStr)
			if err != nil {
				fmt.Println("invalid multiaddr:", err)
				continue
			}
			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				fmt.Println("invalid peer addr info:", err)
				continue
			}
			h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.TempAddrTTL)
			ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
			if err := h.Connect(ctx2, *info); err != nil {
				fmt.Println("connect failed:", err)
			} else {
				addKnownPeer(info.ID)
				fmt.Println("connected to", info.ID.String())
			}
			cancel2()
		case "exit":
			// 退出程序命令
			return
		default:
			fmt.Println("unknown command")
		}
	}
}
