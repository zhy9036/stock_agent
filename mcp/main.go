package main

import (
	"fmt"
	"log"
	"net/http"
	"stock_mcp/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newServer() *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "stock-mcp-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Tool: get_stock_price
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_stock_price",
			Description: fmt.Sprintf(
				"Get stock price for a given symbol.\n%s",
				tools.StructSchema[tools.PriceInput](),
			),
		},
		tools.GetStockPrice,
	)

	// Tool: buy_stock
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "buy_stock",
			Description: fmt.Sprintf(
				"Buy stock for a given symbol.\n%s",
				tools.StructSchema[tools.TradeInput](),
			),
		},
		tools.BuyStock,
	)

	// Tool: sell_stock
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "sell_stock",
			Description: fmt.Sprintf(
				"Sell stock for a given symbol.\n%s",
				tools.StructSchema[tools.TradeInput](),
			),
		},
		tools.SellStock,
	)

	// Tool: portfolio
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_portfolio",
			Description: "Get portfolio",
		},
		tools.GetPortfolio,
	)

	return server
}

func main() {

	log.Println("MCP server starting...")
	s := newServer()

	handler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			// IMPORTANT: new instance per request (safe for isolation)
			return s
		},
		&mcp.StreamableHTTPOptions{},
	)

	httpServer := &http.Server{
		Addr:    ":8123",
		Handler: handler,
	}

	log.Println("MCP HTTP server listening on :8123")

	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
