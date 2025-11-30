package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"mini_chain/internal/blockchain"
	"mini_chain/internal/p2p"
)

type API struct {
	BC  *blockchain.Blockchain
	P2P *p2p.Node
	WS  *WSManager
}

func NewAPI(bc *blockchain.Blockchain, p2p *p2p.Node) *API {
	ws := NewWSManager()
	go ws.Run()
	return &API{
		BC:  bc,
		P2P: p2p,
		WS:  ws,
	}
}

func (api *API) Run(addr string) {
	r := mux.NewRouter()

	// REST endpoints
	r.HandleFunc("/chain", api.GetChain).Methods("GET")
	r.HandleFunc("/tx", api.PostTx).Methods("POST")

	// WebSocket
	r.HandleFunc("/ws", api.WS.ServeWS)

	log.Println("REST + WS server running at", addr)
	http.ListenAndServe(addr, r)
}

// GET /chain
func (api *API) GetChain(w http.ResponseWriter, r *http.Request) {
	// Get the latest block from the blockchain
	latest := api.BC.GetLatest()
	// For now, we'll just return the latest block
	// In a real implementation, we would return the full chain
	json.NewEncoder(w).Encode(latest)
}

// POST /tx
func (api *API) PostTx(w http.ResponseWriter, r *http.Request) {
	var tx blockchain.UTXOTx
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	
	// Validate transaction structure
	if err := blockchain.ValidateTxStructure(tx); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	// Generate transaction ID
	txid, err := blockchain.TxID(tx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Add transaction to mempool
	blockchain.AddToMempool(txid)

	// Broadcast transaction to P2P network
	msg := &p2p.Message{
		Type: p2p.MsgTx,
		Data: mustMarshal(tx),
	}
	api.P2P.Broadcast(msg)

	// 推送给所有 WS 客户端
	api.WS.broadcast <- mustMarshal(tx)
	w.WriteHeader(http.StatusCreated)
}

func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}