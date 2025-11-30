package p2p

import "encoding/json"

// MsgType 定义节点间消息类型
type MsgType string

// 定义各种消息类型的常量
const (
	MsgTx       MsgType = "TX"       // 交易消息
	MsgBlock    MsgType = "BLOCK"    // 区块消息
	MsgGetChain MsgType = "GETCHAIN" // 请求区块链
	MsgChain    MsgType = "CHAIN"    // 返回区块链
)

// Message 节点间传输的数据结构
type Message struct {
	Type MsgType         `json:"type"` // 消息类型
	Data json.RawMessage `json:"data"` // 消息数据
}

// Encode 将消息序列化为JSON
// 返回序列化后的字节切片和可能的错误
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// Decode 将JSON数据解析为Message
// raw: 待解析的JSON数据
// 返回解析后的Message指针和可能的错误
func Decode(raw []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(raw, &msg)
	return &msg, err
}