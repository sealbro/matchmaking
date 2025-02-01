package metrics

import (
	prometheusclient "github.com/prometheus/client_golang/prometheus"
)

var (
	TotalPlayers *prometheusclient.GaugeVec
)

func RegisterOn(registerer prometheusclient.Registerer) {
	TotalPlayers = prometheusclient.NewGaugeVec(prometheusclient.GaugeOpts{
		Name: "matchmaking_total",
		Help: "Total number of players in the matchmaking service.",
	}, []string{"type"})

	registerer.MustRegister(
		TotalPlayers,
	)
}

func UnRegisterFrom(registerer prometheusclient.Registerer) {
	registerer.Unregister(TotalPlayers)
}
