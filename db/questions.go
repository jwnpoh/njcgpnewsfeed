// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Question struct {
	Year    string
	Number  string
	Wording string
}

type QuestionsDB map[string]Question

// MapQuestions maps a list of questions in a file named by filename and maps them to a questions database.
func InitQuestionsDB() (QuestionsDB, error) {
	qnDB := make(map[string]Question)

	filename := "db/files/pastyressayqns.txt"
	file, err := os.Open(filename)
	if err != nil {
		return qnDB, fmt.Errorf("unable to open file %s - %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		s := scanner.Text()
		xs := strings.SplitN(s, " ", 3)

		year := xs[0]
		number := xs[1]
		wording := xs[2]

		qn := Question{year, number, wording}

		key := year + " " + number
		qnDB[key] = qn
	}

	if err := scanner.Err(); err != nil {
		return qnDB, fmt.Errorf("problem scanning lines in file and mapping questions")
	}
	return qnDB, nil
}
