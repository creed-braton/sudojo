package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	SetActiveLobbies(count int)
}

type metrics struct {
	activeLobbies prometheus.Gauge
}

var _ Metrics = &metrics{}

func New(reg prometheus.Registerer) *metrics {
	m := &metrics{
		activeLobbies: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "active_lobbies_total",
			Help: "Number of currently active lobbies",
		}),
	}
	reg.MustRegister(m.activeLobbies)
	return m
}

func (m *metrics) SetActiveLobbies(count int) {
	m.activeLobbies.Set(float64(count))
}
