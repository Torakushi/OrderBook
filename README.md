# OrderBook Project

## How it works

### Input

It take a bunch of instructions which can be:

- New order: `N	user(int)	symbol(string)	price(int)	qty(int)	side(char B or S)	userOrderId(int)`
				
- Cancel order: `C	user(int)	userOrderId(int)	`						

- Flush orderbook: `F`

Notes:

- Price is 0 for market order		
- Between scenarios flush order books		

### Output

Publish order or cancel acknowledgement format: `A, userId, userOrderId `

Publish changes in Top Of Book per side using format, use ‘-‘ for side elimination: 
`B, side (B orS),price, totalQuantity`

Publish rejects for orders that would make or book crossed: 
`R, userId, userOrderId `

or if we want to trade orders that crossed the book:

Publish trades (matched orders) format: 
`T, userIdBuy, userOrderIdBuy, userIdSell, userOrderIdSell,price,quantity` 


## How to build

Use the docker file:

`sudo docker build --progress=plain -t kraken .`

It builds the docker file **and runs the tests** (I added two tests with partial orders)

All this exercise is **test-centered**, I prefered wrapping all scenarios in unit tests (like the exercise advises)
instead of having a main. It has more sense, so that we can compare automaticaly inputs and outputs.
I worked mainly on unit testing and not on the main as it does the same thing but without a stdout

The main simply processes and prints all outputs of scenarios that we can find in `internal/orderbook/testdata`.

If you want to change input/output, you have to change `input.txt` and `output.txt` in `internal/orderbook/testdata`

If you really want to have all the output:

`sudo docker run kraken`

I used `go1.18` to use `bytes.Cut()` (I Could have fun with generics as well maybe but don't have the time to
think about it)

## Test

My unit tests use https://github.com/maxatome/go-testdeep a very nice testing dll

I created one input.txt (with all scenarios) and one output.txt (with outputs).

A scenario looks like this:
input:
```
# B Scenario X: My Description
N, 1, IBM, 10, 100, B, 1
.....

# B Scenario X+1: My Description
```
**`B` is equal to 1 if it is possible to trade (0 otherwise)**

output:
```
# Scenario 1
A, 1, 2
.....
```

My test runs all of them (using the description as a sub-test name) and compares the
result to the output.

## Structures Choices

## Heap (Priority queue) (order_queue.go)

An order book is composed by two sides (Ask/Bid).
To represent a side of my book, I chose a heap (more precisely, a priority queue)
implementing the [ heap interface ](https://pkg.go.dev/container/heap).

This choice sounds good because after each insertion (new order) or cancelation (cancel order), we need to sort again
our queue... Using a heap, insertion and deletion have a complexity of **O(log(n))**.
Furthermore, I am only interested, most of the time, with the TOB,
so maintaining a priority queue (heap) at low cost seems a good idea.

Additionally, to retrieve easily orders on our queue for cancelation order for example,
I use a map (O(1) + pointer use for memory). 

Another challenge was to check the TOB status without the need to "pop" a part of the queue.
Indeed, for example to detect a volume change in the TOB we would need to pop a part of the queue to have
the total quantity for a user and at a given price.
Instead, I use a `map[string]int` which contains the total quantity of user's orders at a given price.
I update this map at each Pop or Push if needed.

So retrieving the TOB status is easy: 
   - Check the first order of the queue and retrieve its information
   - Check the total quantity using the previous map for the key `User-Price` (PriceUserIdentifier)

For example, given two orders:

N, 1, IBM, 10, 100, B, 1 --> `Identifier: 1-1`, `PriceUserIdentifier: 1-100`
N, 1, IBM, 10, 100, B, 2 --> `Identifier: 1-2`, `PriceUserIdentifier: 1-100`

and the map which contains quantity looks like this: `map[string]int{"1-100": 200}` 
so I know that on the TOB we have a volume of 200

## Order structures (order.go && cancel_order.go)

I decided to do one structure per type of order (here only two NewOrModifyOrder and CancelOrder).
I could use only one structure but it's cleaner to separate things as, for example, we can imagine that we want to add
some specific logic (log, ...) in a order type but not on the other.

I defined two kinds of identifier for an order:

- `Identifier`: it is unique --> `UserID-UserOrderID`
- `PriceUserIdentifier`: Not unique, it is used to retrieve the total volume of all orders with this identifier
   It is useful to retrieve volume changes in TOB.

## OrderBook (order_book.go)

The order book is a structure which contains two priority queues (ask/bid) and a boolean to indicate if it can trade
when an order crosses the book (if not, it creates a reject order).

The order book has a bunch of instructions as argument and returns the output as a string.

In order to ease tests and make it cleaner, my orderbook stops after a `F (flush) order, so I can 
process scenario per scenario and make tests well separated, (and it is simpler to add a new test if needed).

If we really need to process all the file in one run, it is very simple:
Instead of stopping after a `flush`, I could simply print outputs and then continue:
In `ProcessFromStringInstructions` (order_book.go), instead of breaking when encounter `F`,
print the output and clean the orderbook.
We didn't choose this way because I prefer to be **test-centered**, and running one scenario per scenario is cleaner
for tests.

When order book can trade, I chose (As I don't really know if I needed to) to make partial order possible 
(like in a low-volume market).

For example:

N, 1, IBM, 11, 100, B, 1
N, 1, IBM, 10, 100, B, 2 
N, 2, IBM, 10, 150, S, 100

will trade the first Bid order completely and only the half of the seconds


# To Improve

I really wanted to do the test on real condition so, there are some points to improve:

- Some logic could be simplify: for example when I try to make some generic function on order_book.go for
  Trades. Indeed, having an ask or bid queue, to know if a given order cross the book, we need 
  to have a different comparison function (a Bid order cross a book if Bid.Price > minAsk but for an ask order
  it is Ask.Price < maxBid). I could either split functions, create a function pointer on the `OrderQueue `struct
  or maybe use generics

- a better `main`! As I prefer to use test to compare input vs output

- I have a coverage of 86% it can be improved. For example adding tests when the instruction hasn't the right number
  of parameters on the order book

- Check performance of the heap structure (that I did) VS for example a Price list with 2 sorted list one ask
  and one bid (it should be larger in memory, but in time ?)

- Take more time to really check all the code and optimize it, but as it was a test with limited time, I follow rules

- More scenarios

- In order_book.go instead using a `[]string` use a `string.Buffer`, or maybe a `chan string` with a goroutine that
  will write the output.

- Other way to feed the orderbook instead a simple string

- Better documentation

- Better Dockerfile :
  --> Create a mount volume in order to give access to the exec file;
  --> Use `BUILD_KIT` to dynamically choose the right plateform and set `GOOS` and `GOARCH`
