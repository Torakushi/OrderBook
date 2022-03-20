package orderbook

import (
	"bufio"
	"container/heap"
	"fmt"
	"strings"
)

// OrderBook is the main structure that represents the order book.
// It contains 2 queues (one for ask and one for bid), a boolean to indicate
// to trade or to reject orders that cross the book.
type OrderBook struct {
	AskQueue    *OrderQueue
	BidQueue    *OrderQueue
	ShouldTrade bool

	// For cancel, we store only if an order is a buy one (if not it is a sell by default)
	// to consume less memory
	// As well using an empty struct consumes less memory
	mapOrderIsBuy map[string]struct{}
}

// NewOrderBook create an order book that will be able to trade or not depending of
// the given 'shouldTrade'
func NewOrderBook(shouldTrade bool) *OrderBook {
	return &OrderBook{
		AskQueue:      NewOrderQueue(AskOrderType),
		BidQueue:      NewOrderQueue(BidOrderType),
		ShouldTrade:   shouldTrade,
		mapOrderIsBuy: map[string]struct{}{},
	}
}

// ProcessFromStringInstructions processes all the instructions until a Flush message.
// Then it returns all the output of the given instructions.
// It can process 3 types of instruction: 'N' (New or Modify), 'C' (Cancel) and 'F' (Flush).
func (ob *OrderBook) ProcessFromStringInstructions(instructions string) (string, error) {
	var result []string

	scanner := bufio.NewScanner(strings.NewReader(instructions))
L:
	for scanner.Scan() {
		instruction := strings.Replace(scanner.Text(), " ", "", -1)
		if instruction == "" {
			continue
		}

		switch instruction[0] {
		case 'N':
			order, err := NewOrderFromInstruction(instruction)
			if err != nil {
				return "", err
			}

			result = append(result, ob.processNewOrModifyOrder(order)...)

		case 'C':
			cancelOrder, err := NewCancelOrderFromInstruction(instruction)
			if err != nil {
				return "", err
			}

			result = append(result, ob.processCancelOrder(cancelOrder)...)

		case 'F':
			break L

		default:
			return "", fmt.Errorf("Unknown transaction type: %q", string(instruction[0]))
		}
	}

	return strings.Join(result, "\n"), nil
}

// processNewOrModifyOrder processes a NewOrModify order.
// It adds the order to the given queue depending of its side.
// If orders cross the book, it creates a Reject or Trade output depending if the
// order book can trade or not
func (ob *OrderBook) processNewOrModifyOrder(order *Order) []string {
	var queue, queueToCompare *OrderQueue

	isBuy := order.OrderSide == "B"
	if isBuy {
		queue = ob.BidQueue
		queueToCompare = ob.AskQueue
	} else {
		queue = ob.AskQueue
		queueToCompare = ob.BidQueue
	}

	// Check if the book is crossed
	orderToCompare := queueToCompare.Peak()
	if orderToCompare != nil &&
		orderToCompare.User != order.User &&
		((isBuy && (order.Price >= orderToCompare.Price)) || (!isBuy && (order.Price <= orderToCompare.Price))) {
		if ob.ShouldTrade {
			// Generate acknoledgement output
			result := []string{ob.generateAcknowledgmentOutput(order)}

			// Generate trade output
			return append(result, ob.generateTrade(order, queue, queueToCompare)...)

		} else {
			// Generate a reject output
			return []string{ob.generateRejectOutput(order)}
		}
	}

	// To check if the TOB changes
	oldTOB := queue.GetTOBInfo()

	heap.Push(queue, order)

	if isBuy {
		// To easily know if a given order is a Sell or a Buy
		ob.mapOrderIsBuy[order.GetIdentifier()] = struct{}{}
	}

	// Generate acknoledgement output
	result := []string{ob.generateAcknowledgmentOutput(order)}

	// If the TOB is modified, generate a change of TOB
	if oldTOB != queue.GetTOBInfo() {
		result = append(result, ob.generateTopOfBookChangeOutput(queue))
	}

	return result
}

// processCancelOrder processes a cancel order.
// It removes the given order on the queue, and then check if the TOB changes.
// If it changes, it publishes a TOB change output
func (ob *OrderBook) processCancelOrder(cancelOrder *CancelOrder) []string {
	identifier := cancelOrder.GetIdentifier()
	// Get the side
	var queue *OrderQueue
	if _, ok := ob.mapOrderIsBuy[identifier]; ok {
		queue = ob.BidQueue
	} else {
		queue = ob.AskQueue
	}

	// To check if the TOB changes
	oldTOB := queue.GetTOBInfo()

	// Remove order
	order := queue.Delete(identifier)
	if order == nil {
		return nil
	}

	// Acknowledge
	result := []string{ob.generateAcknowledgmentOutput(order)}

	// If TOB changes, generate a TOB change output
	if oldTOB != queue.GetTOBInfo() {
		result = append(result, ob.generateTopOfBookChangeOutput(queue))
	}

	return result
}

// Flush cleans all the order book
func (ob *OrderBook) Flush() {
	ob.BidQueue = NewOrderQueue(BidOrderType)
	ob.AskQueue = NewOrderQueue(AskOrderType)
	ob.mapOrderIsBuy = map[string]struct{}{}
}

// generateTrade processes a trade when an order crosses the book.
// It will trade the order until, the price of the opposite queue is to High or Low (depending of the
// order type) or if the order quantity becomes 0.
func (ob *OrderBook) generateTrade(order *Order, queue, queueToCompare *OrderQueue) []string {
	var result []string
	isBuy := order.OrderSide == "B"

	// Trade while quantity is greater than 0 and until the opposite queue
	// is not empty and the top price is not too high/low depending the order type
	for order.Quantity > 0 {
		if queueToCompare.Peak() == nil {
			break
		}

		// Price to low/high for the given ask/bid
		if (isBuy && (order.Price < queueToCompare.Peak().Price)) ||
			(!isBuy && (order.Price > queueToCompare.Peak().Price)) {
			break
		}

		// If the quantity of the order on the opposite queue is greater that than
		// the actual order quantity, do a partial trade.
		orderToCompare := heap.Pop(queueToCompare).(*Order)
		if orderToCompare.Quantity > order.Quantity {
			orderToCompare.Quantity -= order.Quantity
			heap.Push(queueToCompare, orderToCompare)
			return append(result,
				ob.generateTradeOutput(order, orderToCompare, orderToCompare.Price, order.Quantity),
				ob.generateTopOfBookChangeOutput(queueToCompare),
			)
		}

		// Else trade the entire order
		order.Quantity -= orderToCompare.Quantity
		result = append(result, ob.generateTradeOutput(order, orderToCompare, orderToCompare.Price, orderToCompare.Quantity))
	}

	// The TOB of the opposite queue necesserly changed
	result = append(result, ob.generateTopOfBookChangeOutput(queueToCompare))

	// If we cannot trade all our quantity, push back the order in the right queue and
	// change the TOB
	if order.Quantity > 0 {
		heap.Push(queue, order)
		result = append(result, ob.generateTopOfBookChangeOutput(queue))
	}

	return result
}

// generateAcknowledgmentOutput generates an aknowledgement output
func (ob *OrderBook) generateAcknowledgmentOutput(order *Order) string {
	return fmt.Sprintf("A, %d, %d", order.User, order.UserOrderId)
}

// generateTopOfBookChangeOutput generates an TOB changes output
func (ob *OrderBook) generateTopOfBookChangeOutput(queue *OrderQueue) string {
	return fmt.Sprintf("B, %s", queue.GetTOBInfo())
}

// generateRejectOutput generates an reject output
func (ob *OrderBook) generateRejectOutput(order *Order) string {
	return fmt.Sprintf("R, %d, %d", order.User, order.UserOrderId)
}

// generateTradeOutput generates an trade output
func (ob *OrderBook) generateTradeOutput(buyOrder, sellOrder *Order, price, quantity int) string {
	return fmt.Sprintf("T, %d, %d, %d, %d, %d, %d",
		buyOrder.User,
		buyOrder.UserOrderId,
		sellOrder.User,
		sellOrder.UserOrderId,
		price,
		quantity,
	)
}
