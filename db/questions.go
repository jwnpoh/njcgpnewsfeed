package db

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Question struct {
    Year int
    Number int
    Wording string
}

func (q *Question) Contains(s, substr string) bool {
    return strings.Contains(s, substr)
}

type QuestionsDB struct {
    QnNumberAndQn map[int]string
    ListOfQuestions map[int]map[int]string
}

func NewQnDB() *QuestionsDB {
    var qnDB QuestionsDB
    qnDB.QnNumberAndQn = make(map[int]string)
    qnDB.ListOfQuestions = make(map[int]map[int]string)

    return &qnDB
}

// MapQuestions maps a list of questions in a file named by filename and maps them to a qnDB *PastYearQuestions.
func MapQuestions(qnDB *QuestionsDB) error {
    filename := "db/files/pastyressayqns.txt"
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("unable to open file %s - %w", filename, err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        mapqn(scanner.Text(), qnDB)
    }
    if err := scanner.Err(); err != nil {
        return fmt.Errorf("problem scanning lines in file and mapping questions")
    }
    return nil
}

func mapqn(s string, qnDB *QuestionsDB) {
    xs := strings.SplitN(s, " ", 3)
    i, _ := strconv.Atoi(xs[1])
    qnDB.QnNumberAndQn[i] = (xs[2])

    j, _ := strconv.Atoi(xs[0])
    qnDB.ListOfQuestions[j] = qnDB.QnNumberAndQn
}
