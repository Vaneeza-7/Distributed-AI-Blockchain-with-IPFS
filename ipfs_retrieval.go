package main

import (
	"bytes"
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
	cid_algo := "QmR32KB2MVzTHVAnTgovPERtzhwkYVUzVwfoz6HATn3QJJ"

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
