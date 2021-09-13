package main

import (
	"errors"
	"log"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type level struct {
	menu string
	id   string
	from int
	to   int
	time time.Duration
}

// String implements Stringer.
func (l *level) String() string {
	return l.menu
}

var levels = []level{
	level{
		menu: "Time:    30sec",
		id:   "t30s",
		from: 0,
		to:   999,
		time: 30 * time.Second,
	},
	level{
		menu: "Time:     1min",
		id:   "t1m",
		from: 0,
		to:   999,
		time: 1 * time.Minute,
	},
	level{
		menu: "Level 1:  1~10",
		id:   "level1",
		from: 0,
		to:   9,
	},
	level{
		menu: "Level 2: 11~20",
		id:   "level2",
		from: 10,
		to:   19,
	},
	level{
		menu: "Level 3: 21~30",
		id:   "level3",
		from: 20,
		to:   29,
	},
	level{
		menu: "Level 4: 31~40",
		id:   "level4",
		from: 30,
		to:   39,
	},
	level{
		menu: "Level 5: 41~50",
		id:   "level5",
		from: 40,
		to:   49,
	},
	level{
		menu: "Test 1:   1~50",
		id:   "test1",
		from: 0,
		to:   49,
	},
}

type ranks struct {
	Scores map[string]float64 `yaml:"scores"`
}

func checkScore(l level, score float64) error {
	var saved ranks

	if d, err := os.ReadFile("scores.yml"); err == nil {
		err = yaml.Unmarshal(d, &saved)
		if err != nil {
			return err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if saved.Scores == nil {
		saved.Scores = map[string]float64{}
	}

	if s, ok := saved.Scores[l.id]; ok {
		if score > s {
			saved.Scores[l.id] = score
		}
	} else {
		saved.Scores[l.id] = score
	}

	d, err := yaml.Marshal(saved)
	if err != nil {
		return err
	}
	err = os.WriteFile("scores.yml", d, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getRank(l level) string {
	var saved ranks

	if d, err := os.ReadFile("scores.yml"); err == nil {
		err = yaml.Unmarshal(d, &saved)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
		saved.Scores = map[string]float64{}
	}

	s := saved.Scores[l.id]

	return checkRank(s)
}

func checkRank(score float64) string {
	ranking := []string{
		"-----",
		"*----",
		"**---",
		"***--",
		"****-",
		"*****",
		"*****+",
		"*****++",
	}

	adj := int((score - 50) / 10)
	if adj < 0 {
		adj = 0
	} else if adj >= len(ranking) {
		adj = len(ranking) - 1
	}

	return ranking[adj]
}
