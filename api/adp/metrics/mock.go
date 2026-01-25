package metrics

type mock struct{}

var _ Metrics = &mock{}

func NewMock() *mock {
	return &mock{}
}

func (m *mock) SetSessions(count int) {}
