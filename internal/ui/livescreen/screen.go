package livescreen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// printedLineMsg is one line work printed, without its newline.
type printedLineMsg string

// elapsedTickMsg comes once a second so the screen redraws the elapsed time.
type elapsedTickMsg time.Time

// screen is the Bubble Tea model drawn in a terminal: every line work printed, then the elapsed time.
type screen struct {
	started  time.Time
	now      time.Time
	lines    []string
	stopWork context.CancelFunc // cancels work's context when Ctrl+C is pressed on the screen
}

// Run runs work with every line it prints drawn on a Bubble Tea screen, and returns work's error once both have ended.
// Ctrl+C on the screen cancels the ctx work got.
func Run(ctx context.Context, work func(ctx context.Context, stdout io.Writer) error) error {
	ctx, stopWork := context.WithCancel(ctx)
	defer stopWork()
	started := time.Now()
	program := tea.NewProgram(screen{started: started, now: started, stopWork: stopWork})
	workErr := make(chan error, 1)
	go func() {
		workErr <- work(ctx, &lineSender{program: program})
		program.Quit()
	}()
	_, screenErr := program.Run()
	return errors.Join(<-workErr, screenErr)
}

func (current screen) Init() tea.Cmd {
	return tickEverySecond()
}

func (current screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case printedLineMsg:
		current.lines = append(current.lines, string(msg))
	case elapsedTickMsg:
		current.now = time.Time(msg)
		return current, tickEverySecond()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			current.stopWork()
			return current, tea.Quit
		}
	}
	return current, nil
}

func (current screen) View() tea.View {
	elapsed := current.now.Sub(current.started)
	elapsedText := fmt.Sprintf("%02d:%02d:%02d", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
	return tea.NewView(strings.Join(append(current.lines, elapsedText), "\n"))
}

// tickEverySecond sends one elapsedTickMsg a second from now.
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(now time.Time) tea.Msg { return elapsedTickMsg(now) })
}

// lineSender is the stdout work writes to: it cuts what comes in at each newline and sends every whole line to program.
type lineSender struct {
	program *tea.Program
	pending []byte // the part after the last newline, waiting for the rest of its line
}

func (sender *lineSender) Write(text []byte) (int, error) {
	sender.pending = append(sender.pending, text...)
	for {
		newlineAt := bytes.IndexByte(sender.pending, '\n')
		if newlineAt < 0 {
			return len(text), nil
		}
		sender.program.Send(printedLineMsg(sender.pending[:newlineAt]))
		sender.pending = sender.pending[newlineAt+1:]
	}
}
