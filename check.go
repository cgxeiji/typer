package main

import (
	"strings"
	"time"
)

func checkMsg(msgCh <-chan message, source string) (<-chan message, <-chan report) {
	sum := time.Duration(0)
	n := 0

	procCh := make(chan message)
	repCh := make(chan report)

	go func() {
		defer close(procCh)
		defer close(repCh)

		history := ""
		prev := []rune{}
		mistakes := 0
		strokes := 0
		wrongs := 0
		hasMistake := false
		text := ""
		duration := time.Duration(0)

		tokens := strings.Fields(source)
		total := len(tokens)

		for m := range msgCh {
			duration = time.Duration(0)
			if m.t != time.Duration(0) {
				sum += m.t
				n++
				duration = sum / time.Duration(n)
			}

			text = ""
			o := []rune(tokens[0])
			hasMistake = false

			if len(m.s) > len(o) {
				text += mistake(string(o))
				if len(m.s) > len(prev) {
					mistakes++
				}
				hasMistake = true
			} else {
				for i, c := range m.s {
					if i >= len(o) {
						break
					}
					if c == o[i] {
						text += fade(string(o[i]))
					} else {
						if i < len(prev) {
							if c != prev[i] {
								mistakes++
							}
						} else {
							mistakes++
						}
						text += mistake(string(o[i]))
						hasMistake = true
					}
				}
			}
			if len(m.s) < len(o) {
				if m.submit {
					text += mistake(string(o[len(m.s):]))
					hasMistake = true
					mistakes++
				} else {
					text += underline(string(o[len(m.s)]))
					text += string(o[len(m.s)+1:])
				}
			}

			if m.submit {
				if hasMistake {
					wrongs++
				}

				if len(tokens) == 1 {
					break
				}

				history += text + " "
				tokens = tokens[1:]
				next := tokens[0]
				text = underline(string(next[0]))
				if len(next) > 1 {
					text += string(next[1:])
				}
			}
			text += " " + strings.Join(tokens[1:], " ")

			if !m.correction {
				strokes++
			}

			procCh <- message{
				s: history + text,
			}
			repCh <- report{
				mistakes: mistakes,
				duration: duration,
				lenght:   strokes,
				wrongs:   wrongs,
				total:    total,
			}
			prev = []rune(m.s)
		}

		procCh <- message{
			s: history + text,
		}
		repCh <- report{
			mistakes: mistakes,
			duration: duration,
			lenght:   strokes,
			end:      true,
			wrongs:   wrongs,
			total:    total,
		}

	}()

	return procCh, repCh
}
