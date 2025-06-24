package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Spinner struct {
	message string
	frames  []string
	current int
	active  bool
	mu      sync.Mutex
	done    chan bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		done:    make(chan bool),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.done:
				return
			case <-ticker.C:
				s.mu.Lock()
				s.current = (s.current + 1) % len(s.frames)
				frame := s.frames[s.current]
				s.mu.Unlock()

				fmt.Printf("\r%s %s %s", color.CyanString(frame), s.message, strings.Repeat(" ", 20))
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	s.mu.Unlock()

	s.done <- true
	fmt.Print("\r" + strings.Repeat(" ", len(s.message)+30) + "\r")
}

func (s *Spinner) Success(message string) {
	s.Stop()
	fmt.Printf("%s %s\n", color.GreenString("✓"), message)
}

func (s *Spinner) Error(message string) {
	s.Stop()
	fmt.Printf("%s %s\n", color.RedString("✗"), message)
}

type ProgressBar struct {
	Total   int
	Current int
	Width   int
	Message string
}

func NewProgressBar(total int, message string) *ProgressBar {
	return &ProgressBar{
		Total:   total,
		Current: 0,
		Width:   40,
		Message: message,
	}
}

func (p *ProgressBar) Update(current int) {
	p.Current = current
	p.render()
}

func (p *ProgressBar) Increment() {
	p.Current++
	p.render()
}

func (p *ProgressBar) Complete() {
	p.Current = p.Total
	p.render()
	fmt.Println()
}

func (p *ProgressBar) render() {
	percent := float64(p.Current) / float64(p.Total)
	filled := int(percent * float64(p.Width))
	empty := p.Width - filled

	bar := color.GreenString(strings.Repeat("█", filled)) + strings.Repeat("░", empty)

	fmt.Printf("\r%s [%s] %3.0f%% (%d/%d)", p.Message, bar, percent*100, p.Current, p.Total)
}

func WithProgress(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := fn()

	if err != nil {
		spinner.Error(fmt.Sprintf("%s failed", message))
	} else {
		spinner.Success(fmt.Sprintf("%s completed", message))
	}

	return err
}
