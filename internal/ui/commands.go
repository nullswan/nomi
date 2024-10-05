package ui

type Sender string

const (
	Human Sender = "Human"
	AI    Sender = "AI"
)

type PagerMsg struct {
	Msg  string
	From Sender
	stop bool
}

func (m PagerMsg) Stop() bool {
	return m.stop
}

func (m PagerMsg) String() string {
	return m.Msg
}

func NewPagerMsg(msg string, from Sender) PagerMsg {
	return PagerMsg{Msg: msg, From: from, stop: false}
}

func (m PagerMsg) WithStop() PagerMsg {
	m.stop = true
	return m
}
