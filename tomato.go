package tomato

import "time"

var (
	shortRest    = 5 * time.Minute
	longRest     = 15 * time.Minute
	focus        = 25 * time.Minute
	maxIteration = 4
)

type Tomato struct {
	focus  time.Duration
	end    time.Time
	timer  *time.Timer
	cancel chan (struct{})
}

func NewTomato(focus time.Duration) *Tomato {
	return &Tomato{focus: focus, cancel: make(chan struct{})}
}

func (t *Tomato) Start(f func()) {
	t.timer = time.AfterFunc(t.focus, f)
	t.end = time.Now()
}

func (t *Tomato) Done() time.Duration {
	if t.Stop() {
		return 0 * time.Nanosecond
	}

	return t.end.Sub(time.Now())
}

func (t *Tomato) Stop() bool {
	stopped := t.timer.Stop()
	if !stopped {
		<-t.timer.C
	}

	return stopped
}

func (t *Tomato) Reset() {
	t.Stop()
	t.timer.Reset(t.focus)
}

type Rest int

const (
	Short Rest = iota
	Long
)

type Notepad struct {
	iterations int
}

func NewNotepad() *Notepad {
	return &Notepad{}
}

func (n *Notepad) Check() Rest {
	n.iterations += 1

	if n.iterations%maxIteration == 0 {
		return Long
	} else {
		return Short
	}
}
