package db

import "sort"

// Topic represents a searchable tag for each article.
type Topic string

// TopicsMap is a map to track the number of articles tagged with each topic in the database.
type TopicsMap map[Topic]int

type TopicObject struct {
	Key   Topic
	Value int
}

// TopicsCount provides a type to sort the TopicsMap.
type TopicsCount []TopicObject

// Implement sort.Sort interface.
func (tc TopicsCount) Len() int           { return len(tc) }
func (tc TopicsCount) Swap(i, j int)      { tc[i], tc[j] = tc[j], tc[i] }
func (tc TopicsCount) Less(i, j int) bool { return tc[i].Value < tc[j].Value }

func (tm TopicsMap) Increment(topic string) {
	tm[Topic(topic)]++
}

func InitTopicsMap() TopicsMap { return make(map[Topic]int) }

func GetTopicsCount(tm TopicsMap) TopicsCount {
	tc := make(TopicsCount, len(tm))

	var i int
	for k, v := range tm {
		tc[i] = TopicObject{k, v}
		i++
	}

	sort.Sort(sort.Reverse(tc))

	return tc
}
