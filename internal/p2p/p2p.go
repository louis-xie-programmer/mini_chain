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

// Node 表示一个libp2p节点，包含主机、发布订阅和主题相关信息
type Node struct {
	Host   host.Host      // libp2p主机实例
	PubSub *pubsub.PubSub // 发布订阅实例
	Topic  *pubsub.Topic  // 主题实例
	Sub    *pubsub.Subscription // 订阅实例
}

// NewNode 创建libp2p节点并初始化gossipsub
// ctx: 上下文
// listenPort: 监听端口
func NewNode(ctx context.Context, listenPort int) (*Node, error) {
	// 创建libp2p主机实例
	h, err := libp2p.New(
		libp2p.ListenAddrStrings(
			// 支持TCP + 随机端口
			// "/ip4/0.0.0.0/tcp/<port>"
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建GossipSub实例
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, err
	}

	// 加入"mini-chain"主题
	topic, err := ps.Join("mini-chain")
	if err != nil {
		return nil, err
	}

	// 订阅主题
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	// 创建节点实例
	node := &Node{
		Host:   h,
		PubSub: ps,
		Topic:  topic,
		Sub:    sub,
	}

	// 启动mDNS服务用于局域网节点发现
	if err := SetupMdns(ctx, h); err != nil {
		log.Println("mDNS warning:", err)
	}

	// 异步接收消息
	go node.handleMessages(ctx)

	return node, nil
}

// Broadcast 广播消息到网络
// msg: 要广播的消息
func (n *Node) Broadcast(msg *Message) {
	data, _ := msg.Encode() // 编码消息
	// 发布消息到主题
	if err := n.Topic.Publish(context.Background(), data); err != nil {
		log.Println("Failed to broadcast:", err)
	}
}

// handleMessages 循环接收gossipsub消息
// ctx: 上下文
func (n *Node) handleMessages(ctx context.Context) {
	for {
		// 获取下一条消息
		msg, err := n.Sub.Next(ctx)
		if err != nil {
			return
		}
		// 忽略自己发送的消息
		if msg.ReceivedFrom == n.Host.ID() {
			continue
		}
		// 解码消息
		m, err := Decode(msg.Data)
		if err != nil {
			log.Println("invalid message:", err)
			continue
		}
		// TODO: 根据消息类型调用blockchain/txpool等处理函数
		log.Println("Received msg from", msg.ReceivedFrom, "type:", m.Type)
	}
}

// ConnectPeer 手动连接到指定的peer
// addr: peer地址字符串
func (n *Node) ConnectPeer(addr string) error {
	// 从地址字符串解析peer地址信息
	pi, err := peer.AddrInfoFromString(addr)
	if err != nil {
		return err
	}
	// 连接到peer
	return n.Host.Connect(context.Background(), *pi)
}