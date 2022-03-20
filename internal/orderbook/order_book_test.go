package orderbook_test

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"kraken/internal/orderbook"
)

func TestOrderBook(t *testing.T) {
	assert, require := td.AssertRequire(t)

	// Retrieve all scenarios in 'inout.txt' and their outputs in 'output.txt'
	scenarios, err := orderbook.GetScenarios("testdata")
	require.CmpNoError(err)

	for _, s := range scenarios {
		assert.RunAssertRequire(s.Description,
			func(assert, require *td.T) {
				ob := orderbook.NewOrderBook(s.ShouldTrade)

				output, err := ob.ProcessFromStringInstructions(s.Instructions)
				assert.CmpNoError(err)
				assert.Cmp(output, s.Output)
			})
	}

}
