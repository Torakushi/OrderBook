package main

import (
	"fmt"
	"log"

	"kraken/internal/orderbook"
)

func main() {

	scenarios, err := orderbook.GetScenarios("internal/orderbook/testdata")
	if err != nil {
		log.Fatalf("Error when reading scenarios: %s", err.Error())
	}

	for i, scenario := range scenarios {
		fmt.Println(scenario.Description)
		ob := orderbook.NewOrderBook(scenario.ShouldTrade)
		output, err := ob.ProcessFromStringInstructions(scenario.Instructions)
		if err != nil {
			log.Fatalf("Error when processing scenario %d: %s", i+1, err.Error())
		}
		fmt.Printf("%s\n\n", output)
	}
}
