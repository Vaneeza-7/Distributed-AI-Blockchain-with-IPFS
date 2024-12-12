package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const _MINING_DIFFICULTY_ = 5

type Transaction struct {
	Data string
}

type Block struct {
	currentBlockHash  string
	prevBlockHash     string
	timestamp         int64
	nonce             int
	merkleroot        string
	BlockTransactions []Transaction
	next              *Block
}

type BlockChain struct {
	head *Block
}

type MerkleNode struct {
	hash  string
	left  *MerkleNode
	right *MerkleNode
}

// Displaying the BlockChain
func (obj *BlockChain) displayBlockChain() {
	currentNode := obj.head
	for currentNode != nil {
		fmt.Printf("Block Hash: %s\n", currentNode.currentBlockHash)
		currentNode = currentNode.next
	}
}

// Adding the block in the BlockChain
func (obj *BlockChain) addBlock(b *Block) {
	if obj.head == nil {
		obj.head = b
	} else {
		currentNode := obj.head
		for currentNode.next != nil {
			currentNode = currentNode.next
		}
		currentNode.next = b
	}
}

// Calculating the hash of current Block
func (block1 *Block) blockHashCalculation() string {
	//Block header consist of prevBlockhash,nonce,timestamp,merkleroot and trasactions in that block
	blockHeader := fmt.Sprintf("%s%d%d%s%s", block1.prevBlockHash, block1.timestamp, block1.nonce, block1.merkleroot, block1.BlockTransactions)
	hash_value := sha256.Sum256([]byte(blockHeader))
	hash_string := hex.EncodeToString(hash_value[:])
	return hash_string
}

// Creation of new Block
func blockCreation(prevBlockHash string, trasactions []Transaction) *Block {
	block := &Block{
		prevBlockHash:     prevBlockHash,
		timestamp:         time.Now().Unix(),
		nonce:             0,
		BlockTransactions: trasactions,
		next:              nil,
	}

	block.merkleroot = merkleRoot(trasactions).hash
	mined := block.mineBlock()
	if mined {
		return block
	} else {
		fmt.Println("Block is not minned")
	}
	return nil
}

// Mining the block
func (b *Block) mineBlock() bool {
	targetPrefix := strings.Repeat("0", _MINING_DIFFICULTY_)
	for {
		b.currentBlockHash = b.blockHashCalculation()
		currentPrefix := b.currentBlockHash[:_MINING_DIFFICULTY_]
		//print this after every 500 iterations
		if b.nonce%500 == 0 {
			fmt.Printf("Current Hash: %s, Current Prefix: %s, Target Prefix: %s\n", b.currentBlockHash, currentPrefix, targetPrefix)
		}
		if currentPrefix == targetPrefix {
			fmt.Println("Block Mined:", b.currentBlockHash)
			return true
		}
		b.nonce++
	}
}

// Checking the validity of block along with chain
func (obj *BlockChain) validityCheck() bool {
	currentNode := obj.head
	for currentNode != nil && currentNode.next != nil {
		currentBlock := currentNode
		nextBlock := currentNode.next

		if currentBlock.currentBlockHash != nextBlock.prevBlockHash {
			return false
		}

		if currentBlock.currentBlockHash != currentBlock.blockHashCalculation() && nextBlock.currentBlockHash != nextBlock.blockHashCalculation() {
			return false
		}

		currentNode = currentNode.next
	}
	return true
}

// Changing the block
func changeBlock(b *Block, transactions []Transaction) {
	b.BlockTransactions = transactions
	b.merkleroot = merkleRoot(transactions).hash
	b.currentBlockHash = b.blockHashCalculation()
}

// Displaying Block
func (b *Block) String() string {
	return fmt.Sprintf("Block:\n"+
		"|  Previous Block Hash: %s\n"+
		"|  Current Block Hash:  %s\n"+
		"|  Timestamp:           %s\n"+
		"|  Nonce:               %d\n"+
		"|  Merkle Root:         %s\n"+
		"|  Transactions:        %v\n",
		b.prevBlockHash,
		b.currentBlockHash,
		time.Unix(b.timestamp, 0).Format("2006-01-02 15:04:05"),
		b.nonce,
		b.merkleroot,
		b.BlockTransactions,
	)
}

// Calculating the hash of the data in merkle root
func hashCalculation(data string) string {
	hash_value := sha256.Sum256([]byte(data))
	hash_string := hex.EncodeToString(hash_value[:])
	return hash_string
}

// Merkle Root implementation
func merkleRoot(data []Transaction) *MerkleNode {
	//Calculating the hash of all the data and then appending in nodes of merkle tree it will be leaf nodes
	var nodes []*MerkleNode
	for _, val := range data {
		hash_data := &MerkleNode{hash: hashCalculation(val.Data)}
		nodes = append(nodes, hash_data)
	}
	//if there are more than 1 node than we can create merkle tree
	for len(nodes) > 1 {
		var level []*MerkleNode
		//Merkle tree is also known as binary hash tree so we have iterated i to i+=2
		for i := 0; i < len(nodes); i += 2 {
			left_node := nodes[i]
			right_node := left_node
			//If there is right tree than right node will be this
			if i+1 < len(nodes) {
				right_node = nodes[i+1]
			}
			parent_node := &MerkleNode{hash: hashCalculation(left_node.hash + right_node.hash), left: left_node, right: right_node}
			level = append(level, parent_node)
		}
		nodes = level
	}
	return nodes[0]
}

// Displaying merkle tree
func displayMerkleTree(root_node *MerkleNode, identation string) {
	if root_node != nil {
		fmt.Println(identation+"Hash_Value:", root_node.hash)
		if root_node.left != nil {
			displayMerkleTree(root_node.left, identation+"    ")
		}
		if root_node.right != nil {
			displayMerkleTree(root_node.right, identation+"    ")
		}
	}
}
