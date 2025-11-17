package player

type mockPool struct {
	size   func() int
	create func(name string) (string, error)
	join   func(token string) ([]Player, error)
	leave  func(token string) ([]Player, error)
}

var _ Pool = &mockPool{}

func NewMockPool(
	size func() int,
	create func(name string) (string, error),
	join func(token string) ([]Player, error),
	leave func(token string) ([]Player, error),
) *mockPool {
	return &mockPool{
		size:   size,
		create: create,
		join:   join,
		leave:  leave,
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
