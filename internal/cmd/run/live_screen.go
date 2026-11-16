package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// printedLineMsg is one line Run printed, without its newline.
type printedLineMsg string

// elapsedTickMsg comes once a second so the screen redraws the elapsed time.
type elapsedTickMsg time.Time

// liveScreen is the Bubble Tea model nightcap run draws in a terminal: every line Run printed, then the elapsed time.
type liveScreen struct {
	started time.Time
	now     time.Time
	lines   []string
	stopRun context.CancelFunc // cancels Run's context when Ctrl+C is pressed on the screen
}

// RunOnLiveScreen runs Run with its lines drawn on a Bubble Tea screen, and returns Run's error once both have ended.
func RunOnLiveScreen(ctx context.Context, args []string, version string) error {
	ctx, stopRun := context.WithCancel(ctx)
	defer stopRun()
	started := time.Now()
	program := tea.NewProgram(liveScreen{started: started, now: started, stopRun: stopRun})
	runErr := make(chan error, 1)
	go func() {
		runErr <- Run(ctx, args, version, &lineSender{program: program})
		program.Quit()
	}()
	_, screenErr := program.Run()
	return errors.Join(<-runErr, screenErr)
}

func (screen liveScreen) Init() tea.Cmd {
	return tickEverySecond()
}

func (screen liveScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case printedLineMsg:
		screen.lines = append(screen.lines, string(msg))
	case elapsedTickMsg:
		screen.now = time.Time(msg)
		return screen, tickEverySecond()
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			screen.stopRun()
			return screen, tea.Quit
		}
	}
	return screen, nil
}

func (screen liveScreen) View() tea.View {
	elapsed := screen.now.Sub(screen.started)
	elapsedText := fmt.Sprintf("%02d:%02d:%02d", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60)
	return tea.NewView(strings.Join(append(screen.lines, elapsedText), "\n"))
}

// tickEverySecond sends one elapsedTickMsg a second from now.
func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(now time.Time) tea.Msg { return elapsedTickMsg(now) })
}

// lineSender is the stdout Run writes to on a live screen: it cuts what comes in at each newline and sends every whole line to program.
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
