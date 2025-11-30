package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"mini_chain/internal/blockchain"
	"mini_chain/internal/p2p"
)

// API 结构体，包含区块链、P2P节点和WebSocket管理器
type API struct {
	BC  *blockchain.Blockchain // 区块链实例
	P2P *p2p.Node              // P2P节点实例
	WS  *WSManager             // WebSocket管理器实例
}

// NewAPI 创建新的API实例
// bc: 区块链实例
// p2p: P2P节点实例
func NewAPI(bc *blockchain.Blockchain, p2p *p2p.Node) *API {
	ws := NewWSManager() // 创建WebSocket管理器
	go ws.Run()          // 启动WebSocket管理器
	return &API{
		BC:  bc,
		P2P: p2p,
		WS:  ws,
	}
}

// Run 启动API服务器
// addr: 服务器监听地址
func (api *API) Run(addr string) {
	r := mux.NewRouter()

	// REST端点
	r.HandleFunc("/chain", api.GetChain).Methods("GET")   // 获取区块链信息
	r.HandleFunc("/tx", api.PostTx).Methods("POST")       // 提交交易

	// WebSocket端点
	r.HandleFunc("/ws", api.WS.ServeWS)

	log.Println("REST + WS server running at", addr)
	http.ListenAndServe(addr, r)
}

// GET /chain 处理获取区块链信息的请求
func (api *API) GetChain(w http.ResponseWriter, r *http.Request) {
	// 从区块链获取最新区块
	latest := api.BC.GetLatest()
	// 目前只返回最新区块
	// 在实际实现中，会返回完整链
	json.NewEncoder(w).Encode(latest)
}

// POST /tx 处理提交交易的请求
func (api *API) PostTx(w http.ResponseWriter, r *http.Request) {
	var tx blockchain.UTXOTx
	// 解析请求体中的交易数据
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	// 验证交易结构
	if err := blockchain.ValidateTxStructure(tx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// 生成交易ID
	txid, err := blockchain.TxID(tx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 将交易添加到内存池
	blockchain.AddToMempool(txid)

	// 广播交易到P2P网络
	msg := &p2p.Message{
		Type: p2p.MsgTx,
		Data: mustMarshal(tx),
	}
	api.P2P.Broadcast(msg)

	// 推送给所有WebSocket客户端
	api.WS.broadcast <- mustMarshal(tx)
	w.WriteHeader(http.StatusCreated)
}

// mustMarshal 将接口对象序列化为JSON字节切片
// v: 待序列化的对象
// 返回序列化后的字节切片
func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}