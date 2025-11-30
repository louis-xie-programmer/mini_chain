# Simple P2P Blockchain Implementation

This is a simple implementation of a blockchain node using TCP-based peer-to-peer communication.

## Features

- Basic blockchain implementation with proof-of-work consensus
- Transaction signing and verification using ECDSA
- Peer-to-peer communication over TCP
- Command-line interface for interaction
- Automatic mining and block broadcasting

## How to Run

### Single Node

To run a single node:

```bash
go run main.go 3000
```

This will start a node listening on port 3000.

### Multiple Nodes Network

To run multiple nodes that can communicate with each other:

Terminal 1:
```bash
go run main.go 3000 localhost:3001,localhost:3002
```

Terminal 2:
```bash
go run main.go 3001 localhost:3000,localhost:3002
```

Terminal 3:
```bash
go run main.go 3002 localhost:3000,localhost:3001
```

Each node will:
1. Listen on its specified port
2. Connect to the peer nodes specified in the command line
3. Synchronize blockchain data with peers
4. Participate in mining and transaction processing

## Available Commands

Once a node is running, you can use the following commands:

- `tx <to> <amount>` - Create and broadcast a new transaction
- `chain` - Display the current blockchain
- `pool` - Display pending transactions in the pool
- `peers` - Show connected peer nodes
- `addpeer <host:port>` - Add a new peer node
- `exit` - Stop the node

## Architecture Overview

### Core Components

1. **Blockchain** - Stores the chain of blocks
2. **Transaction Pool** - Holds pending transactions
3. **P2P Network Layer** - Handles TCP connections and message passing
4. **Consensus Engine** - Implements Proof-of-Work mining
5. **Wallet System** - ECDSA key generation and transaction signing

### Data Structures

- `Transaction` - Represents a value transfer between addresses
- `Block` - Contains a collection of transactions with metadata
- `Message` - Network protocol message format

### Network Protocol

Messages are exchanged in JSON format with the following types:
- `TX` - Transaction propagation
- `BLOCK` - New block announcement
- `GETCHAIN` - Request for blockchain data
- `CHAIN` - Response with blockchain data

## Security Notes

This is a simplified educational implementation and lacks many security features required for production use:
- No authentication between nodes
- No encryption of network traffic
- Simplified consensus mechanism
- Basic transaction validation only

## Testing

To test the implementation:

1. Start multiple nodes as described above
2. Create transactions using the `tx` command
3. Observe how transactions propagate through the network
4. Watch blocks being mined and added to the chain
5. Check that all nodes synchronize their blockchains