package p2p

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Node struct {
	Host   host.Host
	PubSub *pubsub.PubSub
	Topic  *pubsub.Topic
	Sub    *pubsub.Subscription
}

// NewNode 创建 libp2p 节点并初始化 gossipsub
func NewNode(ctx context.Context, listenPort int) (*Node, error) {
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			// 支持 TCP + 随机端口
			// "/ip4/0.0.0.0/tcp/<port>"
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
		),
	)
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	topic, err := ps.Join("mini-chain")
	if err != nil {
		return nil, err
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	node := &Node{
		Host:   h,
		PubSub: ps,
		Topic:  topic,
		Sub:    sub,
	}

	// 启动 mDNS
	if err := SetupMdns(ctx, h); err != nil {
		log.Println("mDNS warning:", err)
	}

	// 异步接收消息
	go node.handleMessages(ctx)

	return node, nil
}

// Broadcast 广播消息
func (n *Node) Broadcast(msg *Message) {
	data, _ := msg.Encode()
	if err := n.Topic.Publish(context.Background(), data); err != nil {
		log.Println("Failed to broadcast:", err)
	}
}

// handleMessages 循环接收 gossipsub 消息
func (n *Node) handleMessages(ctx context.Context) {
	for {
		msg, err := n.Sub.Next(ctx)
		if err != nil {
			return
		}
		if msg.ReceivedFrom == n.Host.ID() {
			continue // 忽略自己发送的消息
		}
		m, err := Decode(msg.Data)
		if err != nil {
			log.Println("invalid message:", err)
			continue
		}
		// TODO: 根据 Type 调用 blockchain / txpool 等处理函数
		log.Println("Received msg from", msg.ReceivedFrom, "type:", m.Type)
	}
}

// ConnectPeer 手动连接到指定 peer
func (n *Node) ConnectPeer(addr string) error {
	pi, err := peer.AddrInfoFromString(addr)
	if err != nil {
		return err
	}
	return n.Host.Connect(context.Background(), *pi)
}
