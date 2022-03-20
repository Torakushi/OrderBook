package orderbook_test

import (
	"container/heap"
	"testing"

	"github.com/maxatome/go-testdeep/td"

	"kraken/internal/orderbook"
)

func TestOrderQueue_PushOrder(t *testing.T) {
	assert := td.Assert(t)

	// Create a list of orders to push on our queues
	orders := []*orderbook.Order{
		{
			User:        1,
			Symbol:      "IBM",
			Price:       10,
			Quantity:    100,
			UserOrderId: 1,
		},
		{
			User:        1,
			Symbol:      "IBM",
			Price:       7,
			Quantity:    100,
			UserOrderId: 1,
		},
		{
			User:        1,
			Symbol:      "IBM",
			Price:       9,
			Quantity:    100,
			UserOrderId: 1,
		},
	}

	// Test an AskOrderQueue
	// On a AskOrderQueue, orders are sorted from the lowest ask price to the highest one
	askOrderQueue := orderbook.NewOrderQueue(orderbook.AskOrderType)

	for _, order := range orders {
		heap.Push(askOrderQueue, order)
	}

	// Three elements on the queue
	assert.Cmp(askOrderQueue.Len(), 3)

	// On our case it should be: [7,9,10] (ordered by Price)
	assert.Cmp(heap.Pop(askOrderQueue), orders[1])
	assert.Cmp(heap.Pop(askOrderQueue), orders[2])
	assert.Cmp(heap.Pop(askOrderQueue), orders[0])

	// Try Delete
	for _, order := range orders {
		heap.Push(askOrderQueue, order)
	}

	askOrderQueue.Delete(orders[0].GetIdentifier())
	assert.Cmp(askOrderQueue.Len(), 2)

	// Check Peak()
	assert.Cmp(askOrderQueue.Peak(), orders[1])

	// Test a BidOrderQueue
	// On a BidOrderQueue, orders are sorted from the highest bid price to the lowest one
	bidOrderQueue := orderbook.NewOrderQueue(orderbook.BidOrderType)

	for _, order := range orders {
		heap.Push(bidOrderQueue, order)
	}

	// Three elements on the queue
	assert.Cmp(bidOrderQueue.Len(), 3)

	// On our case it should be: [10,9,7] (ordered by Price)
	assert.Cmp(heap.Pop(bidOrderQueue), orders[0])
	assert.Cmp(heap.Pop(bidOrderQueue), orders[2])
	assert.Cmp(heap.Pop(bidOrderQueue), orders[1])
}

func TestOrderQueue_PushOrderWithSamePriceAndUser(t *testing.T) {
	assert := td.Assert(t)

	order := &orderbook.Order{
		User:        1,
		Symbol:      "IBM",
		Price:       10,
		Quantity:    100,
		UserOrderId: 1,
	}

	// We use AskOrderQueue for the test
	// We could use a BidOrderQueue since the code involved here is the same
	// We could use a custom OrderQueue as well
	askOrderQueue := orderbook.NewOrderQueue(orderbook.AskOrderType)

	// Imagine that the same user creates two orders with the same Price/Symbole
	heap.Push(askOrderQueue, order)
	heap.Push(askOrderQueue, order)

	assert.Cmp(askOrderQueue.Len(), 2)

	// The TOB should have a volume of 200
	assert.Cmp(askOrderQueue.GetTOBInfo(), "S, 10, 200")
}

func TestOrderQueue_PushOrderWithSamePriceButDifferentUser(t *testing.T) {
	assert := td.Assert(t)

	orders := []*orderbook.Order{
		{
			User:        1,
			Symbol:      "IBM",
			Price:       10,
			Quantity:    100,
			UserOrderId: 1,
		},
		{
			User:        2,
			Symbol:      "IBM",
			Price:       10,
			Quantity:    100,
			UserOrderId: 1,
		},
	}

	// We use AskOrderQueue for the test
	// We could use a BidOrderQueue since the code involved here is the same
	// We could use a custom OrderQueue as well
	askOrderQueue := orderbook.NewOrderQueue(orderbook.AskOrderType)

	// Imagine that two users did an order with the same Price/Symbole
	for _, order := range orders {
		heap.Push(askOrderQueue, order)
	}

	// Two elements on the queue
	assert.Cmp(askOrderQueue.Len(), 2)

	// As the price is the same for both orders, the queue should be sorted by time:
	// oldest order should be first (FIFO)
	assert.Cmp(heap.Pop(askOrderQueue), orders[0])
	assert.Cmp(heap.Pop(askOrderQueue), orders[1])
}
