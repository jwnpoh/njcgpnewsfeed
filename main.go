package main

import (
	"fmt"
	"log"

	"github.com/jwnpoh/njcgpnewsfeed/cmd"
	"github.com/jwnpoh/njcgpnewsfeed/db"
)

func main() {
    database, err := db.InitArticlesDBByDate()
    if err != nil {
        log.Fatal(err)
    }

    qnDB := db.NewQnDB()
    if err = db.MapQuestions(qnDB); err != nil {
        log.Fatal(err)
    }


    a, err := db.NewArticle() 
    if err != nil {
        log.Fatal(err)
    }
    a.Title = "How to Train Your Dragon"
    a.URL = "http://www.httyd.com"
    a.SetTopics("Dragons", "Fantasy")
    a.SetQuestions(2020, 2, qnDB)
    a.SetQuestions(2020, 3, qnDB)
    a.SetDate("27 Jan 2020")

    a.AddArticleToDB(database)

    b, err := db.NewArticle() 
    if err != nil {
        log.Fatal(err)
    }
    b.Title = "How to Train Your Dodo"
    b.URL = "http://www.httyd.com"
    b.SetTopics("Dodo", "Fantasy")
    b.SetQuestions(2020, 5, qnDB)
    b.SetDate("25 Mar 2021")

    b.AddArticleToDB(database)

    c, err := db.NewArticle() 
    if err != nil {
        log.Fatal(err)
    }
    c.Title = "How to Train Your Dog"
    c.URL = "http://www.httyd.com"
    c.SetTopics("Dog", "Food")
    c.SetQuestions(2020, 6, qnDB)
    c.SetQuestions(2020, 7, qnDB)
    c.SetDate("26 Feb 2018")

    c.AddArticleToDB(database)

    d, err := db.NewArticle() 
    if err != nil {
        log.Fatal(err)
    }
    d.Title = "How to Train Your Duck"
    d.URL = "http://www.httyd.com"
    d.SetTopics("Duck", "Birds")
    d.SetQuestions(2005, 6, qnDB)
    d.SetQuestions(2010, 7, qnDB)
    d.SetDate("26 May 2020")

    d.AddArticleToDB(database)

    term := "FAR"
    results, err := cmd.Search(term, database)
    if err != nil {
        fmt.Println(err)
    } else {
        for _, j := range *results {
            fmt.Printf("%s\n%s\n%s\n%v\n%v\n\n", j.DisplayDate, j.Title, j.URL, j.Topics, j.Questions)
        }
    }


    // for _, j := range *database {
    //     for _, k := range j.Questions {
    //         if strings.Contains(strings.ToLower(k.Wording), "how") {
    //             fmt.Printf("%s\n%s\n%s\n%v\n%v\n\n", j.DisplayDate, j.Title, j.URL, j.Topics, j.Questions)
    //         }
    //     }
        // fmt.Println(j.DisplayDate)
        // fmt.Println(j.Title)
        // fmt.Println(j.URL)
        // fmt.Println(j.Topics)
        // fmt.Println(j.Questions)
        // fmt.Println()
    // }

    
}
