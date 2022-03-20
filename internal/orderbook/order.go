package orderbook

import (
	"fmt"
	"strconv"
	"strings"
)

// Order is the main structure to represents an order
type Order struct {
	User        int
	Symbol      string
	Price       int
	Quantity    int
	UserOrderId int
	OrderSide   string

	index int // It will be used by the priority queue
	time  int // To track which order is the oldest
}

// Create a new order from a string.
// Instruction should be: N, user(int),symbol(string),price(int),qty(int),side(char B or S),userOrderId(int).
// It returns an error if it can't parse the string.
func NewOrderFromInstruction(instruction string) (*Order, error) {
	params := strings.Split(instruction, ",")
	if len(params) != 7 {
		return nil,
			fmt.Errorf("Can't create new order from instruction as it has less than 7 parameters: %q", instruction)
	}

	user, err := strconv.Atoi(params[1])
	if err != nil {
		return nil, err
	}

	symbole := params[2]

	price, err := strconv.Atoi(params[3])
	if err != nil {
		return nil, err
	}

	quantity, err := strconv.Atoi(params[4])
	if err != nil {
		return nil, err
	}

	side := params[5]
	if side != "S" && side != "B" {
		return nil, fmt.Errorf("Unknown side for order: %q", side)
	}

	userOrderId, err := strconv.Atoi(params[6])
	if err != nil {
		return nil, err
	}

	return &Order{
		User:        user,
		Symbol:      symbole,
		Price:       price,
		Quantity:    quantity,
		OrderSide:   side,
		UserOrderId: userOrderId,
	}, nil
}

// GetIdentifier returns the identifier of an order 'User-UserOrderId'
func (o *Order) GetIdentifier() string {
	return fmt.Sprintf("%d-%d", o.User, o.UserOrderId)
}
