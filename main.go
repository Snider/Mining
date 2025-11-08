// Demo main.go for development and testing
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Snider/Mining/pkg/mining"
)

func main() {
	fmt.Println("Mining Package Demo")
	fmt.Println("===================")
	fmt.Println()

	// Create a new manager
	manager := mining.NewManager()
	fmt.Println("✓ Created new mining manager")

	// Start a few miners
	configs := []mining.MinerConfig{
		{
			Name:      "bitcoin-miner-1",
			Algorithm: "sha256",
			Pool:      "pool.bitcoin.com",
			Wallet:    "bc1qxy2kgdygjrsqtzq2n0yrf2493p83kkfjhx0wlh",
		},
		{
			Name:      "ethereum-miner-1",
			Algorithm: "ethash",
			Pool:      "pool.ethereum.org",
			Wallet:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		},
	}

	var minerIDs []string
	for _, config := range configs {
		miner, err := manager.StartMiner(config)
		if err != nil {
			log.Fatalf("Failed to start miner: %v", err)
		}
		minerIDs = append(minerIDs, miner.ID)
		fmt.Printf("✓ Started miner: %s (ID: %s)\n", miner.Name, miner.ID)
		time.Sleep(10 * time.Millisecond) // Small delay for unique IDs
	}

	fmt.Println()

	// Update hash rates
	hashRates := []float64{150.5, 320.75}
	for i, id := range minerIDs {
		err := manager.UpdateHashRate(id, hashRates[i])
		if err != nil {
			log.Fatalf("Failed to update hash rate: %v", err)
		}
		fmt.Printf("✓ Updated hash rate for %s: %.2f H/s\n", id, hashRates[i])
	}

	fmt.Println()

	// List all miners
	fmt.Println("Active Miners:")
	fmt.Println("--------------")
	miners := manager.ListMiners()
	for _, miner := range miners {
		fmt.Printf("  %s: %s (%.2f H/s, %s)\n",
			miner.ID,
			miner.Name,
			miner.HashRate,
			miner.Status,
		)
	}

	fmt.Println()

	// Get specific miner status
	if len(minerIDs) > 0 {
		miner, err := manager.GetMiner(minerIDs[0])
		if err != nil {
			log.Fatalf("Failed to get miner: %v", err)
		}
		fmt.Printf("Detailed Status for %s:\n", miner.Name)
		fmt.Printf("  ID:         %s\n", miner.ID)
		fmt.Printf("  Status:     %s\n", miner.Status)
		fmt.Printf("  Start Time: %s\n", miner.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Hash Rate:  %.2f H/s\n", miner.HashRate)
		fmt.Println()
	}

	// Stop a miner
	if len(minerIDs) > 0 {
		err := manager.StopMiner(minerIDs[0])
		if err != nil {
			log.Fatalf("Failed to stop miner: %v", err)
		}
		fmt.Printf("✓ Stopped miner: %s\n", minerIDs[0])
	}

	fmt.Println()
	fmt.Printf("Demo completed successfully! Version: %s\n", mining.GetVersion())
}
