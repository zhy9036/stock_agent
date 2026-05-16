package tools

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var portfolio = map[string]int{}
var cash float64 = 100000

func GetStockPrice(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input PriceInput,
) (*mcp.CallToolResult, any, error) {
	fmt.Println("get_stock_price tool called")
	price := fakePrice(input.Symbol)

	return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf(
						"%s price = %.2f",
						input.Symbol,
						price,
					),
				},
			},
		}, map[string]any{
			"symbol": input.Symbol,
			"price":  price,
		}, nil
}

func BuyStock(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input TradeInput,
) (*mcp.CallToolResult, any, error) {

	fmt.Println("buy_stock tool called")
	price := fakePrice(input.Symbol)

	total := price * float64(input.Quantity)

	if total > cash {
		return nil, nil, fmt.Errorf("not enough cash")
	}

	cash -= total
	portfolio[input.Symbol] += input.Quantity

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"Bought %d shares of %s",
					input.Quantity,
					input.Symbol,
				),
			},
		},
	}, nil, nil
}

func SellStock(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input TradeInput,
) (*mcp.CallToolResult, any, error) {

	fmt.Println("sell_stock tool called")
	owned := portfolio[input.Symbol]

	if owned < input.Quantity {
		return nil, nil, fmt.Errorf("not enough shares")
	}

	price := fakePrice(input.Symbol)

	total := price * float64(input.Quantity)

	cash += total
	portfolio[input.Symbol] -= input.Quantity

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"Sold %d shares of %s",
					input.Quantity,
					input.Symbol,
				),
			},
		},
	}, nil, nil
}

func GetPortfolio(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input struct{},
) (*mcp.CallToolResult, any, error) {

	fmt.Println("get_portfolio tool called")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf(
					"Cash: %.2f Portfolio: %+v",
					cash,
					portfolio,
				),
			},
		},
	}, nil, nil
}

func fakePrice(symbol string) float64 {
	base := map[string]float64{
		"AAPL": 190,
		"NVDA": 1200,
		"TSLA": 175,
	}

	price, ok := base[symbol]
	if !ok {
		price = 100
	}

	return price + rand.Float64()*10
}
