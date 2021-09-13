package main

import (
	"encoding/csv"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func wordScram() string {
	w := "the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all the she he all"

	tokens := strings.Fields(w)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tokens), func(i, j int) {
		tokens[i], tokens[j] = tokens[j], tokens[i]
	})

	scram := strings.Join(tokens, " ")

	return scram
}

func freqEN(from, to int) string {
	f, err := os.Open("./4000commonEN.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	sel := records[from:to]
	tokens := []string{}
	repeat := 3
	for _, w := range sel {
		for i := 0; i < repeat; i++ {
			tokens = append(tokens, w...)
		}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tokens), func(i, j int) {
		tokens[i], tokens[j] = tokens[j], tokens[i]
	})

	return strings.Join(tokens, " ")
}
