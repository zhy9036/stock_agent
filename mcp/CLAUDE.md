# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o mcp-server ./main.go   # build
go run .                            # run (starts HTTP server on :8123)
```

No linter or test suite is configured.

## Architecture

This is a standalone Go MCP (Model Context Protocol) server that exposes stock trading tools over HTTP (`StreamableHTTPHandler`). It is part of the larger `stock_agent` monorepo — the sibling `agent/` directory contains a separate Go module for the LLM agent that consumes this MCP server.

### Server setup (`main.go`)
- Creates an `mcp.Server` with four registered tools: `get_stock_price`, `buy_stock`, `sell_stock`, `get_portfolio`.
- Each tool uses `StructSchema[T]()` to embed a JSON schema snippet into the tool description, generated reflectively from Go struct tags.
- Serves via `mcp.NewStreamableHTTPHandler` on port `:8123`. A new server instance is returned per request for isolation, but all instances share the same global state.

### Tools (`tools/`)
- **`tools.go`** — Tool handler implementations. Uses **global mutable state** (`var portfolio map[string]int`, `var cash float64`) for portfolio tracking. Prices are fake: seeded from a hardcoded map (`AAPL`, `NVDA`, `TSLA`) plus a random float. `BuyStock`/`SellStock` mutate cash and portfolio directly with no concurrency protection.
- **`types.go`** — Input structs (`PriceInput`, `TradeInput`) with `json` and `jsonschema` tags. `StructSchema[T]()` uses reflection to walk struct fields and build a human-readable schema string for tool descriptions.

### Dependencies
- `github.com/modelcontextprotocol/go-sdk` — the official Go SDK for building MCP servers.
