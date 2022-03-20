package orderbook

import (
	"container/heap"
	"fmt"
)

// Type of order (BID or ASK)
type OrderType int

const (
	AskOrderType OrderType = iota
	BidOrderType
)

// CompareFunc is a function to implement the 'Less' function for the sort interface.
// It is generic because Ask and Bid are sorted differently.
type CompareFunc func(i, j int) bool

// OrderQueue represents a priority queue (heap) which contains all orders of a type (Ask or Bid).
// This queue is sorted by price and ultimately by time.
// The order dependsof the order type: Highest to smallest for Bids and conversely for Asks.
// It contains as well two maps in order to retrieve easily orders by identifier (userID-userOrderID),
// or to retrieve the total quantity for a given price and a given user (Easy to detect changement of volume).
type OrderQueue struct {
	orders                []*Order
	compareFunc           CompareFunc       // For the 'Less(i,j)' implemementation
	mapPriceToQuantity    map[int]int       // For modification of the quantity (and TOB status)
	mapSearchByIdentifier map[string]*Order // Use to make 'Cancel' order quicker
	OrderSide             string            // Order type
	timer                 int               // Simulate a timer in order to know which order is older
}

// NewOrderQueue creates a new priority queue (heap) for Ask or Bid orders depending
// of the OrderType given.
func NewOrderQueue(orderType OrderType) *OrderQueue {
	var cf CompareFunc
	var side string
	if orderType == AskOrderType {
		cf = askCompareFunc()
		side = "S"
	} else {
		cf = bidCompareFunc()
		side = "B"
	}

	return &OrderQueue{
		compareFunc:           cf,
		mapPriceToQuantity:    map[int]int{},
		mapSearchByIdentifier: map[string]*Order{},
		OrderSide:             side,
	}
}

// askCompareFunc is the function to implement 'Less()' function of the sort interface.
// It sorted orders from Lowest Price to Highest.
func askCompareFunc() CompareFunc {
	return func(i, j int) bool {
		return i < j
	}
}

// bidCompareFunc is the function to implement 'Less()' function of the sort interface.
// It sorted orders from Highest Price to Lowest.
func bidCompareFunc() CompareFunc {
	return func(i, j int) bool {
		return i > j
	}
}

// Len returns the queue lenght.
// To implement sort interface
func (oq OrderQueue) Len() int { return len(oq.orders) }

// Less indicates how to sort the queue.
// It sorts the queue using the 'CompareFunc' of the OrderQueue.
// For order with same price, the oldest order will be in front.
// To implement sort interface
func (oq OrderQueue) Less(i, j int) bool {
	if oq.orders[i].Price != oq.orders[j].Price {
		return oq.compareFunc(oq.orders[i].Price, oq.orders[j].Price)
	}
	return oq.orders[i].time < oq.orders[j].time
}

// Swap swaps two elements at given position.
// To implement sort interface.
func (oq OrderQueue) Swap(i, j int) {
	oq.orders[i], oq.orders[j] = oq.orders[j], oq.orders[i]
	oq.orders[i].index = i
	oq.orders[j].index = j
}

// Push put an element in the priority queue and then re-heapify the queue.
// The complexity is O(log(n)).
// To implement heap interface.
func (oq *OrderQueue) Push(x interface{}) {
	n := len(oq.orders)
	order := x.(*Order)

	// Populate maps in order to retrieve easily the order
	oq.addToMaps(order)

	order.index = n

	// set time of order and increment the timer
	order.time = oq.timer
	oq.timer++

	oq.orders = append(oq.orders, order)
}

// Delete removes an element in the priority queue and then re-heapify the queue.
// The complexity is O(log(n)).
// To implement heap interface.
func (oq *OrderQueue) Delete(orderIdentifier string) *Order {
	if o, ok := oq.mapSearchByIdentifier[orderIdentifier]; ok {
		// Call Pop() so don't need to call 'deleteFromMaps'
		return heap.Remove(oq, o.index).(*Order)
	}

	return nil
}

// Peak retrieves the first element  of the heap.
// It is always the first element of the underlying list of the queue (see Pop() implementation)
func (oq *OrderQueue) Peak() *Order {
	if len(oq.orders) > 0 {
		return oq.orders[0]
	}

	return nil
}

// Pop returns the first element of the priority queue
// To implement heap interface.
func (oq *OrderQueue) Pop() interface{} {
	old := oq.orders
	n := len(old)
	order := old[n-1]
	old[n-1] = nil   // avoid memory leak
	order.index = -1 // for safety
	oq.orders = old[0 : n-1]
	oq.deleteFromMaps(order)
	return order
}

// GetTOBInfo returns information of the top of the book side.
// It returns the side, the price and the sum of quantities (volume) for the top order.
// If the queue is empty, it returns 'Side, -, -'.
func (oq *OrderQueue) GetTOBInfo() string {
	o := oq.Peak()
	if o == nil {
		return fmt.Sprintf("%s, -, -", oq.OrderSide)
	}

	return fmt.Sprintf("%s, %d, %d",
		oq.OrderSide,
		o.Price,
		oq.mapPriceToQuantity[o.Price], // Get total quantity
	)
}

// addToMaps adds the orders to all maps in order to retrieve it easily
// using Identifier or to retrieve the volume of a kind of order using price.
func (oq *OrderQueue) addToMaps(o *Order) {
	oq.mapSearchByIdentifier[o.GetIdentifier()] = o
	oq.mapPriceToQuantity[o.Price] += o.Quantity
}

// deleteFromMaps removes the order of all the maps.
func (oq *OrderQueue) deleteFromMaps(o *Order) {
	delete(oq.mapSearchByIdentifier, o.GetIdentifier())
	if oq.mapPriceToQuantity[o.Price] == 0 {
		return
	}
	oq.mapPriceToQuantity[o.Price] -= o.Quantity
}
