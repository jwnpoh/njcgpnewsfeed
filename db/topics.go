package db

import "sort"

// Topic represents a searchable tag for each article.
type Topic string

// TopicsMap is a map to track the number of articles tagged with each topic in the database.
type TopicsMap map[Topic]int

// TopicObject provides an object to implement sorting by value in order to rank most and least tagged topics.
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

// Increment increases the count for the question in the TopicsMap map.
func (tm TopicsMap) Increment(key Topic) {
	tm[key]++
}

// Decrement decreases the count for the question in the TopicsMap map.
func (tm TopicsMap) Decrement(key Topic) {
	tm[key]--
	if tm[key] < 1 {
		delete(tm, key)
	}
}

// InitTopicsMap returns an initialised TopicsMap.
func InitTopicsMap() TopicsMap { return make(map[Topic]int) }

// GetTopicsCount is a function to count topics on first initialisation of the Articles DB.
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

func GetTopics(tm TopicsMap) ([]string, error) {
	topics := make([]string, 0, len(tm))

	for k := range tm {
		topics = append(topics, string(k))
	}

	sort.Strings(topics)

	return topics, nil
}
