package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mini_chain/internal/api"
	"mini_chain/internal/blockchain"
	"mini_chain/internal/p2p"
	"os"
	"strconv"
	"strings"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <p2p_port> [api_port] [bootstrap_peers]")
		fmt.Println("Example: go run main.go 3000 8080 /ip4/127.0.0.1/tcp/3001/p2p/QmPeerId")
		os.Exit(1)
	}
	
	// 从命令行解析P2P端口
	p2pPort, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid P2P port:", err)
	}
	
	// 从命令行解析API端口（默认为8080）
	apiPort := 8080
	if len(os.Args) >= 3 {
		apiPort, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal("Invalid API port:", err)
		}
	}
	
	// 解析引导节点（可选）
	var bootstrapPeers []string
	if len(os.Args) >= 4 {
		bootstrapPeers = strings.Split(os.Args[3], ",")
	}

	ctx := context.Background()
	// 1️⃣ 启动区块链，默认难度为3
	bc := blockchain.NewBlockchain(3)

	// 2️⃣ 启动libp2p节点，P2P端口通过命令行传值
	node, err := p2p.NewNode(ctx, p2pPort)
	if err != nil {
		log.Fatal(err)
	}

	// 连接到引导节点（如果提供了的话）
	for _, addr := range bootstrapPeers {
		if err := node.ConnectPeer(addr); err != nil {
			log.Printf("Failed to connect to bootstrap peer %s: %v", addr, err)
		} else {
			log.Printf("Connected to bootstrap peer: %s", addr)
		}
	}

	// 3️⃣ 启动REST + WebSocket API，API端口通过命令行传值
	apiSrv := api.NewAPI(bc, node)
	go apiSrv.Run(fmt.Sprintf(":%d", apiPort))

	// 打印节点信息
	fmt.Printf("Node ID: %s\n", node.Host.ID().String())
	for _, addr := range node.Host.Addrs() {
		fmt.Printf("Node address: %s/p2p/%s\n", addr.String(), node.Host.ID().String())
	}

	// 4️⃣ 启动挖矿协程
	go mineRoutine(bc, node)

	// 阻塞主线程
	select {}
}

// mineRoutine 挖矿例程，持续挖掘新区块
// bc: 区块链实例
// node: P2P节点实例
func mineRoutine(bc *blockchain.Blockchain, node *p2p.Node) {
	for {
		// 尝试挖取包含内存池交易的新区块
		newBlock, err := bc.MinePending()
		if err != nil {
			// 如果没有交易可挖，等待一段时间再试
			continue
		}
		
		// 验证并应用新区块
		if err := bc.ValidateAndApplyBlock(newBlock); err != nil {
			log.Printf("Failed to validate and apply block: %v", err)
			continue
		}
		
		// 通过P2P网络传播新区块
		msg := &p2p.Message{
			Type: p2p.MsgBlock,
			Data: mustMarshal(newBlock),
		}
		node.Broadcast(msg)
		
		log.Printf("Mined new block: %s", newBlock.Hash)
	}
}

// mustMarshal 简化的序列化函数
// 在实际实现中应该处理错误
// v: 待序列化的对象
func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}