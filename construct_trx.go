package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type AIOutput struct {
	ClusterCenters [][]float64    `json:"cluster_centers"`
	Inertia        float64        `json:"inertia"`
	ClusterSizes   map[string]int `json:"cluster_sizes"`
}

func parseAIOutput(input string) (AIOutput, error) {
	var result AIOutput

	// Regex for parsing cluster centers
	clusterCentersRegex := regexp.MustCompile(`Cluster Centers:\s+\[\[([\s\S]+?)\]\]`)
	clusterCentersMatch := clusterCentersRegex.FindStringSubmatch(input)
	if len(clusterCentersMatch) > 1 {
		clusterCentersString := strings.ReplaceAll(clusterCentersMatch[1], "\n", " ")
		clusterCentersString = strings.ReplaceAll(clusterCentersString, "][", "] [")
		lines := strings.Split(clusterCentersString, "] [")
		for _, line := range lines {
			line = strings.ReplaceAll(line, "[", "")
			line = strings.ReplaceAll(line, "]", "")
			line = strings.TrimSpace(line)
			values := strings.Fields(line)
			var center []float64
			for _, val := range values {
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return result, fmt.Errorf("error parsing cluster center value: %v", err)
				}
				center = append(center, f)
			}
			result.ClusterCenters = append(result.ClusterCenters, center)
		}
	}

	// Regex for parsing inertia
	inertiaRegex := regexp.MustCompile(`Inertia:\s+([\d.]+)`)
	inertiaMatch := inertiaRegex.FindStringSubmatch(input)
	if len(inertiaMatch) > 1 {
		inertia, err := strconv.ParseFloat(inertiaMatch[1], 64)
		if err != nil {
			return result, fmt.Errorf("error parsing inertia: %v", err)
		}
		result.Inertia = inertia
	}

	// Regex for parsing cluster sizes
	clusterSizesRegex := regexp.MustCompile(`Cluster Sizes:\s+\{([\s\S]+?)\}`)
	clusterSizesMatch := clusterSizesRegex.FindStringSubmatch(input)
	if len(clusterSizesMatch) > 1 {
		clusterSizesString := strings.TrimSpace(clusterSizesMatch[1])
		clusterSizes := strings.Split(clusterSizesString, ",")
		result.ClusterSizes = make(map[string]int)
		for _, size := range clusterSizes {
			parts := strings.Split(size, ":")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value, err := strconv.Atoi(strings.TrimSpace(parts[1]))
				if err != nil {
					return result, fmt.Errorf("error parsing cluster size value: %v", err)
				}
				result.ClusterSizes[key] = value
			}
		}
	}

	return result, nil
}

func createTransaction(datasetHash, algoHash string, output *AIOutput) *Transaction {
	return &Transaction{
		DatasetHash: datasetHash,
		AlgoHash:    algoHash,
		Output:      *output,
	}
}
