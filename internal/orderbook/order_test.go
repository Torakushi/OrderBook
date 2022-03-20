package orderbook_test

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"kraken/internal/orderbook"
)

func TestNewOrderFromInstruction(t *testing.T) {
	assert := td.Assert(t)

	// When we are in NewOrderFromInstruction, orderbook should have remove all empty spaces
	_, err := orderbook.NewOrderFromInstruction("N,1,IBM,10,100,B,1")
	assert.CmpNoError(err)

	_, err = orderbook.NewOrderFromInstruction("N,2")
	assert.CmpError(err, `Can't create new cancel order from instruction as it has less than 7 parameters: "N,2"`)

}
