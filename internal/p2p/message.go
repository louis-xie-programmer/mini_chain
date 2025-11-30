package p2p

import "encoding/json"

// MsgType 定义节点间消息类型
type MsgType string

const (
	MsgTx       MsgType = "TX"       // 交易消息
	MsgBlock    MsgType = "BLOCK"    // 区块消息
	MsgGetChain MsgType = "GETCHAIN" // 请求区块链
	MsgChain    MsgType = "CHAIN"    // 返回区块链
)

// Message 节点间传输的数据结构
type Message struct {
	Type MsgType         `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Encode 将消息序列化为 JSON
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 将 JSON 数据解析为 Message
func Decode(raw []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(raw, &msg)
	return &msg, err
}
