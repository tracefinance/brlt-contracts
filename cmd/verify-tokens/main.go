package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"vault0/internal/types"
	"vault0/internal/wire"
)

type TokenVerificationResult struct {
	Chain        types.Chain
	Token        types.Token
	URL          string
	IsVerified   bool
	Error        error
	ContractName string
}

func main() {
	// Initialize container with all dependencies
	container, err := wire.BuildContainer()
	if err != nil {
		log.Fatalf("Failed to build container: %v", err)
	}
	defer container.Core.DB.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Process each chain's tokens
	var wg sync.WaitGroup
	results := make(chan TokenVerificationResult, 100)

	// Get tokens for each chain
	chainList := container.Core.Chains.List()
	for _, chain := range chainList {
		tokens, err := container.Core.TokenStore.ListTokensByChain(ctx, chain.Type, 0, "")
		if err != nil {
			continue
		}

		wg.Add(1)
		go verifyTokens(ctx, &wg, results, chain, tokens.Items, container)
	}

	// Wait for all verifications to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Group results by blockchain
	groupedResults := make(map[types.ChainType][]TokenVerificationResult)
	for result := range results {
		groupedResults[result.Chain.Type] = append(groupedResults[result.Chain.Type], result)
	}

	// Print results grouped by blockchain
	printResults(groupedResults, container.Core.Chains)
}

func printResults(groupedResults map[types.ChainType][]TokenVerificationResult, chains *types.Chains) {
	// Sort network names for consistent output
	chainList := chains.List()
	networks := make([]string, 0, len(chainList))
	chainTypeToName := make(map[types.ChainType]string)
	nameToChainType := make(map[string]types.ChainType)

	for _, chain := range chainList {
		networks = append(networks, chain.Name)
		chainTypeToName[chain.Type] = chain.Name
		nameToChainType[chain.Name] = chain.Type
	}
	sort.Strings(networks)

	// Print results for each network
	for _, network := range networks {
		fmt.Printf("\n%s Network\n", network)
		fmt.Println(strings.Repeat("=", len(network)+8))
		fmt.Println()

		chainType := nameToChainType[network]
		results := groupedResults[chainType]

		if len(results) == 0 {
			fmt.Printf("No tokens found for %s\n", network)
			continue
		}

		// Sort results by symbol
		sort.Slice(results, func(i, j int) bool {
			return results[i].Token.Symbol < results[j].Token.Symbol
		})

		// Print token information
		for _, result := range results {
			if result.Error != nil {
				fmt.Printf("❌ %s (%s)\n", result.Token.Symbol, result.Token.Address)
				fmt.Printf("   Error: %v\n", result.Error)
				fmt.Printf("   URL: %s\n\n", result.URL)
				continue
			}

			var status string
			if result.Token.IsNative() {
				status = "🟢 Native Token"
			} else if result.IsVerified {
				status = "✅ Verified"
			} else {
				status = "⚠️  Unverified"
			}

			fmt.Printf("%s %s (%s)\n", status, result.Token.Symbol, result.Token.Address)
			if result.ContractName != "" {
				fmt.Printf("   Contract: %s\n", result.ContractName)
			}
			fmt.Printf("   URL: %s\n\n", result.URL)
		}
	}
}

func verifyTokens(
	ctx context.Context,
	wg *sync.WaitGroup,
	results chan<- TokenVerificationResult,
	chain types.Chain,
	tokens []types.Token,
	container *wire.Container,
) {
	defer wg.Done()

	// Get explorer for this chain
	explorer, err := container.Core.BlockExplorerFactory.NewExplorer(chain.Type)
	if err != nil {
		results <- TokenVerificationResult{
			Chain: chain,
			Error: fmt.Errorf("error getting explorer: %v", err),
		}
		return
	}

	// Check each token
	for _, token := range tokens {
		result := TokenVerificationResult{
			Chain: chain,
			Token: token,
			URL:   explorer.GetTokenURL(token.Address),
		}

		if token.IsNative() {
			results <- result
			continue
		}

		// Get contract information to verify it exists and is verified
		info, err := explorer.GetContract(ctx, token.Address)
		if err != nil {
			result.Error = err
			results <- result
			continue
		}

		result.IsVerified = info.IsVerified
		result.ContractName = info.ContractName
		results <- result
	}
}
