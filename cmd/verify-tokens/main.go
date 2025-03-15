package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/types"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create chains map
	chains := types.Chains{
		types.ChainTypeEthereum: {
			Type: types.ChainTypeEthereum,
			ID:   cfg.Blockchains.Ethereum.ChainID,
			Name: "Ethereum",
		},
		types.ChainTypePolygon: {
			Type: types.ChainTypePolygon,
			ID:   cfg.Blockchains.Polygon.ChainID,
			Name: "Polygon",
		},
		types.ChainTypeBase: {
			Type: types.ChainTypeBase,
			ID:   cfg.Blockchains.Base.ChainID,
			Name: "Base",
		},
	}

	// Create block explorer factory
	factory := blockexplorer.NewFactory(chains, cfg)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Process each chain's tokens
	var wg sync.WaitGroup
	results := make(chan verificationResult, 100)

	// Process Ethereum tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypeEthereum, cfg.Tokens.Ethereum)

	// Process Polygon tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypePolygon, cfg.Tokens.Polygon)

	// Process Base tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypeBase, cfg.Tokens.Base)

	// Collect and group results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Group results by blockchain
	groupedResults := make(map[types.ChainType][]string)
	for result := range results {
		groupedResults[result.chainType] = append(groupedResults[result.chainType], result.message)
	}

	// Print results grouped by blockchain
	printGroupedResults(groupedResults, chains)
}

// verificationResult represents a single token verification result
type verificationResult struct {
	chainType types.ChainType
	message   string
}

// printGroupedResults prints verification results grouped by blockchain
func printGroupedResults(groupedResults map[types.ChainType][]string, chains types.Chains) {
	// Sort network names for consistent output
	networks := make([]string, 0, len(groupedResults))
	chainTypeToName := make(map[types.ChainType]string)
	nameToChainType := make(map[string]types.ChainType)

	for chainType, chain := range chains {
		networks = append(networks, chain.Name)
		chainTypeToName[chainType] = chain.Name
		nameToChainType[chain.Name] = chainType
	}
	sort.Strings(networks)

	// Print results for each network
	for _, network := range networks {
		fmt.Printf("\n> %s Network <\n\n", network)

		// Print messages for this network
		chainType := nameToChainType[network]
		results := groupedResults[chainType]
		if len(results) == 0 {
			fmt.Printf("  No tokens verified for %s\n", network)
		} else {
			for _, msg := range results {
				fmt.Printf("  %s\n", msg)
			}
		}
	}
}

func verifyTokens(
	ctx context.Context,
	wg *sync.WaitGroup,
	results chan<- verificationResult,
	factory blockexplorer.Factory,
	chainType types.ChainType,
	tokens []config.TokenConfig,
) {
	defer wg.Done()

	// Get explorer for this chain
	explorer, err := factory.GetExplorer(chainType)
	if err != nil {
		results <- verificationResult{
			chainType: chainType,
			message:   fmt.Sprintf("Error getting explorer for %s: %v", chainType, err),
		}
		return
	}
	defer explorer.Close()

	// Check each token
	for _, token := range tokens {
		if token.Type == "native" {
			results <- verificationResult{
				chainType: chainType,
				message:   fmt.Sprintf("[+] [%s] %s (native token)", strings.ToUpper(string(chainType)), token.Symbol),
			}
			continue
		}

		// Get contract information to verify it exists and is verified
		info, err := explorer.GetContract(ctx, token.Address)
		if err != nil {
			var evmErr *blockexplorer.EVMExplorerError
			rawResponse := ""
			if errors.As(err, &evmErr) {
				rawResponse = fmt.Sprintf("\nRaw Response: %s", evmErr.RawResponse)
			}

			results <- verificationResult{
				chainType: chainType,
				message: fmt.Sprintf("[-] [%s] %s (%s): Error - %v%s",
					strings.ToUpper(string(chainType)),
					token.Symbol,
					token.Address,
					err,
					rawResponse,
				),
			}
			continue
		}

		if info.IsVerified {
			results <- verificationResult{
				chainType: chainType,
				message: fmt.Sprintf("[+] [%s] %s (%s): Verified contract - %s",
					strings.ToUpper(string(chainType)),
					token.Symbol,
					token.Address,
					info.ContractName,
				),
			}
		} else {
			results <- verificationResult{
				chainType: chainType,
				message: fmt.Sprintf("[!] [%s] %s (%s): Unverified contract",
					strings.ToUpper(string(chainType)),
					token.Symbol,
					token.Address,
				),
			}
		}
	}
}
