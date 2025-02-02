package metrics

import (
	prometheusclient "github.com/prometheus/client_golang/prometheus"
)

var (
	OnlinePlayers  prometheusclient.Counter
	OfflinePlayers prometheusclient.Counter
	TotalPlayers   *prometheusclient.GaugeVec
)

func RegisterOn(registerer prometheusclient.Registerer) {
	TotalPlayers = prometheusclient.NewGaugeVec(prometheusclient.GaugeOpts{
		Name: "matchmaking_total",
		Help: "Total number of players in the matchmaking service.",
	}, []string{"type"})
	OnlinePlayers = prometheusclient.NewCounter(prometheusclient.CounterOpts{
		Name: "matchmaking_online",
		Help: "Total number of online players in the matchmaking service.",
	})
	OfflinePlayers = prometheusclient.NewCounter(prometheusclient.CounterOpts{
		Name: "matchmaking_offline",
		Help: "Total number of offline players in the matchmaking service.",
	})

	registerer.MustRegister(
		TotalPlayers,
		OnlinePlayers,
		OfflinePlayers,
	)
}

func UnRegisterFrom(registerer prometheusclient.Registerer) {
	registerer.Unregister(TotalPlayers)
}
