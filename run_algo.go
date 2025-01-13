package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func executePythonScript(filePath string, nClusters int) (string, error) {
	cmd := exec.Command("C:\\Python312\\python.exe", "kmeans.py", filePath, fmt.Sprintf("%d", nClusters))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	fmt.Printf("Executing Python script on %s with %d clusters...\n", filePath, nClusters)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute script: %v, stderr: %s", err, stderr.String())
	}

	return out.String(), nil
}

func run_algo() {
	datasets := map[string]int{
		"iris.csv": 3,
		"wine.csv": 3,
		"wdbc.csv": 2,
	}

	algoHash, err := calculateHash("kmeans.py", "sha256")
	if err != nil {
		fmt.Printf("Error calculating hash of kmeans.py: %v\n", err)
		return
	}

	for filePath, nClusters := range datasets {
		datasetHash, err := calculateHash(filePath, "sha256")
		if err != nil {
			fmt.Printf("Error calculating hash of %s: %v\n", filePath, err)
			return
		}

		output, err := executePythonScript(filePath, nClusters)
		if err != nil {
			log.Fatalf("Error running script for %s: %v", filePath, err)
		}

		//print the output
		//fmt.Println(output)

		// Parse the Python script output
		aiOutput, err := parseAIOutput(output)
		if err != nil {
			log.Fatalf("Error parsing script output for %s: %v", filePath, err)
		}

		// Create a transaction
		transaction := createTransaction(datasetHash, algoHash, &aiOutput)

		// Serialize transaction to JSON
		transactionJSON, err := json.MarshalIndent(transaction, "", "  ")
		if err != nil {
			log.Fatalf("Error serializing transaction for %s: %v", filePath, err)
		}

		fmt.Printf("Transaction for %s:\n%s\n", filePath, string(transactionJSON))
	}
}
