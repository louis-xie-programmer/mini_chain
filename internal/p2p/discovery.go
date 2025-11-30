package p2p

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// 定义用于局域网节点发现的标识字符串
const rendezvous = "mini-chain"

// Notifee 处理新发现节点的结构体
type Notifee struct {
	h host.Host // 主机实例
}

// HandlePeerFound 当发现新节点时调用的处理函数
// pi: 新发现的节点地址信息
func (n *Notifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("Discovered new peer:", pi.ID.String(), pi.Addrs)
	// 可以直接连接到新发现的节点
	n.h.Connect(context.Background(), pi)
}

// SetupMdns 启动本地mDNS服务用于局域网节点发现
// ctx: 上下文
// h: 主机实例
func SetupMdns(ctx context.Context, h host.Host) error {
	// 创建Notifee实例
	n := &Notifee{h: h}
	// 创建mDNS服务实例
	service := mdns.NewMdnsService(h, rendezvous, n)
	// 服务会在后台运行
	_ = service
	return nil
}