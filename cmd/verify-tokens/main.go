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
	"vault0/internal/core/db"
	"vault0/internal/core/tokenstore"
	"vault0/internal/logger"
	"vault0/internal/types"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger, err := logger.NewLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize database
	db, err := db.NewDatabase(cfg, logger)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create token store
	tokenStore := tokenstore.NewTokenStore(db)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Process each chain's tokens
	var wg sync.WaitGroup
	results := make(chan verificationResult, 100)

	// Get tokens for each chain
	chains, err := types.NewChains(cfg)
	if err != nil {
		log.Fatalf("Error creating chains: %v", err)
	}
	chainList := chains.List()
	for _, chain := range chainList {
		tokens, err := tokenStore.GetTokensByChain(ctx, chain.Type)
		if err != nil {
			log.Printf("Error getting tokens for %s: %v", chain.Type, err)
			continue
		}

		wg.Add(1)
		go verifyTokens(ctx, &wg, results, chain.Type, tokens)
	}

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
func printGroupedResults(groupedResults map[types.ChainType][]string, chains *types.Chains) {
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
	chainType types.ChainType,
	tokens []*types.Token,
) {
	defer wg.Done()

	// Create block explorer factory
	factory := blockexplorer.NewFactory(types.Chains{types.Chain{Type: chainType}}, nil)

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
		if token.IsNative() {
			url := explorer.GetTokenURL(token.Address)
			results <- verificationResult{
				chainType: chainType,
				message:   fmt.Sprintf("[+] [%s] %s (native token)\n    URL: %s", strings.ToUpper(string(chainType)), token.Symbol, url),
			}
			continue
		}

		// Get contract information to verify it exists and is verified
		info, err := explorer.GetContract(ctx, token.Address)
		if err != nil {
			if errors.IsRPCError(err) {
				log.Printf("RPC error: %v", err)
				continue
			}
			log.Printf("Failed to verify token %s: %v", token.Address, err)
			continue
		}

		url := explorer.GetTokenURL(token.Address)
		if info.IsVerified {
			results <- verificationResult{
				chainType: chainType,
				message: fmt.Sprintf("[+] [%s] %s (%s): Verified contract - %s\n    URL: %s",
					strings.ToUpper(string(chainType)),
					token.Symbol,
					token.Address,
					info.ContractName,
					url,
				),
			}
		} else {
			results <- verificationResult{
				chainType: chainType,
				message: fmt.Sprintf("[!] [%s] %s (%s): Unverified contract\n    URL: %s",
					strings.ToUpper(string(chainType)),
					token.Symbol,
					token.Address,
					url,
				),
			}
		}
	}
}
