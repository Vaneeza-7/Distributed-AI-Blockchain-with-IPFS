package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

// Node represents a peer in the P2P network
type Node struct {
	ID                   int
	Port                 int
	IP                   string
	Neighbors            []string
	CurrentBlock         Block
	ReceivedBlock        Block
	mu                   sync.Mutex
	receivedTransactions map[string]bool
	receivedBlocks       map[string]bool
}

// BootstrapNode represents the bootstrap node
var BootstrapNode = Node{
	ID:   0,
	Port: 5000,
	IP:   "127.0.0.1",
}

var nodeIDCounter = 1
var portCounter = 6000
var minedTransactions = make(map[string]bool)
var minedMu sync.Mutex

// RegisterNode registers a new node with the bootstrap node
func RegisterNode(node *Node) {
	BootstrapNode.mu.Lock()
	defer BootstrapNode.mu.Unlock()

	// Assign a unique ID and port to the new node
	node.ID = nodeIDCounter
	nodeIDCounter++

	node.Port = portCounter
	portCounter++

	// Update the BootstrapNode's Neighbors list
	BootstrapNode.Neighbors = append(BootstrapNode.Neighbors, fmt.Sprintf("%s:%d", node.IP, node.Port))
}

func assigningNeighbor(node *Node) {
	BootstrapNode.mu.Lock()
	existingNodes := append([]string{}, BootstrapNode.Neighbors...)
	BootstrapNode.mu.Unlock()

	if len(existingNodes) == 0 {
		fmt.Printf("Node %d: No neighbors available to assign.\n", node.ID)
		return
	}

	rand.Shuffle(len(existingNodes), func(i, j int) {
		existingNodes[i], existingNodes[j] = existingNodes[j], existingNodes[i]
	})

	maxNeighbors := 5
	neighborSet := make(map[string]bool)
	for _, potentialNeighbor := range existingNodes {
		if potentialNeighbor != fmt.Sprintf("%s:%d", node.IP, node.Port) && !neighborSet[potentialNeighbor] {
			node.Neighbors = append(node.Neighbors, potentialNeighbor)
			neighborSet[potentialNeighbor] = true
			maxNeighbors--
		}
		if maxNeighbors == 0 {
			break
		}
	}

	fmt.Printf("Node %d assigned neighbors: %v\n", node.ID, node.Neighbors)
}

// StartNode starts a new node as both a server and a client
func StartNode(node *Node) {
	node.CurrentBlock = Block{
		PrevBlockHash: blockchain.head.CurrentBlockHash,
		Timestamp:     time.Now().Unix(),
		Nonce:         0,
	}
	go node.startServer()
	go node.startClient()
	go node.mineCheck()
}

// startServer starts the server for the node
func (node *Node) startServer() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", node.Port))
	if err != nil {
		fmt.Printf("Node %d: There was error starting server: %v\n", node.ID, err)
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("Node %d: There was error accepting connection: %v\n", node.ID, err)
			continue
		}

		go node.handleClient(conn)
	}
}

func (node *Node) handleTranscation(trx []Transaction) {
	floodArr := []Transaction{}
	node.mu.Lock()
	defer node.mu.Unlock()

	for _, upcomingTrx := range trx {
		// Add all transactions to the flood array for broadcasting
		floodArr = append(floodArr, upcomingTrx)

		// Check if the transaction has already been mined
		minedMu.Lock()
		_, isMined := minedTransactions[upcomingTrx.DatasetHash]
		minedMu.Unlock()

		// Add transaction to the current block only if it is unmined
		if !isMined && !node.receivedTransactions[upcomingTrx.DatasetHash] {
			// Check if the transaction is a duplicate in the CurrentBlock
			duplicate := false
			for _, existingTrx := range node.CurrentBlock.BlockTransactions {
				if existingTrx.DatasetHash == upcomingTrx.DatasetHash {
					duplicate = true
					break
				}
			}

			// Add the transaction if it's not a duplicate and the block isn't full
			if !duplicate && len(node.CurrentBlock.BlockTransactions) < 5 {
				node.receivedTransactions[upcomingTrx.DatasetHash] = true
				node.CurrentBlock.BlockTransactions = append(node.CurrentBlock.BlockTransactions, upcomingTrx)
				fmt.Printf("Node %d: Added transaction to current block: %+v\n", node.ID, upcomingTrx)

				//as soon as length of block transaction is 5, we will start mining the block
				if len(node.CurrentBlock.BlockTransactions) == 5 {
					// Update MerkleRoot for the block based on the first 5 transactions
					node.CurrentBlock.MerkleRoot = merkleRoot(node.CurrentBlock.BlockTransactions).hash
					go node.mineCheck()
				}
			}
		}
	}

	// Flood all received transactions (both mined and unmined)
	if len(floodArr) > 0 {
		node.floodingTrx(floodArr)
	}

	// Mine the block if it has exactly 5 transactions
	// if len(node.CurrentBlock.BlockTransactions) == 5 {
	// 	node.mineCheck()
	// }
}

// func (node *Node) mineCheck() {

// 	//node.mu.Lock()
// 	//defer node.mu.Unlock()

// 	// Ensure block is ready for mining
// 	if len(node.CurrentBlock.BlockTransactions) == 5 {
// 		fmt.Printf("Node %d: Starting to mine block with transactions: %+v\n", node.ID, node.CurrentBlock.BlockTransactions)

// 		// Mine the block
// 		node.CurrentBlock.mineBlock()

// 		// Broadcast the mined block
// 		fmt.Printf("Node %d mined a block: %+v\n", node.ID, node.CurrentBlock)
// 		node.mu.Lock()
// 		node.floodingBlock(node.CurrentBlock)
// 		node.mu.Unlock()
// 		fmt.Printf("Node %d: Vaneeza Mined block broadcasted.\n", node.ID)

// 		// Mark transactions as mined globally
// 		minedMu.Lock()
// 		for _, trx := range node.CurrentBlock.BlockTransactions {
// 			minedTransactions[trx.Data] = true
// 		}
// 		minedMu.Unlock()

// 		// Reset the current block
// 		node.CurrentBlock = Block{
// 			PrevBlockHash:     node.CurrentBlock.CurrentBlockHash,
// 			Timestamp:         time.Now().Unix(),
// 			BlockTransactions: []Transaction{},
// 			Nonce:             0,
// 		}
// 	}
// }

func (node *Node) mineCheck() {
	if len(node.CurrentBlock.BlockTransactions) < 5 {
		fmt.Printf("Node %d: Not enough transactions to mine a block.\n", node.ID)
		return
	}
	blockToMine := node.CurrentBlock

	// Mine the block
	fmt.Printf("Node %d: Starting to mine block with transactions: %+v\n", node.ID, blockToMine.BlockTransactions)
	blockToMine.mineBlock()
	fmt.Printf("Node %d: Block mined: %+v\n", node.ID, blockToMine)

	// Broadcast the block
	fmt.Printf("Node %d: Broadcasting mined block.\n", node.ID)
	node.floodingBlock(blockToMine)

	// Mark transactions as mined globally
	minedMu.Lock()
	for _, trx := range blockToMine.BlockTransactions {
		minedTransactions[trx.DatasetHash] = true
	}
	minedMu.Unlock()

	node.mu.Lock()
	node.CurrentBlock = Block{
		PrevBlockHash:     blockToMine.CurrentBlockHash,
		Timestamp:         time.Now().Unix(),
		BlockTransactions: []Transaction{},
		Nonce:             0,
	}
	node.mu.Unlock()
}

func (bc *BlockChain) containsBlock(hash string) bool {
	current := bc.head
	for current != nil {
		if current.CurrentBlockHash == hash {
			return true
		}
		current = current.Next
	}
	return false
}

func (node *Node) handleBlock(block Block) {
	fmt.Printf("Node %d: +++++++++++++++++++++++Handling block++++++++++++++++ %+v\n", node.ID, block)

	//node.mu.Lock()
	// defer node.mu.Unlock()

	if _, exists := node.receivedBlocks[block.CurrentBlockHash]; exists {
		fmt.Printf("Node %d: Block already received, skipping.\n", node.ID)
		//node.mu.Unlock()
		return
	}

	if condition := blockchain.containsBlock(block.CurrentBlockHash); condition {
		fmt.Printf("Node %d: Block already in blockchain, skipping.\n", node.ID)
		return
	}

	//node.receivedBlocks[block.CurrentBlockHash] = true
	//node.mu.Unlock()

	if block.CurrentBlockHash == block.blockHashCalculation() {
		node.receivedBlocks[block.CurrentBlockHash] = true
		blockchain.addBlock(&block)
		fmt.Printf("Node %d: Block added to blockchain.\n", node.ID)
		node.floodingBlock(block)
		minedMu.Lock()
		for _, trx := range block.BlockTransactions {
			minedTransactions[trx.DatasetHash] = true
		}
		minedMu.Unlock()

		//node.mu.Lock()
		unminedTransactions := []Transaction{}
		for _, trx := range node.CurrentBlock.BlockTransactions {
			if !minedTransactions[trx.DatasetHash] {
				unminedTransactions = append(unminedTransactions, trx)
			}
		}
		node.CurrentBlock.BlockTransactions = unminedTransactions
		//node.mu.Unlock()
	} else {
		fmt.Printf("Node %d: Invalid block received. Hash mismatch.\n", node.ID)
	}

	fmt.Printf("Node %d: Block handling complete.\n", node.ID)
}

// handling the transaction receiveing from the client
func (node *Node) handleClient(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	// Read the raw JSON message
	var rawMessage json.RawMessage
	if err := decoder.Decode(&rawMessage); err != nil {
		fmt.Printf("Node %d: Error reading data from neighbor: %v\n", node.ID, err)
		return
	}

	// Try decoding as a Block
	var block Block
	if err := json.Unmarshal(rawMessage, &block); err == nil {
		fmt.Printf("Node %d: Received block: %+v\n", node.ID, block)
		node.handleBlock(block)
		return
	}

	// Try decoding as a list of Transactions
	var transactions []Transaction
	if err := json.Unmarshal(rawMessage, &transactions); err == nil {
		fmt.Printf("Node %d: Received transactions: %+v\n", node.ID, transactions)
		node.handleTranscation(transactions)
		return
	}

	// Log the failure to interpret the data
	fmt.Printf("Node %d: Failed to decode received data\n", node.ID)
}

// startClient simulates the client functionality by periodically contacting neighbors
func (node *Node) startClient() {
	for {
		node.mu.Lock()
		neighbors := append([]string{}, node.Neighbors...)
		node.mu.Unlock()

		for _, neighbor := range neighbors {
			go node.contactNeighbor(neighbor)
		}
	}
}

// contactNeighbor simulates the client contacting a neighbor
func (node *Node) contactNeighbor(neighbor string) {
}

// Function to broadcast transactions
func (node *Node) floodingTrx(transactions []Transaction) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if len(node.Neighbors) == 0 {
		fmt.Printf("Node %d: No neighbors to broadcast to.\n", node.ID)
		return
	}

	// Add only the first 5 transactions to the current block
	for _, trx := range transactions {
		if len(node.CurrentBlock.BlockTransactions) < 5 {
			node.CurrentBlock.BlockTransactions = append(node.CurrentBlock.BlockTransactions, trx)
		}
	}

	// Update MerkleRoot for the block based on the first 5 transactions
	if len(node.CurrentBlock.BlockTransactions) > 0 {
		node.CurrentBlock.MerkleRoot = merkleRoot(node.CurrentBlock.BlockTransactions).hash
	}

	// Broadcast all received transactions, not just the first 5
	for _, neighbor := range node.Neighbors {
		go node.brodcastingTrxToNeigborNodes(neighbor, transactions)
	}
}

func (node *Node) floodingBlock(block Block) {
	fmt.Printf("Hello from node %d\n", node.ID)
	//node.mu.Lock()
	if len(node.Neighbors) == 0 {
		fmt.Printf("Node %d: No neighbors to broadcast to.\n", node.ID)
		node.mu.Unlock()
		return
	}
	neighborsCopy := append([]string{}, node.Neighbors...)
	//node.mu.Unlock()

	fmt.Printf("Node %d: Broadcasting block to neighbors: %+v\n", node.ID, neighborsCopy)
	for _, neighbor := range neighborsCopy {
		go node.brodcastingBlockToNeighborNodes(neighbor, block)
	}
}

func (node *Node) brodcastingBlockToNeighborNodes(neighbor string, block Block) {
	fmt.Printf("Node %d: Broadcasting block to neighbor %s\n", node.ID, neighbor)
	conn, err := net.Dial("tcp", neighbor)
	if err != nil {
		fmt.Printf("Node %d: Error connecting to neighbor %s: %v\n", node.ID, neighbor, err)
		return
	}
	defer conn.Close()

	// Serialize the block as JSON
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(block); err != nil {
		fmt.Printf("Node %d: Error encoding block to neighbor %s: %v\n", node.ID, neighbor, err)
		return
	}
}

func (node *Node) brodcastingTrxToNeigborNodes(neighbor string, transactions []Transaction) {

	fmt.Printf("Node %d: Broadcasting transactions to neighbor %s\n", node.ID, neighbor)
	node.mu.Lock()
	defer node.mu.Unlock()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", neighbor)
	if err != nil {
		fmt.Printf("Node %d: There was error connecting to neighbor: %v\n", node.ID, err)
		return
	}
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(transactions); err != nil {
		fmt.Printf("Node %d: There was an error encoding and sending transactions to neighbor: %v\n", node.ID, err)
		return
	}
}

// DisplayP2PNetwork prints the details of the P2P network
func DisplayP2PNetwork(nodes []Node) {
	fmt.Println("P2P Network:")
	for i, node := range nodes {
		fmt.Printf("Node %d: ID=%d, IP=%s, Port=%d, Neighbors=%v, Block=%v\n", i+1, node.ID, node.IP, node.Port, node.Neighbors, node.CurrentBlock)
	}
	fmt.Println("Bootstrap Node:", BootstrapNode)
}

var blockchain = BlockChain{}

func main() {
	var choice int

	fmt.Println("Select an option:")
	fmt.Println("1. Part #01")
	fmt.Println("2. Part #02")
	fmt.Println("3. Retrieve from IPFS")

	fmt.Print("Enter your choice (1 or 2 or 3): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		part01()
	case 2:
		part02()
	case 3:
		//retrieve_all_files()
		//all_files_hash()
		run_algo()
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}

func part01() {

	clusterCenters := [][]float64{
		{0.07996, -0.90932, -0.38071},
		{1.34745, 0.18654, 0.90497},
		{-1.15133, 0.83523, -0.30381},
	}

	// Example data for cluster sizes
	clusterSizes := map[string]int{
		"0": 67,
		"1": 49,
		"2": 62,
	}

	// Create an AI output object
	aiOutput := AIOutput{
		ClusterCenters: clusterCenters,
		Inertia:        1285.6677396078076,
		ClusterSizes:   clusterSizes,
	}

	transactions1 := []Transaction{
		{
			DatasetHash: "sha256_hash_of_dataset1",
			AlgoHash:    "sha256_hash_of_algorithm1",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset2",
			AlgoHash:    "sha256_hash_of_algorithm2",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset3",
			AlgoHash:    "sha256_hash_of_algorithm3",
			Output:      aiOutput,
		},
	}

	transactions2 := []Transaction{
		{
			DatasetHash: "sha256_hash_of_dataset4",
			AlgoHash:    "sha256_hash_of_algorithm4",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset5",
			AlgoHash:    "sha256_hash_of_algorithm5",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset6",
			AlgoHash:    "sha256_hash_of_algorithm6",
			Output:      aiOutput,
		},
	}

	changedTransactions := []Transaction{
		{
			DatasetHash: "sha256_hash_of_dataset1",
			AlgoHash:    "sha256_hash_of_algorithm1",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset2",
			AlgoHash:    "sha256_hash_of_algorithm2",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset3",
			AlgoHash:    "sha256_hash_of_algorithm3",
			Output:      aiOutput,
		},
	}

	prevBlockHash := ""
	block1 := blockCreation(prevBlockHash, transactions1)
	block2 := blockCreation(block1.CurrentBlockHash, transactions2)
	fmt.Println(block1)
	fmt.Println(block2)

	blockchain := BlockChain{}
	blockchain.addBlock(block1)
	blockchain.addBlock(block2)

	fmt.Println("*****BLOCK CHAIN*****")
	blockchain.displayBlockChain()

	if blockchain.validityCheck() {
		fmt.Println("VALID BLOCKS :)")
	} else {
		fmt.Println("The transaction in block has been tampered")
	}

	changeBlock(block1, changedTransactions)
	fmt.Println(block1)

	if blockchain.validityCheck() {
		fmt.Println("VALID BLOCKS :)")
	} else {
		fmt.Println("The transaction in block has been tampered")
	}

	fmt.Println("********MERKLE TREE********")
	displayMerkleTree(merkleRoot(block1.BlockTransactions), "")
}

// func part02() {
// 	var wg sync.WaitGroup

// 	// Define the number of transactions to generate
// 	totalTransactions := 100

// 	// Genesis block creation
// 	genesisBlock := Block{
// 		PrevBlockHash:     "",
// 		Timestamp:         time.Now().Unix(),
// 		Nonce:             0,
// 		MerkleRoot:        "",
// 		BlockTransactions: []Transaction{},
// 	}

// 	genesisBlock.CurrentBlockHash = genesisBlock.blockHashCalculation()
// 	blockchain.addBlock(&genesisBlock)

// 	// Create nodes
// 	nodes := make([]Node, 8)

// 	for i := range nodes {
// 		nodes[i].IP = "127.0.0.1"
// 		nodes[i].receivedTransactions = make(map[string]bool)
// 		nodes[i].receivedBlocks = make(map[string]bool)
// 		RegisterNode(&nodes[i])
// 	}

// 	for i := range nodes {
// 		assigningNeighbor(&nodes[i])
// 		StartNode(&nodes[i])
// 	}

// 	// Start the transaction generator with WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		generateTransactions(nodes, totalTransactions)
// 	}()

// 	// Allow the network to operate and wait for goroutines
// 	time.Sleep(12 * time.Second)

// 	// Display the P2P network state
// 	DisplayP2PNetwork(nodes)

// 	fmt.Println("*****BLOCK CHAIN*****")
// 	blockchain.displayBlockChain()
// 	//print the block in blockchain
// 	for current := blockchain.head; current != nil; current = current.Next {
// 		fmt.Println(current)
// 	}

// 	// Wait for all goroutines to finish
// 	wg.Wait()

// 	fmt.Println("All work is done. Exiting program.")

// 	select {}
// }

// // Function to generate a specific number of transactions
// func generateTransactions(nodes []Node, totalTransactions int) {
// 	batchSize := 5 // Number of transactions in each batch
// 	transactionCount := 0

// 	for transactionCount < totalTransactions {
// 		var batch []Transaction
// 		for i := 0; i < batchSize && transactionCount < totalTransactions; i++ {
// 			transaction := Transaction{Data: fmt.Sprintf("%dUSD Sent", rand.Intn(10000))}
// 			batch = append(batch, transaction)
// 			transactionCount++
// 		}
// 		fmt.Printf("Generated transactions batch: %v\n", batch)

// 		// Select a random node to initiate the broadcast
// 		randomNode := nodes[rand.Intn(len(nodes))]
// 		randomNode.floodingTrx(batch)

// 		time.Sleep(500 * time.Millisecond) // Delay between batches
// 	}

// 	fmt.Println("All transactions generated.")
// }

func part02() {

	genesisBlock := Block{
		PrevBlockHash:     "",
		Timestamp:         time.Now().Unix(),
		Nonce:             0,
		MerkleRoot:        "",
		BlockTransactions: []Transaction{},
	}

	genesisBlock.CurrentBlockHash = genesisBlock.blockHashCalculation()
	blockchain.addBlock(&genesisBlock)

	clusterCenters := [][]float64{
		{0.07996, -0.90932, -0.38071},
		{1.34745, 0.18654, 0.90497},
		{-1.15133, 0.83523, -0.30381},
	}

	// Example data for cluster sizes
	clusterSizes := map[string]int{
		"0": 67,
		"1": 49,
		"2": 62,
	}

	// Create an AI output object
	aiOutput := AIOutput{
		ClusterCenters: clusterCenters,
		Inertia:        1285.6677396078076,
		ClusterSizes:   clusterSizes,
	}

	transactions := []Transaction{
		{
			DatasetHash: "sha256_hash_of_dataset1",
			AlgoHash:    "sha256_hash_of_algorithm1",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset2",
			AlgoHash:    "sha256_hash_of_algorithm2",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset3",
			AlgoHash:    "sha256_hash_of_algorithm3",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset4",
			AlgoHash:    "sha256_hash_of_algorithm4",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset5",
			AlgoHash:    "sha256_hash_of_algorithm5",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset6",
			AlgoHash:    "sha256_hash_of_algorithm6",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset7",
			AlgoHash:    "sha256_hash_of_algorithm7",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset8",
			AlgoHash:    "sha256_hash_of_algorithm8",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset9",
			AlgoHash:    "sha256_hash_of_algorithm9",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset10",
			AlgoHash:    "sha256_hash_of_algorithm10",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset11",
			AlgoHash:    "sha256_hash_of_algorithm11",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset12",
			AlgoHash:    "sha256_hash_of_algorithm12",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset13",
			AlgoHash:    "sha256_hash_of_algorithm13",
			Output:      aiOutput,
		},
		{
			DatasetHash: "sha256_hash_of_dataset14",
			AlgoHash:    "sha256_hash_of_algorithm14",
			Output:      aiOutput,
		},
	}
	nodes := make([]Node, 8)

	for i := range nodes {
		nodes[i].IP = "127.0.0.1"
		nodes[i].receivedTransactions = make(map[string]bool)
		nodes[i].receivedBlocks = make(map[string]bool)
		RegisterNode(&nodes[i])
	}

	for i := range nodes {
		assigningNeighbor(&nodes[i])
		StartNode(&nodes[i])
	}
	nodes[4].floodingTrx(transactions)

	time.Sleep(12 * time.Second)
	DisplayP2PNetwork(nodes)

	fmt.Println("*****BLOCK CHAIN*****")
	blockchain.displayBlockChain()
	//print the block in blockchain
	for current := blockchain.head; current != nil; current = current.Next {
		fmt.Println(current)
	}

	select {}
}
