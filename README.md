# Matchmaking Service

This service is responsible for matchmaking players into games. It is a simple service that takes a list of players and matches them into games based on their skill level.

[![Hub](https://badgen.net/docker/pulls/sealbro/matchmaking?icon=docker&label=matchmaking)](https://hub.docker.com/r/sealbro/matchmaking/)

## How to run

### Docker

```bash
docker compose up --build
```

### Local

- Run the matchmaking service
```bash
go run cmd/service/main.go
```

- Run the workload generator
```bash
go run cmd/workload/main.go
```

### Environment variables

[.env](./example.env)
| Name                          | Description                    | Default  |
|-------------------------------|--------------------------------|----------|
| `PRIVATE_ADDRESS`             | Private metrics address        | `:8081`  |
| `GRPC_PROTOCOL`               | gRPC protocol                  | `tcp`    |
| `GRPC_ADDRESS`                | gRPC address                   | `:32023` |
| `LOG_LEVEL`                   | slog level                     | `DEBUG`  |
| `QUEUE_SIZE`                  | Size of the matchmaking queue  | `25`     |
| `MIN_GROUP_SIZE`              | Minimum group size             | `10`     |
| `MAX_LEVEL_DIFF`              | Maximum level difference       | `10`     |
| `FIND_GROUP_EVERY_SECONDS`    | Find group every seconds       | `1`      |
| `MATCH_TIMEOUT_AFTER_SECONDS` | Matchmaking timeout in seconds | `60`     |


### Metrics

[localhost:8081/metrics](http://localhost:8081/metrics)

```
# HELP matchmaking_offline Total number of offline players in the matchmaking service.
# TYPE matchmaking_offline counter
matchmaking_offline 4034
# HELP matchmaking_online Total number of online players in the matchmaking service.
# TYPE matchmaking_online counter
matchmaking_online 4200
# HELP matchmaking_total Total number of players in the matchmaking service.
# TYPE matchmaking_total gauge
matchmaking_total{type="added"} 4200
matchmaking_total{type="matched"} 4140
matchmaking_total{type="removed"} 23
matchmaking_total{type="timeout"} 37
```


## Features

- [X] Add players into matchmaking queue
- [X] Delete players from matchmaking queue
- [X] Return match ID and the list of players in the match
- [X] Setup timeout for the matchmaking process
- [X] Configure matchmaking group size
- [X] Get metrics for the matchmaking process
- [X] Make simple API for the service

## What is not covered

- [ ] Security (authentication, authorization)
- [ ] Load balancing and high availability
- [ ] Monitoring and Tracing
- [ ] Permanent storage and restore after service restart
- [ ] How to process next matchmaking for players which can be in session now?

## Links

- <https://en.wikipedia.org/wiki/Elo_rating_system>
- <https://www.geeksforgeeks.org/elo-rating-algorithm/>
