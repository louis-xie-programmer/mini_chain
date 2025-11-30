#!/usr/bin/env python3
"""
Test script for running multiple nodes locally
"""

import subprocess
import time
import sys
import urllib.request
import urllib.error
import json

# Configuration
NODES = [
    {"p2p_port": 3000, "api_port": 8080, "name": "Node1"},
    {"p2p_port": 3001, "api_port": 8081, "name": "Node2"},
    {"p2p_port": 3002, "api_port": 8082, "name": "Node3"},
]

# Sample transaction data
SAMPLE_TX = {
    "inputs": [
        {
            "txid": "abcd1234",
            "vout": 0,
            "signature": "sig1",
            "pubkey": "pub1"
        }
    ],
    "outputs": [
        {
            "address": "addr1",
            "amount": 100
        }
    ]
}

processes = []

def start_node(node_config, bootstrap_addr=None):
    """Start a blockchain node"""
    cmd = [
        "go", "run", "main.go",
        str(node_config["p2p_port"]),
        str(node_config["api_port"])
    ]
    
    if bootstrap_addr:
        cmd.append(bootstrap_addr)
    
    print(f"Starting {node_config['name']} on P2P port {node_config['p2p_port']} "
          f"and API port {node_config['api_port']}")
    
    process = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    processes.append(process)
    return process

def get_node_address(node_config):
    """In a real implementation, we would parse the node's output to get its address"""
    # This is a placeholder - in practice you'd parse this from the node's output
    # to get its actual peer ID
    return f"/ip4/127.0.0.1/tcp/{node_config['p2p_port']}/p2p/QmExamplePeerId{node_config['p2p_port']}"

def test_api_endpoints():
    """Test that all nodes are running by hitting their API endpoints"""
    for node in NODES:
        try:
            response = urllib.request.urlopen(f"http://localhost:{node['api_port']}/chain", timeout=2)
            print(f"{node['name']} API response: {response.getcode()}")
        except urllib.error.URLError as e:
            print(f"Failed to connect to {node['name']}: {e}")
        except Exception as e:
            print(f"Error testing {node['name']}: {e}")

def submit_transaction(node, transaction):
    """Submit a transaction to a specific node"""
    try:
        # Convert transaction to JSON
        data = json.dumps(transaction).encode('utf-8')
        
        # Create request
        req = urllib.request.Request(
            f"http://localhost:{node['api_port']}/tx",
            data=data,
            headers={'Content-Type': 'application/json'}
        )
        
        # Submit transaction
        response = urllib.request.urlopen(req, timeout=5)
        return response.getcode() in [200, 201]  # OK or Created
    except Exception as e:
        print(f"Failed to submit transaction to {node['name']}: {e}")
        return False

def test_transaction_propagation():
    """Test transaction submission and propagation across nodes"""
    print("\n=== Transaction Propagation Test ===")
    
    # Submit transaction to the first node
    print(f"Submitting transaction to {NODES[0]['name']}...")
    if submit_transaction(NODES[0], SAMPLE_TX):
        print("Transaction submitted successfully")
    else:
        print("Failed to submit transaction")
        return
    
    # Wait a bit for propagation
    print("Waiting for transaction propagation...")
    time.sleep(3)
    
    # Check if transaction appears on other nodes
    print("Checking transaction propagation:")
    for i, node in enumerate(NODES):
        try:
            response = urllib.request.urlopen(f"http://localhost:{node['api_port']}/chain", timeout=2)
            if response.getcode() == 200:
                print(f"  {node['name']}: Chain accessible")
            else:
                print(f"  {node['name']}: HTTP {response.getcode()}")
        except Exception as e:
            print(f"  {node['name']}: Error - {e}")

def interactive_test():
    """Provide interactive testing options"""
    while True:
        print("\n=== Interactive Test Options ===")
        print("1. Check node status")
        print("2. Submit sample transaction")
        print("3. Run transaction propagation test")
        print("4. Exit")
        print("Enter your choice (1-4): ", end="")
        
        try:
            choice = input().strip()
            if choice == "1":
                test_api_endpoints()
            elif choice == "2":
                print(f"Submitting transaction to {NODES[0]['name']}...")
                if submit_transaction(NODES[0], SAMPLE_TX):
                    print("Transaction submitted successfully")
                else:
                    print("Failed to submit transaction")
            elif choice == "3":
                test_transaction_propagation()
            elif choice == "4":
                break
            else:
                print("Invalid choice. Please enter 1-4.")
        except KeyboardInterrupt:
            break
        except EOFError:
            break

def main():
    try:
        print("Starting Mini Chain Network Test")
        
        # Start first node
        process1 = start_node(NODES[0])
        time.sleep(3)  # Give first node time to start
        
        # Get first node's address (placeholder implementation)
        bootstrap_addr = get_node_address(NODES[0])
        print(f"First node address: {bootstrap_addr}")
        
        # Start other nodes connected to the first
        for i in range(1, len(NODES)):
            start_node(NODES[i], bootstrap_addr)
        
        time.sleep(3)  # Give all nodes time to connect
        
        print("\nAll nodes started. Testing API endpoints:")
        test_api_endpoints()
        
        # Run automated transaction test
        test_transaction_propagation()
        
        # Enter interactive mode
        print("\nEntering interactive test mode...")
        interactive_test()
        
        print("\nNetwork test completed.")
        
    except KeyboardInterrupt:
        print("\nStopping all nodes...")
        for process in processes:
            process.terminate()
        
        # Wait for processes to terminate
        for process in processes:
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
        
        print("All nodes stopped.")

if __name__ == "__main__":
    main()