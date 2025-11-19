package player

type mockPool struct {
	size    func() int
	create  func(name string) (string, error)
	join    func(token string) ([]Player, error)
	leave   func(token string) ([]Player, error)
	players func() []Player
}

var _ Pool = &mockPool{}

func NewMockPool(
	size func() int,
	create func(name string) (string, error),
	join func(token string) ([]Player, error),
	leave func(token string) ([]Player, error),
	players func() []Player,
) *mockPool {
	return &mockPool{
		size:    size,
		create:  create,
		join:    join,
		leave:   leave,
		players: players,
	}
}

func (p *mockPool) Size() int {
	return p.size()
}

func (p *mockPool) Create(name string) (string, error) {
	return p.create(name)
}

func (p *mockPool) Join(token string) ([]Player, error) {
	return p.join(token)
}

func (p *mockPool) Leave(token string) ([]Player, error) {
	return p.leave(token)
}

func (p *mockPool) Players() []Player {
	return p.players()
}
