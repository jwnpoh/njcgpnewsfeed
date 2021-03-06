// Package db provides functions and types relevant to the backend database for the article feed.
package db

import (
	"regexp"
	"strings"
)

// Search implements a switch statement to check which kind of search to run.
func Search(term string, database *ArticlesDBByDate) *ArticlesDBByDate {
	// handle boolean search
	switch {
	case strings.Contains(term, "AND"):
		return SearchAND(term, database)
	case strings.Contains(term, "OR"):
		return SearchOR(term, database)
	case strings.Contains(term, "NOT"):
		return SearchNOT(term, database)
	default:
		return SearchAll(term, database)
	}
}

// SearchAll runs a search of the given term through all the items stored in the database.
func SearchAll(term string, database *ArticlesDBByDate) *ArticlesDBByDate {
	results := NewArticlesDBByDate()

	for _, i := range *database {
		if !searchTitle(term, i) && !searchTopics(term, i) && !searchQuestions(term, i) && !searchDate(term, i) {
			continue
		} else {
			*results = append(*results, i)
		}
	}
	return results
}

// SearchAND runs an AND boolean search.
func SearchAND(term string, database *ArticlesDBByDate) *ArticlesDBByDate {
	temp := NewArticlesDBByDate()
	results := NewArticlesDBByDate()
	terms := strings.Split(term, "AND")

	for i, t := range terms {
		t = strings.TrimSpace(t)
		temp = SearchAll(t, database)
		database = temp
		temp = NewArticlesDBByDate()
		if i == len(terms)-1 {
			results = database
		}
	}
	return results
}

// SearchOR runs an OR boolean search.
func SearchOR(term string, database *ArticlesDBByDate) *ArticlesDBByDate {
	results := NewArticlesDBByDate()
	terms := strings.Split(term, "OR")

	for _, t := range terms {
		t = strings.TrimSpace(t)
		for _, i := range *database {
			if !searchTitle(t, i) && !searchTopics(t, i) && !searchQuestions(t, i) && !searchDate(t, i) {
				continue
			} else {
				*results = append(*results, i)
			}
		}
	}
	return results
}

// SearchNOT runs a NOT boolean search.
func SearchNOT(term string, database *ArticlesDBByDate) *ArticlesDBByDate {
	temp := NewArticlesDBByDate()
	results := NewArticlesDBByDate()
	terms := strings.Split(term, "NOT")
	termsToExclude := terms[1:]

	for _, exclude := range termsToExclude {
		exclude = strings.TrimSpace(exclude)
		for _, i := range *database {
			if searchTitle(exclude, i) || searchTopics(exclude, i) || searchQuestions(exclude, i) || searchDate(exclude, i) {
				continue
			} else {
				*temp = append(*temp, i)
			}
		}
		database = temp
		temp = NewArticlesDBByDate()
	}

	for _, i := range *database {
		t := strings.TrimSpace(terms[0])
		if !searchTitle(t, i) && !searchTopics(t, i) && !searchQuestions(t, i) && !searchDate(t, i) {
			continue
		} else {
			*results = append(*results, i)
		}
	}

	return results
}

func searchTitle(term string, a Article) bool {
	rx := regexp.MustCompile(`\b` + strings.ToLower(term) + `\b`)
	return rx.MatchString(strings.ToLower(a.Title))
}

func searchTopics(term string, a Article) bool {
	rx := regexp.MustCompile(`\b` + strings.ToLower(term) + `\b`)
	for _, j := range a.Topics {
		return rx.MatchString(strings.ToLower(string(j)))
	}
	return false
}

func searchQuestions(term string, a Article) bool {
	searchYr := regexp.MustCompile(`^\d{4}$`)
	searchYrAndQn := regexp.MustCompile(`^\d{4}\s?-?\s?(q|Q)\d{1,2}$`)
	searchQnNo := regexp.MustCompile(`^(q|Q)\d{1,2}$`)

	switch {
	case searchYr.MatchString(term):
		for _, j := range a.Questions {
			return j.Year == term
		}
	case searchQnNo.MatchString(term):
		cutQnNo := regexp.MustCompile(`(q|Q)\d{1,2}`)
		qnNumber := strings.TrimLeft(strings.ToLower(cutQnNo.FindString(term)), "q")
		for _, j := range a.Questions {
			return j.Number == qnNumber
		}
	case searchYrAndQn.MatchString(term):
		cutQnNo := regexp.MustCompile(`(q|Q)\d{1,2}`)
		qnNumber := strings.TrimLeft(strings.ToLower(cutQnNo.FindString(term)), "q")
		cutYear := regexp.MustCompile(`\d{4}`)
		year := cutYear.FindString(term)
		for _, j := range a.Questions {
			if j.Number == qnNumber && j.Year == year {
				return true
			}
		}
	default:
		for _, j := range a.Questions {
			rx := regexp.MustCompile(`\b` + strings.ToLower(term) + `\b`)
			return rx.MatchString(strings.ToLower(j.Wording))
		}
	}
	return false
}

func searchDate(term string, a Article) bool {
	rx := regexp.MustCompile(`\b` + strings.ToLower(term) + `\b`)
	return rx.MatchString(strings.ToLower(a.DisplayDate))
}
