package main

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nsf/termbox-go"
)

const (
	FailPenaltySeconds         = 3
	FailPenaltyDuration        = time.Second * FailPenaltySeconds
	FastErrorHighlightDuration = time.Millisecond * 333
	ScoreHighlightDuration     = time.Second * 3
)

const (
	black   = termbox.ColorBlack
	red     = termbox.ColorRed
	green   = termbox.ColorGreen
	yellow  = termbox.ColorYellow
	blue    = termbox.ColorBlue
	magenta = termbox.ColorMagenta
	cyan    = termbox.ColorCyan
	white   = termbox.ColorWhite

	bold = termbox.AttrBold
)

type Align int

const (
	Left Align = iota
	Center
	Right
)

type printSpec struct {
	text  string
	x     int
	y     int
	fg    termbox.Attribute
	bg    termbox.Attribute
	align Align
}

func render(s State, now time.Time) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	w, h := termbox.Size()

	byteOffset, runeOffset := errorOffset(s.Phrase.Text, s.Phrase.Input)

	if s.Phrase.ShowFail(now) {
		left := min(int(s.Phrase.CurrentRound().FailedAt.
			Add(FailPenaltyDuration).Sub(now).Seconds()+1), FailPenaltySeconds)
		write(text(failMessage(s.Phrase.CurrentRound().Errors), left).
			X(w / 2).Y(h / 2).Fg(red | bold).Align(Center))
	} else {
		x := (w / 2) - (utf8.RuneCountInString(s.Phrase.Text) / 2)
		write(text(s.Phrase.Text + string('⏎')).X(x).Y(h / 2).Fg(white))

		write(text(spaced(s.Phrase.Input[:byteOffset])).
			X(x).Y(h / 2).Fg(green))
		write(text(spaced(s.Phrase.Input[byteOffset:])).
			X(x + runeOffset).Y(h / 2).Fg(black).Bg(red))
	}

	if s.Repeat {
		write(text("Repeating phrase").X(w - 1).Y(1).Align(Right))
	}

	if now.Before(s.LastScoreUntil) {
		write(text("   Score: %.0f +%.0f (%.0f%%)", s.Score, s.LastScore,
			100.*s.LastScorePercent).X(1).Y(1).Fg(blue | bold))
	}
	write(text("   Score: %.0f", s.Score).X(1).Y(1))
	write(text("   Level: %d", level(s.Score)).X(1).Y(2))
	write(text("Progress: %.0f%%", 100*progress(s.Score)).X(1).Y(3))

	write(text("In %s mode", s.Phrase.Mode.Name()).
		X(w / 2).Y(h/2 - 4).Fg(s.Phrase.Mode.Attr()).Align(Center))
	write(text("(%s!)", s.Phrase.Mode.Desc()).X(w / 2).Y(h/2 - 3).Align(Center))

	seconds, _, _ := computeStats(
		s.Phrase.Input[:byteOffset], s.Phrase.CurrentRound().StartedAt, now)

	write(text("%3d errors", s.Phrase.CurrentRound().Errors).
		X(w/2 - 1).Y(h/2 + 4).Align(Right).Fg(s.Phrase.ErrorCountColor(now)))
	write(text("%4.1f seconds", seconds).X(w/2 + 1).Y(h/2 + 4))

	write(text("What's this fast, slow, medium thing?!").X(1).Y(h - 3))
	write(text("http://steve-yegge.blogspot.com/2008/09/programmings-dirtiest-little-secret.html").X(1).Y(h - 2))
}

func text(t string, args ...interface{}) *printSpec {
	s := &printSpec{}
	if len(args) > 0 {
		s.text = fmt.Sprintf(t, args...)
	} else {
		s.text = t
	}
	return s
}

func write(spec *printSpec) {
	var x int
	switch spec.align {
	case Left:
		x = spec.x
	case Center:
		x = spec.x - utf8.RuneCountInString(spec.text)/2
	case Right:
		x = spec.x - utf8.RuneCountInString(spec.text)
	}

	for _, c := range spec.text {
		termbox.SetCell(x, spec.y, c, spec.fg, spec.bg)
		x++
	}
}

func spaced(s string) string {
	return strings.Replace(s, " ", "␣", -1)
}

func (p *printSpec) Align(align Align) *printSpec {
	p.align = align
	return p
}

func (p *printSpec) X(x int) *printSpec {
	p.x = x
	return p
}

func (p *printSpec) Y(y int) *printSpec {
	p.y = y
	return p
}

func (p *printSpec) Fg(fg termbox.Attribute) *printSpec {
	p.fg = fg
	return p
}

func (p *printSpec) Bg(bg termbox.Attribute) *printSpec {
	p.bg = bg
	return p
}

func failMessage(errs int) string {
	switch errs {
	case 1:
		return "Not quite! Please try again in %d..."
	case 2, 3:
		return "FAIL! Let's do this again in %d..."
	case 4, 5:
		return "Dude?! Try again in %d..."
	case 6, 7, 8:
		return "Are you serious?!? Again in %d..."
	default:
		return "I don't even... %d..."
	}
}
