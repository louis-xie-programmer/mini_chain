package p2p

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const rendezvous = "mini-chain"

// Notifee 处理新发现节点
type Notifee struct {
	h host.Host
}

// HandlePeerFound 被发现新节点时调用
func (n *Notifee) HandlePeerFound(pi peer.AddrInfo) {
	log.Println("Discovered new peer:", pi.ID.String(), pi.Addrs)
	// 可以直接连接
	n.h.Connect(context.Background(), pi)
}

// SetupMdns 启动本地 mDNS 服务用于局域网节点发现
func SetupMdns(ctx context.Context, h host.Host) error {
	n := &Notifee{h: h}
	service := mdns.NewMdnsService(h, rendezvous, n)
	_ = service // service 会在后台运行
	return nil
}