package ui

type Sender string

const (
	Human Sender = "Human"
	AI    Sender = "AI"
)

type PagerMsg struct {
	Msg  string
	From Sender
}

func (m PagerMsg) String() string {
	return string(m.Msg)
}

func NewPagerMsg(msg string, from Sender) PagerMsg {
	return PagerMsg{Msg: msg, From: from}
}
