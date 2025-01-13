package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
)

func retrieve_from_ipfs(cid string, fileName string) error {
	// Connect to the local IPFS daemon
	sh := shell.NewShell("localhost:5001")

	var buf bytes.Buffer
	r, err := sh.Cat(cid)
	if err != nil {
		return fmt.Errorf("failed to retrieve file: %w", err)
	}
	_, err = io.Copy(&buf, r)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Save the file locally
	err = os.WriteFile(fileName, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	fmt.Printf("File retrieved and saved as %s\n", fileName)
	return nil
}

func retrieve_all_files() {
	cid_iris := "QmWb4hxStTFqfs53nfpdpLKCPNPCesKRxfC5RkBP3ZwKmv"
	cid_wdbc := "QmZV3Mm9BkemcZizCnpby9Rx2Mbi2MoXqhZ22L22nS1bf7"
	cid_wine := "QmfTjNwLEB9Rc4T4Jpvs4mEg2RqvRVNVDGpXQVVR3CLJvP"
	cid_algo := "QmQzE8z9V3akLFm8xLXWavqCVcgYtxuZeTyJFz8MVzZ7we"

	if err := retrieve_from_ipfs(cid_iris, "iris.csv"); err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := retrieve_from_ipfs(cid_wdbc, "wdbc.csv"); err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := retrieve_from_ipfs(cid_wine, "wine.csv"); err != nil {
		log.Fatalf("Error: %v", err)
	}
	if err := retrieve_from_ipfs(cid_algo, "kmeans.py"); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func calculateHash(filePath string, algorithm string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var hash []byte
	if algorithm == "sha256" {
		h := sha256.New()
		if _, err := io.Copy(h, file); err != nil {
			return "", err
		}
		hash = h.Sum(nil)
	} else if algorithm == "sha512" {
		h := sha512.New()
		if _, err := io.Copy(h, file); err != nil {
			return "", err
		}
		hash = h.Sum(nil)
	} else {
		return "", fmt.Errorf("unsupported algorithm")
	}

	return fmt.Sprintf("%x", hash), nil
}

func all_files_hash() {
	filePath := "iris.csv"
	hash, err := calculateHash(filePath, "sha256")
	if err != nil {
		fmt.Printf("Error calculating hash of iris dataset: %v\n", err)
		return
	}

	fmt.Printf("SHA-256 hash of %s: %s\n", filePath, hash)

	filePath = "wdbc.csv"
	hash, err = calculateHash(filePath, "sha256")
	if err != nil {
		fmt.Printf("Error calculating hash of wdbc dataset: %v\n", err)
		return
	}

	fmt.Printf("SHA-256 hash of %s: %s\n", filePath, hash)

	filePath = "wine.csv"
	hash, err = calculateHash(filePath, "sha256")
	if err != nil {
		fmt.Printf("Error calculating hash of wine dataset: %v\n", err)
		return
	}

	fmt.Printf("SHA-256 hash of %s: %s\n", filePath, hash)

	filePath = "kmeans.py"
	hash, err = calculateHash(filePath, "sha256")
	if err != nil {
		fmt.Printf("Error calculating hash of kmeans.py: %v\n", err)
		return
	}

	fmt.Printf("SHA-256 hash of %s: %s\n", filePath, hash)
}
