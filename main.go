package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true

	last := time.Now()
	first := true

	msgCh := make(chan message)
	procCh, repCh := checkMsg(msgCh, "")

	currentLevel := levels[0]

	g.SetManagerFunc(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView("input", 0, 0, maxX-1, 2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "INPUT"
			v.Editable = true

			v.Editor = gocui.EditorFunc(
				func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
					submit := false
					correction := false
					t := time.Since(last)
					if first {
						first = false
						t = time.Duration(0)
					}
					last = time.Now()

					switch {
					case ch != 0 && mod == 0:
						v.EditWrite(ch)
					case key == gocui.KeySpace:
						submit = true
					case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
						v.EditDelete(true)
						correction = true
					case key == gocui.KeyDelete:
						v.EditDelete(false)
						correction = true
					case key == gocui.KeyInsert:
						v.Overwrite = !v.Overwrite
					case key == gocui.KeyArrowLeft:
						v.MoveCursor(-1, 0, false)
						correction = true
					case key == gocui.KeyArrowRight:
						v.MoveCursor(1, 0, false)
						correction = true
					case key == gocui.KeyEnter:
						submit = true
					}
					select {
					case msgCh <- message{
						s:          strings.TrimSuffix(v.Buffer(), "\n"),
						submit:     submit,
						correction: correction,
						t:          t,
					}:
					default:
					}
					if submit {
						v.Clear()
						v.SetCursor(0, 0)
					}
				})

			if _, err := g.SetCurrentView("input"); err != nil {
				return err
			}
		}

		if v, err := g.SetView("text", 0, 3, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "TEXT"
			v.Wrap = true
		}

		if v, err := g.SetView("score", maxX-31, maxY-3, maxX-1, maxY-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "WPM ──ACC ──ERR "
			v.FgColor = gocui.ColorGreen

		}

		helpStr := " Ctrl+C: exit, Ctrl+R: restart"

		if v, err := g.SetView("help", 1, maxY-2, 1+2+len(helpStr), maxY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Frame = false
			fmt.Fprintf(v, helpStr)
		}

		return nil
	})

	toSave := false

	workers := func() {
		go func() {
			for m := range procCh {
				g.Update(func(g *gocui.Gui) error {
					if v, err := g.View("text"); err != nil {
						if err != gocui.ErrUnknownView {
							return err
						}
					} else {
						v.Clear()
						fmt.Fprintf(v, "%s", m.s)
					}
					return nil
				})
			}
		}()

		go func() {
			for r := range repCh {
				g.Update(func(g *gocui.Gui) error {
					if v, err := g.View("score"); err != nil {
						if err != gocui.ErrUnknownView {
							return err
						}
					} else {
						v.Clear()
						wpm := 0.0
						if r.duration != time.Duration(0) {
							wpm = 1 / r.duration.Minutes() / 5
						}
						fmt.Fprintf(v, " %3.0f WPM | %6.2f%% | %6.2f%% ",
							wpm,
							(100 - 100*float64(r.mistakes)/float64(r.lenght)),
							(100 * float64(r.wrongs) / float64(r.total)))
					}
					return nil
				})
				if r.end {
					maxX, maxY := g.Size()
					if v, err := g.SetView("end", maxX/2-10, maxY/2-3, maxX/2+10, maxY/2+3); err != nil {
						if err != gocui.ErrUnknownView {
							log.Fatal(err)
						}
						v.Title = "END"
						v.FgColor = gocui.ColorGreen
						g.Cursor = false
						v.Clear()

						wpm := 0.0
						if r.duration != time.Duration(0) {
							wpm = 1 / r.duration.Minutes() / 5
						}
						fmt.Fprintf(v, " Speed:    %3.0f WPM\n", wpm)
						m := float64(r.wrongs) / float64(r.total)
						fmt.Fprintf(v, " Mistakes: %6.2f%%\n", (100 * m))
						acc := float64(r.mistakes) / float64(r.lenght)
						fmt.Fprintf(v, " Accuracy: %6.2f%%\n", (100 - 100*acc))
						score := wpm * (1 - m) * (1 - acc)
						fmt.Fprintf(v, " -----------------\n")
						fmt.Fprintf(v, " Score:    %s", checkRank(score))

						if toSave {
							if err := checkScore(currentLevel, score); err != nil {
								log.Fatal(err)
							}
						}

						if _, err := g.SetCurrentView("end"); err != nil {
							log.Fatal(err)
						}
					}
				}
			}

		}()
	}

	g.SelFgColor = gocui.ColorYellow

	g.Update(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("menu", 2, 2, maxX-2, maxY-5); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "SELECT LEVEL"
			v.Editable = true
			v.Highlight = true
			v.SelFgColor = gocui.ColorBlack
			v.SelBgColor = gocui.ColorGreen
			g.Cursor = false

			v.Editor = gocui.EditorFunc(func(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
				switch {
				case key == gocui.KeyArrowDown:
					v.MoveCursor(0, 1, false)
				case key == gocui.KeyArrowUp:
					v.MoveCursor(0, -1, false)
				case ch == 'j':
					v.MoveCursor(0, 1, false)
				case ch == 'k':
					v.MoveCursor(0, -1, false)
				}
			})

			for _, l := range levels {
				fmt.Fprint(v, l.String())
				rank := getRank(l)
				fmt.Fprintf(v, " %s\n", rank)
			}

			g.SetCurrentView("menu")
		}
		return nil
	})

	workers()

	from := 0
	to := 0

	done := make(chan struct{})
	var wg sync.WaitGroup

	repeat := func(g *gocui.Gui, v *gocui.View) error {
		toSave = false
		first = true
		close(done)
		wg.Wait()
		done = make(chan struct{})
		msgCh = make(chan message)
		procCh, repCh = checkMsg(msgCh, freqEN(from, to))
		msgCh <- message{}
		workers()
		toSave = true

		wg.Add(1)
		go func() {
			defer wg.Done()
			if currentLevel.time == time.Duration(0) {
				return
			}

			t := time.NewTimer(currentLevel.time)
			select {
			case <-t.C:
			case <-done:
			}
			close(msgCh)
		}()

		if err := g.DeleteView("end"); err != nil && err != gocui.ErrUnknownView {
			return err
		}
		g.Cursor = true

		v, err := g.SetCurrentView("input")
		if err != nil {
			return err
		}
		v.Clear()
		err = v.SetCursor(0, 0)
		if err != nil {
			return err
		}

		return nil
	}

	startLevel := func(g *gocui.Gui, v *gocui.View) error {
		_, oy := v.Origin()
		_, cy := v.Cursor()

		currentLevel = levels[oy+cy]
		from = currentLevel.from
		to = currentLevel.to

		g.DeleteView("menu")

		if err := g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, repeat); err != nil {
			return nil
		}

		repeat(g, nil)
		return nil
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Fatal(err)
	}

	if err := g.SetKeybinding("menu", gocui.KeyEnter, gocui.ModNone, startLevel); err != nil {
		log.Fatal(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

type report struct {
	wrongs   int
	total    int
	mistakes int
	lenght   int
	duration time.Duration
	end      bool
}

type message struct {
	s          string
	t          time.Duration
	correction bool
	submit     bool
}

func mistake(s string) string {
	return fmt.Sprintf("\033[4%dm%s\033[0m", 1, s) // red highlight
}

func underline(s string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", 4, s) // underline
}

func fade(s string) string {
	return fmt.Sprintf("\033[3%dm%s\033[0m", 6, s) // gray
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
