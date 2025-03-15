package main

import (
	"context"
	"fmt"
	"log"
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
	results := make(chan string, 100)

	// Process Ethereum tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypeEthereum, cfg.Tokens.Ethereum)

	// Process Polygon tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypePolygon, cfg.Tokens.Polygon)

	// Process Base tokens
	wg.Add(1)
	go verifyTokens(ctx, &wg, results, factory, types.ChainTypeBase, cfg.Tokens.Base)

	// Print results as they come in
	go func() {
		for msg := range results {
			fmt.Println(msg)
		}
	}()

	// Wait for all verifications to complete
	wg.Wait()
	close(results)
}

func verifyTokens(
	ctx context.Context,
	wg *sync.WaitGroup,
	results chan<- string,
	factory blockexplorer.Factory,
	chainType types.ChainType,
	tokens []config.TokenConfig,
) {
	defer wg.Done()

	// Get explorer for this chain
	explorer, err := factory.GetExplorer(chainType)
	if err != nil {
		results <- fmt.Sprintf("Error getting explorer for %s: %v", chainType, err)
		return
	}
	defer explorer.Close()

	// Check each token
	for _, token := range tokens {
		if token.Type == "native" {
			results <- fmt.Sprintf("✅ [%s] %s (native token)", strings.ToUpper(string(chainType)), token.Symbol)
			continue
		}

		// Get contract information to verify it exists and is verified
		info, rawResponse, err := explorer.(*blockexplorer.EVMExplorer).VerifyContractWithRawResponse(ctx, token.Address)
		if err != nil {
			results <- fmt.Sprintf("❌ [%s] %s (%s): Error - %v\nRaw Response: %s",
				strings.ToUpper(string(chainType)),
				token.Symbol,
				token.Address,
				err,
				rawResponse,
			)
			continue
		}

		if info.IsVerified {
			results <- fmt.Sprintf("✅ [%s] %s (%s): Verified contract - %s",
				strings.ToUpper(string(chainType)),
				token.Symbol,
				token.Address,
				info.ContractName,
			)
		} else {
			results <- fmt.Sprintf("⚠️ [%s] %s (%s): Unverified contract",
				strings.ToUpper(string(chainType)),
				token.Symbol,
				token.Address,
			)
		}
	}
}
