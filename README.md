# go-exchange
This project, built with Go, is a high-performance, highly available, full-featured trading system. At its core is a in-memory matching engine capable of processing massive volumes of trading orders with extremely low latency and high throughput.

Beyond the matching engine, we've developed a complete trading ecosystem:

- High-Performance Matching Engine: Using efficient data structures and a concurrency model, all order processing is completed in memory, ensuring millisecond-level response times and providing a solid foundation for high-frequency trading.
- Real-Time Market Data Publishing Service: Processed transaction data is pushed to users in real-time via an efficient message queue, ensuring timely and accurate information.
- Quotation System: After an order is matched, the system automatically executes the settlement logic, ensuring all changes to funds and assets are accurate and providing reliable transaction settlement.

This project leverages Go's native advantages in concurrent processing to ensure the entire system is highly stable and scalable. It is an end-to-end solution, from order entry to final settlement.

# System Architecture
From a logical perspective, the entire system can be divided into the following modules:
```
                               query
                   ┌───────────────────────────┐
                   │                           │
                   │                           ▼
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ Client  │─▶│   API   │──▶│Sequencer│──▶│ Engine │
└─────────┘   └─────────┘   └─────────┘   └─────────┘
                   ▲                           │
                   │                           │
┌─────────┐   ┌─────────┐                      │
│ Browser │──▶│   UI    │                      │
└─────────┘   └─────────┘                      │
    ▲                                         ▼
    │         ┌─────────┐   ┌─────────┐   ┌─────────┐
    └──────── │WebSocket│◀──│  Push   │◀─│Quotation│
              └─────────┘   └─────────┘   └─────────┘
```
- Trading API: The API gateway for traders to place and cancel orders.
- Sequencer: Responsible for sequencing all incoming orders.
- Trading Engine: Matches and settles sequenced orders.
- Quotation: Gathers trade information output from the matching process and generates market data like K-line charts.
- Push: Pushes market data, trade results, asset changes, and other information to users via channels like WebSockets.
- UI: Provides a web-based interface for traders, forwarding their actions to the backend API.

The Trading Engine is the most critical module, so its design requires careful consideration to ensure it is simple, reliable, and highly modular. Internally, the trading engine can be broken down into the following sub-modules:
```
  ┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐
     ┌─────────┐    ┌─────────┐
───▶│  Order  │───▶│  Match  │ │
     └─────────┘    └─────────┘
  │       │              │      │
          │              │
  │       ▼              ▼      │
     ┌─────────┐    ┌─────────┐
  │  │  Asset  │◀───│Clearing │ │
     └─────────┘    └─────────┘
  └ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘
```
- Asset Module: Manages user assets.
- Order Module: Manages users' active orders (orders that are not fully filled and not canceled).
- Matching Engine: Processes buy and sell orders and generates trade information.
- Clearing Module: Clears the trade information output from the matching engine, facilitating the exchange of assets between buyers and sellers.

The trading engine is an event-driven system at its core. Its inputs are a series of sequenced events, and its outputs are matching results, market data, and other information. The relationships between the internal modules of the trading engine are as follows: