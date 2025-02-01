# Matchmaking Service

This service is responsible for matchmaking players into games. It is a simple service that takes a list of players and matches them into games based on their skill level.

## How to run

### Environment variables

| Name                          | Description                    | Default |
|-------------------------------|--------------------------------|---------|
| `PRIVATE_ADDRESS`             | Private metrics address        | `:8081` |
| `LOG_LEVEL`                   | slog level                     | `DEBUG` |
| `QUEUE_SIZE`                  | Size of the matchmaking queue  | `10`    |
| `MIN_GROUP_SIZE`              | Minimum group size             | `10`    |
| `MAX_LEVEL_DIFF`              | Maximum level difference       | `10`    |
| `FIND_GROUP_EVERY_SECONDS`    | Find group every seconds       | `1`     |
| `MATCH_TIMEOUT_AFTER_SECONDS` | Matchmaking timeout in seconds | `60`    |


### Metrics

```
# HELP matchmaking_total Total number of players in the matchmaking service.
# TYPE matchmaking_total gauge
matchmaking_total{type="matched"} 800
matchmaking_total{type="added"}   1000
matchmaking_total{type="timeout"} 80
matchmaking_total{type="removed"} 120
```


## Features

- [X] Add players into matchmaking queue
- [X] Delete players from matchmaking queue
- [X] Return match ID and the list of players in the match
- [X] Setup timeout for the matchmaking process
- [X] Configure matchmaking group size
- [ ] Get metrics for the matchmaking process
- [ ] Make simple API for the service

## What is didn't covered

- [ ] Security (authentication, authorization)
- [ ] Load balancing and high availability
- [ ] Monitoring and Tracing
- [ ] Permanent storage and restore after service restart
- [ ] Web Socket events for real-time updates
- [ ] How to process next matchmaking for players which can be in session now?

## Links

- <https://en.wikipedia.org/wiki/Elo_rating_system>
- <https://www.geeksforgeeks.org/elo-rating-algorithm/>
