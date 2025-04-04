# Event processing challenge

## Publisher

The publisher is designed as `Observable` pattern connected to the Redis which has three subscribers: `[GameSubscriber, PlayerSubscriber, TimeSubscriber]`.

The publisher reads the events from the channel and publishes them to the subscribers as `CASINO_EVENT`.

Each event is enriched with:
- `Common currency` from the API,
- `Player data` from the DB
- `Human-friendly description` dinamically from the event data.

## Subscribers

Connected to the `CASINO_EVENT` Redis channel, read the events and handle the data. Three different subscribers are implemented: `[GameSubscriber, PlayerSubscriber, TimeSubscriber]`

Each subscriber has `BaseSubscriber` that allows the same `Subscribe/Unsubscribe` behaviour (`Template` design pattern in the OOP world) and its own statistics data structure for storing the values required for the API endpoints (e.g `/materialized`)

- `GameSubscriber` - stores for each game:
    - `id` - game id,
    - `name` - game name,
    - `game_played_count` - how many times the games has been played,
    - `bet_per_currency` - how many bets per currency has been staked.

- `PlayerSubscriber` - stores for each player:
    - `bet_count` - how many times the player bet
    - `bet_amount` - how much the player has been bet
    - `deposit_count` - how many times the player deposit
    - `deposit_amount` - how much the player has deposited
    - `won_count` - how many time the player has won

    And calculate the `PlayerStats` for the `/materialized` API.

- `TimeSubscriber` - stores general time statistics for `/materialized` API.
    - `total_events`
    - `events_per_minute`
    - `moving_avg_per_second`

### Concurrency Features 

- `GameSubscriber` - doesn't handle concurrency, becuse `Publisher` publishes events sequentially and there is no need to worry about concurrent approach to the data
    
- `PlayerSubscriber` - uses `atomic` and `sync.Mutex` to handle potential concurrency

- `TimeSubscriber` - relies on the Redis data structures that are multi-thread safe


## Unsubscribtion

The subscribers read the events from the channel until receive a `STOP_SIGNAL` message. After this message, all subscribers unsubscribe.

## Graceful shutdown

API endpoints (Http listener) and Redis connection keep open until the `SIGTERM/SIGINT` signal. After the termination signal is received, all connections close.