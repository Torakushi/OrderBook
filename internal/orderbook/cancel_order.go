package orderbook

import (
	"fmt"
	"strconv"
	"strings"
)

// CancelOrder represents a cancel order.
type CancelOrder struct {
	User        int
	UserOrderId int
}

// NewCancelOrderFromInstruction creates a new CancelOrder from a string.
// Instruction should have the form: 'C, user(int),userOrderId(int)'.
// It returns an error if it can't parse the string.
func NewCancelOrderFromInstruction(instruction string) (*CancelOrder, error) {
	params := strings.Split(instruction, ",")
	if len(params) != 3 {
		return nil,
			fmt.Errorf("Can't create new cancel order from instruction as it has less than 3 parameters: %q", instruction)
	}

	user, err := strconv.Atoi(params[1])
	if err != nil {
		return nil, err
	}

	userOrderId, err := strconv.Atoi(params[2])
	if err != nil {
		return nil, err
	}

	return &CancelOrder{
		User:        user,
		UserOrderId: userOrderId,
	}, nil
}

// GetIdentifier returns the identifier of an order 'User-UserOrderId'
func (co *CancelOrder) GetIdentifier() string {
	return fmt.Sprintf("%d-%d", co.User, co.UserOrderId)
}
