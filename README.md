# 📈 Stock Agent

An AI-powered stock trading system that combines LLM reasoning with real trading tools via the Model Context Protocol (MCP).

## 🚀 Features

- 🧠 LLM-powered stock analysis and trading decisions
- 🛠️ Real trading tools: buy/sell stocks, check prices, view portfolio
- 🔗 MCP-based architecture for clean separation of concerns
- 💬 Interactive CLI interface for natural language trading
- 🔄 Tool calling with automatic retries and error handling

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐
│   LLM Agent     │    │   MCP Server     │
│                 │    │                  │
│  agent/main.go  │◄──►│   mcp/main.go    │
│  agent/llm.go   │    │   mcp/tools/     │
└─────────────────┘    └──────────────────┘
```

The system consists of two separate Go modules:

1. **Agent** (`agent/`) - The LLM client that processes user queries and calls tools
2. **MCP Server** (`mcp/`) - Exposes trading tools over HTTP via the Model Context Protocol

## 📦 Installation

Clone the repository:
```bash
git clone <repository-url>
cd stock_agent
```

Both modules have separate dependencies:

**Agent Module:**
```bash
cd agent
go mod tidy
```

**MCP Server Module:**
```bash
cd ../mcp
go mod tidy
```

## ▶️ Usage

### 1. Start the MCP Server

In one terminal, start the MCP server:
```bash
cd mcp
go run .
# Server will start on http://localhost:8123
```

### 2. Configure the Agent

The agent expects an OpenAI-compatible API endpoint. Update the `config.txt` file in the `agent/` directory:
```txt
base_url="http://localhost:7001/v1"
api_key="your-api-key-here"
model_id="your-model-id"
```

### 3. Run the Agent

In another terminal:
```bash
cd agent
go run .
```

### 4. Interact with the Agent

Example queries:
```
> buy 10 AAPL
> sell 5 NVDA
> should I buy TSLA?
> analyze my portfolio
> what's the price of AAPL?
> quit
```

## 🧰 Tools Provided

The MCP server exposes four trading tools:

1. **get_stock_price** - Get current price for a stock symbol
2. **buy_stock** - Purchase shares of a stock
3. **sell_stock** - Sell shares of a stock  
4. **get_portfolio** - View current cash balance and holdings

All tools maintain state in memory using global variables.

## 📁 Project Structure

```
stock_agent/
├── agent/              # LLM agent client (Go module)
│   ├── main.go         # Main entry point and CLI interface
│   ├── llm.go          # LLM interaction and tool calling
│   ├── types.go        # Data structures
│   ├── utils.go        # Utility functions
│   ├── go.mod          # Agent dependencies
│   └── config.txt      # API configuration
└── mcp/                # MCP server (Go module)
    ├── main.go         # Server entry point
    ├── tools/          # Trading tool implementations
    │   ├── tools.go    # Tool handler functions
    │   └── types.go    # Input type definitions
    ├── go.mod          # MCP server dependencies
    └── CLAUDE.md       # Development guidelines
```

## 🔧 Development

### Building

**Agent:**
```bash
cd agent
go build -o stock-agent .
```

**MCP Server:**
```bash
cd mcp
go build -o mcp-server .
```

### Running Tests

No test suite is currently configured.

## ⚠️ Limitations

- **Fake Pricing**: Stock prices are simulated with hardcoded base values plus random fluctuations
- **In-Memory State**: Portfolio data is stored in global variables and resets on server restart
- **No Authentication**: The MCP server has no authentication mechanism
- **Single User**: Global state means all clients share the same portfolio

## 📄 License

MIT License © 2026

## 🤝 Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss proposed modifications.