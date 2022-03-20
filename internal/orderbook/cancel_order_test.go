package orderbook_test

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"kraken/internal/orderbook"
)

func TestNewCancelOrderFromInstruction(t *testing.T) {
	assert := td.Assert(t)

	// When we are in NewOrderFromInstruction, orderbook should have remove all empty spaces
	_, err := orderbook.NewCancelOrderFromInstruction("C,1,2")
	assert.CmpNoError(err)

	_, err = orderbook.NewCancelOrderFromInstruction("C,2")
	assert.CmpError(err, `Can't create new cancel order from instruction as it has less than 3 parameters: "C,1"`)

}
