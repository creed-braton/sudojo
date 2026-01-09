package message

import "testing"

func TestMessage(t *testing.T) {
	sender := "player-123"
	row, col, val := 4, 7, 9

	msg := New(InsertMsg, &row, &col, &val).SetSender(sender)

	if msg == nil {
		t.Fatal("expected message, got nil")
	}
	if msg.Type() != InsertMsg {
		t.Errorf("expected type '%s', got '%s'", InsertMsg, msg.Type())
	}
	if msg.Sender() != sender {
		t.Errorf("expected sender '%s', got '%s'", sender, msg.Sender())
	}
	if msg.Row() == nil {
		t.Fatal("expected row, got nil")
	}
	if *msg.Row() != row {
		t.Errorf("expected row %d, got %d", row, *msg.Row())
	}
	if msg.Column() == nil {
		t.Fatal("expected column, got nil")
	}
	if *msg.Column() != col {
		t.Errorf("expected column %d, got %d", col, *msg.Column())
	}
	if msg.Value() == nil {
		t.Fatal("expected value, got nil")
	}
	if *msg.Value() != val {
		t.Errorf("expected value %d, got %d", val, *msg.Value())
	}
}

func TestValidate(t *testing.T) {
	pointer := func(v int) *int { return &v }

	var tests = []struct {
		name string
		msg  Message
		want bool
	}{
		{
			name: "valid insert message",
			msg:  New(InsertMsg, pointer(0), pointer(0), pointer(1)).SetSender("player"),
			want: false,
		},
		{
			name: "valid ping message",
			msg:  New(PingMsg, pointer(0), pointer(0), nil).SetSender("player"),
			want: false,
		},
		{
			name: "valid state message",
			msg:  New(StateMsg, nil, nil, nil).SetSender("player"),
			want: false,
		},
		{
			name: "missing sender",
			msg:  New(InsertMsg, pointer(0), pointer(0), pointer(1)),
			want: true,
		},
		{
			name: "insert missing row",
			msg:  New(InsertMsg, nil, pointer(0), pointer(1)).SetSender("player"),
			want: true,
		},
		{
			name: "insert missing column",
			msg:  New(InsertMsg, pointer(0), nil, pointer(1)).SetSender("player"),
			want: true,
		},
		{
			name: "insert missing value",
			msg:  New(InsertMsg, pointer(0), pointer(0), nil).SetSender("player"),
			want: true,
		},
		{
			name: "ping missing row",
			msg:  New(PingMsg, nil, pointer(0), nil).SetSender("player"),
			want: true,
		},
		{
			name: "ping missing column",
			msg:  New(PingMsg, pointer(0), nil, nil).SetSender("player"),
			want: true,
		},
		{
			name: "ping with forbidden value",
			msg:  New(PingMsg, pointer(0), pointer(0), pointer(1)).SetSender("player"),
			want: true,
		},
		{
			name: "state with forbidden row",
			msg:  New(StateMsg, pointer(0), nil, nil).SetSender("player"),
			want: true,
		},
		{
			name: "state with forbidden column",
			msg:  New(StateMsg, nil, pointer(0), nil).SetSender("player"),
			want: true,
		},
		{
			name: "state with forbidden value",
			msg:  New(StateMsg, nil, nil, pointer(1)).SetSender("player"),
			want: true,
		},
		{
			name: "unknown message type",
			msg:  New("unknown", nil, nil, nil).SetSender("player"),
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.msg.Validate()
			if test.want && err == nil {
				t.Error("expected error, got nil")
			}
			if !test.want && err != nil {
				t.Errorf("expected no error, got: '%v'", err)
			}
		})
	}
}
