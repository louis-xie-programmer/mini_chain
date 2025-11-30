package main

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"mini_chain/gossip/core"

	libp2p "github.com/libp2p/go-libp2p"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	network "github.com/libp2p/go-libp2p/core/network"
	peer "github.com/libp2p/go-libp2p/core/peer"
	peerstore "github.com/libp2p/go-libp2p/core/peerstore"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

const (
	ProtocolID       = "/mini-chain/1.0.0"
	GossipTopic      = "mini-chain-gossip-v1"
	RendezvousString = "mini-chain-rendezvous"
)

// Transaction 交易结构体，表示一笔转账交易
type Transaction struct {
	From      string `json:"from"`      // 发送方地址
	To        string `json:"to"`        // 接收方地址
	Amount    int    `json:"amount"`    // 转账金额
	Signature string `json:"signature"` // 交易签名，用于验证交易有效性
}

// Block 区块结构体，包含区块的所有信息
type Block struct {
	Index        int           `json:"index"`        // 区块索引（高度）
	Timestamp    int64         `json:"timestamp"`    // 时间戳
	Transactions []Transaction `json:"transactions"` // 包含的交易列表
	PrevHash     string        `json:"prev_hash"`    // 前一个区块的哈希值
	Nonce        int64         `json:"nonce"`        // 工作量证明的随机数
	Hash         string        `json:"hash"`         // 当前区块的哈希值
}

// Message 网络消息结构体，用于节点间通信
type Message struct {
	Type string          `json:"type"` // "TX", "BLOCK", "GETCHAIN", "CHAIN"
	Data json.RawMessage `json:"data"`
}

var (
	blockchain  *core.Blockchain   // 使用core包中的Blockchain类型
	txPool      []core.Transaction // 使用core包中的Transaction类型
	chainMutex  sync.Mutex
	txPoolMutex sync.Mutex
	// difficulty  = 3  // 不再需要，因为core包中已经定义

	h           host.Host
	ctx         context.Context
	cancel      context.CancelFunc
	knownPeers  = make(map[peer.ID]struct{})
	knownPeersM sync.Mutex

	gsub  *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription
)

// --- Wallet / TX utils ---
// 移除了core包中已实现的函数：NewKeyPair, HashTransaction, SignTransaction, VerifyTransaction

// --- Block / PoW ---
// 移除了core包中已实现的函数：CalculateHash, MineBlock

// --- Chain operations ---
// 移除了core包中已实现的函数：InitGenesis, AddBlock, ReplaceChain

// --- TX pool ---
func handleTx(tx core.Transaction) { // 使用core.Transaction类型
	if !core.VerifyTransaction(tx) {
		log.Println("Invalid tx signature for tx from:", tx.From[:8], "to:", tx.To[:8], "amount:", tx.Amount)
		return
	}
	txPoolMutex.Lock()
	defer txPoolMutex.Unlock()
	for _, t := range txPool {
		if t.Signature == tx.Signature {
			log.Println("Transaction already in pool")
			return
		}
	}
	txPool = append(txPool, tx)
	publish(Message{Type: "TX", Data: mustMarshal(tx)})
	log.Println("Accepted tx into pool. Pool size:", len(txPool))
}

func removeTxs(txs []core.Transaction) { // 使用core.Transaction类型
	txPoolMutex.Lock()
	defer txPoolMutex.Unlock()
	newPool := []core.Transaction{}
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

// --- gossipsub ---
func mustMarshal(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func publish(msg Message) {
	if topic == nil {
		return
	}
	b, _ := json.Marshal(msg)
	ctxPub, cancelPub := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPub()
	topic.Publish(ctxPub, b)
}

func subLoop() {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(time.Second)
			continue
		}
		if msg.ReceivedFrom == h.ID() {
			continue
		}
		var m Message
		if err := json.Unmarshal(msg.Data, &m); err != nil {
			continue
		}
		switch m.Type {
		case "TX":
			var tx core.Transaction
			if err := json.Unmarshal(m.Data, &tx); err == nil {
				log.Println("Received TX message via gossip from:", msg.ReceivedFrom)
				handleTx(tx)
			} else {
				log.Println("Failed to unmarshal transaction:", err)
			}
		case "BLOCK":
			var b core.Block
			if err := json.Unmarshal(m.Data, &b); err == nil {
				if AddBlock(b) {
					removeTxs(b.Transactions)
				} else {
					go requestChainFrom(msg.ReceivedFrom)
				}
			}
		}
	}
}

// --- stream chain sync ---
func setStreamHandler() {
	h.SetStreamHandler(ProtocolID, func(s network.Stream) {
		defer s.Close()
		r := bufio.NewReader(s)
		for {
			raw, err := r.ReadBytes('\n')
			if err != nil {
				return
			}
			var msg Message
			if err := json.Unmarshal(raw, &msg); err != nil {
				continue
			}
			switch msg.Type {
			case "GETCHAIN":
				chainMutex.Lock()
				// Get blocks from the core blockchain
				blocks := blockchain.GetBlocks()
				data, _ := json.Marshal(blocks)
				chainMutex.Unlock()
				resp := Message{Type: "CHAIN", Data: data}
				out, _ := json.Marshal(resp)
				out = append(out, '\n')
				s.Write(out)
			}
		}
	})
}

func requestChainFrom(pid peer.ID) {
	s, err := h.NewStream(ctx, pid, ProtocolID)
	if err != nil {
		return
	}
	defer s.Close()
	msg := Message{Type: "GETCHAIN"}
	data, _ := json.Marshal(msg)
	data = append(data, '\n')
	s.Write(data)
	r := bufio.NewReader(s)
	respRaw, err := r.ReadBytes('\n')
	if err != nil {
		return
	}
	var resp Message
	if err := json.Unmarshal(respRaw, &resp); err != nil {
		return
	}
	if resp.Type != "CHAIN" {
		return
	}
	var newChain []core.Block
	if err := json.Unmarshal(resp.Data, &newChain); err != nil {
		return
	}
	ReplaceChain(newChain)
	log.Println("Chain synchronized from peer:", pid.String())
}

// --- known peers ---
func addKnownPeer(pid peer.ID) {
	knownPeersM.Lock()
	defer knownPeersM.Unlock()
	if _, ok := knownPeers[pid]; !ok {
		knownPeers[pid] = struct{}{}
	}
}

// --- mDNS ---
type mdnsNotifee struct{ h host.Host }

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	m.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.h.Connect(ctx2, pi); err == nil {
		log.Println("Connected to peer via mDNS:", pi.ID.String())
		addKnownPeer(pi.ID)
	}
}

func setupMdns() {
	mdnsSvc := mdns.NewMdnsService(h, RendezvousString, &mdnsNotifee{h: h})
	_ = mdnsSvc
	log.Println("mDNS service started")
}

// --- DHT ---
func setupDHTAndBootstrap(enable bool) (*kaddht.IpfsDHT, error) {
	if !enable {
		return nil, nil
	}
	dht, err := kaddht.New(ctx, h)
	if err != nil {
		return nil, err
	}
	for _, addr := range kaddht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}
		h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		h.Connect(ctx2, *pi)
		cancel()
	}
	dht.Bootstrap(ctx)
	routingDiscovery := discovery.NewRoutingDiscovery(dht)
	go func() {
		for {
			peerChan, _ := routingDiscovery.FindPeers(ctx, RendezvousString)
			for pi := range peerChan {
				if pi.ID == h.ID() {
					continue
				}
				h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
				ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				h.Connect(ctx2, pi)
				cancel()
				addKnownPeer(pi.ID)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	return dht, nil
}

// --- miner ---
func mineRoutine(priv *ecdsa.PrivateKey) {
	for {
		txPoolMutex.Lock()
		if len(txPool) == 0 {
			txPoolMutex.Unlock()
			time.Sleep(2 * time.Second)
			continue
		}
		txs := make([]core.Transaction, len(txPool))
		copy(txs, txPool)
		txPoolMutex.Unlock()

		// Get the last block from the blockchain
		blocks := blockchain.GetBlocks()
		last := blocks[len(blocks)-1]

		newB := core.MineBlock(txs, last)
		if AddBlock(newB) {
			removeTxs(txs)
			publish(Message{Type: "BLOCK", Data: mustMarshal(newB)})
			log.Println("Mined block:", newB.Index, newB.Hash[:10])
		}
	}
}

// AddBlock 向区块链添加新区块
func AddBlock(b core.Block) bool {
	return blockchain.AddBlock(b)
}

// ReplaceChain 用更长的链替换当前链（共识机制的一部分）
func ReplaceChain(newChain []core.Block) {
	// For this simplified implementation, we're just logging
	// A real implementation would validate and replace the chain in the core blockchain
	log.Println("Received new chain with", len(newChain), "blocks")
	// In a full implementation, we would update the core blockchain here
}

// --- CLI helpers ---
func printChain() {
	chainMutex.Lock()
	defer chainMutex.Unlock()
	fmt.Println("=== Blockchain ===")

	// Get blocks from the core blockchain
	blocks := blockchain.GetBlocks()
	for _, b := range blocks {
		fmt.Printf("Index:%d Hash:%s Prev:%s Tx:%d\n", b.Index, b.Hash[:8], b.PrevHash[:8], len(b.Transactions))
	}
}

func printPeers() {
	knownPeersM.Lock()
	defer knownPeersM.Unlock()
	fmt.Println("Known peers:")
	for p := range knownPeers {
		fmt.Println(" -", p.String())
	}
}

// --- main ---
func main() {
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run mini_chain_gossip_stream_mdns.go <port>")
	}

	blockchain = core.NewBlockchain()  // 使用core包中的NewBlockchain函数
	priv, pubAddr := core.NewKeyPair() // 使用core包中的NewKeyPair函数
	fmt.Println("Wallet address:", pubAddr)

	var lpHost host.Host
	var err error
	p := "0"
	if len(os.Args) >= 2 {
		p = os.Args[1]
	}
	lpHost, err = libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/"+p),
		libp2p.NATPortMap(),
		libp2p.EnableRelay())
	if err != nil {
		log.Fatal(err)
	}
	h = lpHost

	fmt.Println("Host ID:", h.ID().String())
	for _, a := range h.Addrs() {
		fmt.Println("Addr:", a.String()+"/p2p/"+h.ID().String())
	}

	setStreamHandler()

	gsub, err = pubsub.NewGossipSub(ctx, h)
	if err != nil {
		log.Fatal(err)
	}

	topic, err = gsub.Join(GossipTopic)
	if err != nil {
		log.Fatal(err)
	}

	sub, err = topic.Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	go subLoop()

	setupMdns()
	setupDHTAndBootstrap(true)

	go mineRoutine(priv)

	time.Sleep(1 * time.Second)

	for _, pid := range h.Peerstore().Peers() {
		if pid == h.ID() {
			continue
		}
		go requestChainFrom(pid)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, " ", 3)

		switch parts[0] {
		case "tx":
			if len(parts) < 3 {
				fmt.Println("usage: tx <to> <amount>")
				continue
			}
			to := parts[1]
			amt, _ := strconv.Atoi(parts[2])

			tx := core.Transaction{From: pubAddr, To: to, Amount: amt}
			sig, _ := core.SignTransaction(priv, tx)
			tx.Signature = sig

			handleTx(tx)
		case "chain":
			printChain()
		case "peers":
			printPeers()
		case "exit":
			return
		}
	}
}
