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
		if _, exists := node.receivedTransactions[upcomingTrx.Data]; !exists {
			// Check if the transaction already exists in the CurrentBlock
			duplicate := false
			for _, existingTrx := range node.CurrentBlock.BlockTransactions {
				if existingTrx.Data == upcomingTrx.Data {
					duplicate = true
					break
				}
			}

			if !duplicate {
				node.receivedTransactions[upcomingTrx.Data] = true
				floodArr = append(floodArr, upcomingTrx)
				node.CurrentBlock.BlockTransactions = append(node.CurrentBlock.BlockTransactions, upcomingTrx)
			}
		}
	}

	if len(floodArr) > 0 {
		node.floodingTrx(floodArr)
	}

	if len(node.CurrentBlock.BlockTransactions) >= 4 {
		node.mineCheck()
	}
}

func (node *Node) mineCheck() {
	if len(node.CurrentBlock.BlockTransactions) >= 4 {
		fmt.Printf("Node %d: Starting to mine block with transactions: %+v\n", node.ID, node.CurrentBlock.BlockTransactions)
		node.CurrentBlock.mineBlock()
		fmt.Printf("Node %d mined a block: %+v\n", node.ID, node.CurrentBlock)

		node.floodingBlock(node.CurrentBlock)
		// Reset the current block
		node.CurrentBlock = Block{
			PrevBlockHash:     node.CurrentBlock.CurrentBlockHash,
			Timestamp:         time.Now().Unix(),
			BlockTransactions: []Transaction{},
			Nonce:             0,
		}
	}
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
	node.mu.Lock()
	defer node.mu.Unlock()

	if _, exists := node.receivedBlocks[block.CurrentBlockHash]; exists {
		fmt.Printf("Node %d: Block already received, skipping.\n", node.ID)
		return
	}

	// Validate block
	if block.CurrentBlockHash == block.blockHashCalculation() {
		fmt.Printf("Node %d: Valid block received. Adding to blockchain.\n", node.ID)
		node.receivedBlocks[block.CurrentBlockHash] = true
		blockchain.addBlock(&block)
		node.floodingBlock(block)
	} else {
		fmt.Printf("Node %d: Invalid block received. Hash mismatch.\n", node.ID)
	}
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

	node.CurrentBlock.BlockTransactions = append(node.CurrentBlock.BlockTransactions, transactions...)

	for _, neighbor := range node.Neighbors {
		go node.brodcastingTrxToNeigborNodes(neighbor, node.CurrentBlock.BlockTransactions)
	}
	node.CurrentBlock.MerkleRoot = merkleRoot(node.CurrentBlock.BlockTransactions).hash
}

func (node *Node) floodingBlock(block Block) {
	node.mu.Lock()
	if len(node.Neighbors) == 0 {
		fmt.Printf("Node %d: No neighbors to broadcast to.\n", node.ID)
		node.mu.Unlock()
		return
	}
	neighborsCopy := append([]string{}, node.Neighbors...)
	node.mu.Unlock()

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

	fmt.Print("Enter your choice (1 or 2): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		part01()
	case 2:
		part02()
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}

func part01() {

	transactions1 := []Transaction{
		{Data: "ahmed yassin"},
		{Data: "ikhwan al-muslimoon"},
		{Data: "hayaat tahrir al sham"},
		{Data: "hamas"},
	}

	transactions2 := []Transaction{
		{Data: "syria"},
		{Data: "jolani"},
		{Data: "hts"},
		{Data: "hamas"},
	}

	changedTransactions := []Transaction{
		{Data: "khaled mashal"},
		{Data: "muhammad deif"},
		{Data: "yahya sinwar"},
		{Data: "ismail haniya"},
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

// 	// Wait for all goroutines to finish
// 	wg.Wait()

// 	fmt.Println("All work is done. Exiting program.")
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

	transactions := []Transaction{
		{Data: "1500USD Sent"},
		{Data: "1600USD Sent"},
		{Data: "1700USD Sent"},
		{Data: "1800USD Sent"},
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

	select {}
}
