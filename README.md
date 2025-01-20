# Distributed AI Blockchain with IPFS

This project demonstrates the integration of a deterministic AI algorithm, InterPlanetary File System (IPFS), and blockchain technology using a peer-to-peer (P2P) network built in Golang. The system enables distributed execution of an AI algorithm (K-means clustering) and verifies the results through blockchain consensus.

---

## Features

- **IPFS Integration**: The AI algorithm and datasets are stored and retrieved using IPFS with their Content Identifier (CID).
- **P2P Network**: A decentralized peer-to-peer network is implemented using Golang's network programming.
- **AI Algorithm Execution**: The Bootstrap node retrieves data from IPFS and runs the K-means clustering algorithm on the datasets.
- **Blockchain Mining**: Results from the AI algorithm are hashed into transactions, mined into blocks, and propagated through the network.
- **Verification and Consensus**: Nodes independently verify the algorithm results by retrieving data from IPFS and re-running the algorithm. Verified blocks are added to the blockchain.
- **Merkle Tree Generation**: The system generates a Merkle tree for transaction validation within each block.

---

## Architecture

### System Workflow
1. **Upload to IPFS**: The AI algorithm and datasets are uploaded to IPFS, and their CID is recorded.
2. **Bootstrap Node**: 
   - Retrieves the algorithm and datasets using the CIDs.
   - Executes the K-means clustering algorithm.
   - Creates transactions by hashing the algorithm and output.
   - Propagates transactions to the P2P network.
3. **Mining Process**:
   - Nodes receive transactions and begin mining blocks.
   - Miners validate transactions by re-running the algorithm using data from IPFS.
   - After consensus, the block is added to the blockchain.
4. **Blockchain and Merkle Tree**:
   - The final blockchain with all mined and verified blocks is printed.
   - A Merkle tree is generated for transaction validation.

### Components
- **Golang P2P Network**: Implements communication between nodes.
- **IPFS Integration**: Handles data storage and retrieval.
- **K-means Clustering Algorithm**: Executes AI on datasets.
- **Blockchain**: Stores verified transactions.
- **Merkle Tree**: Ensures transaction integrity.

---

## Requirements

- [Golang](https://golang.org/) (1.16 or higher)
- [IPFS](https://ipfs.io/)
- Git

---

## How to start?
- Clone the repo.
- Install dependencies listed in requirements.txt.
- Setup your private IPFS or IPFS desktop.
- Run `go build`.
- Run `./chain`.

---

## Usage

1. Upload datasets and the AI algorithm to IPFS.
2. Note the CID for each file.
3. Configure the Bootstrap node with the CIDs.
4. Start the network and observe the following:
   - Execution of the K-means clustering algorithm.
   - Creation of transactions and mining of blocks.
   - Validation and propagation of verified blocks.
   - Printing of the final blockchain and Merkle tree.

---

## License

This project is licensed under the [MIT License](LICENSE).
