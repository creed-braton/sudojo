package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics interface {
	SetSessions(count int)
}

type metrics struct {
	sessions prometheus.Gauge
}

var _ Metrics = &metrics{}

func New(reg prometheus.Registerer) *metrics {
	m := &metrics{
		sessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "active_sessions_total",
			Help: "Number of currently active game sessions",
		}),
	}
	reg.MustRegister(m.sessions)
	return m
}

func (m *metrics) SetSessions(count int) {
	m.sessions.Set(float64(count))
}
