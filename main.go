package main

import (
	// "fmt"
	"context"
	"log"
	// "os"

	"github.com/jwnpoh/njcgpnewsfeed/db"
	"github.com/jwnpoh/njcgpnewsfeed/web"
)

func main() {
    ctx := context.Background()
    database := db.InitArticlesDBByDate()
    qnDB, err := db.InitQuestionsDB()
    if err != nil {
        log.Fatal(err)
    }

    if err := database.SetupArticlesDB(ctx, qnDB); err != nil {
        log.Fatal(err)
    }

    s := web.NewServer(database, qnDB)

    s.Port = "8080"
    s.TemplateDir = "html"

    log.Fatal(s.Start())
}

// func main() {
//     ctx := context.Background()
// 
//     Database, err := db.InitArticlesDBByDate()
//     if err != nil {
//         log.Fatal(err)
//     }
// 
//     qnDB := db.NewQnDB()
//     if err := db.MapQuestions(qnDB); err != nil {
//         log.Fatal(err)
//     }
// 
//     if err := db.SetupArticlesDB(ctx, Database, qnDB); err != nil {
//         log.Fatal(err)
//     }
// 
//     term := "best"
//     results, err := web.Search(term, Database)
//     if err != nil {
//         fmt.Println(err)
//     } else {
//         for _, j := range *results {
//             fmt.Printf("%s\n%s\n%s\n%v\n%v\n\n", j.DisplayDate, j.Title, j.URL, j.Topics, j.Questions)
//         }
//     }
// 
//     if err := db.BackupToSheets(ctx, Database); err != nil {
//         log.Fatal(err)
//     }
// 
// }

// func myTest(qnDB *db.QuestionsDB, Database *db.ArticlesDBByDate) {
//     a, err := db.NewArticle() 
//     if err != nil {
//         log.Fatal(err)
//     }
//     a.Title = "How to Train Your Dragon"
//     a.URL = "http://www.httyd.com"
//     a.SetTopics("Dragons", "Fantasy")
//     a.SetQuestions(2020, 2, qnDB)
//     a.SetQuestions(2020, 3, qnDB)
//     a.SetDate("Jan 27 2020")
// 
//     a.AddArticleToDB(Database)
// 
//     b, err := db.NewArticle() 
//     if err != nil {
//         log.Fatal(err)
//     }
//     b.Title = "How to Train Your Dodo"
//     b.URL = "http://www.httyd.com"
//     b.SetTopics("Dodo", "Fantasy")
//     b.SetQuestions(2020, 5, qnDB)
//     b.SetDate("Mar 25 2021")
// 
//     b.AddArticleToDB(Database)
// 
//     c, err := db.NewArticle() 
//     if err != nil {
//         log.Fatal(err)
//     }
//     c.Title = "How to Train Your Dog"
//     c.URL = "http://www.httyd.com"
//     c.SetTopics("Dog", "Food")
//     c.SetQuestions(2020, 6, qnDB)
//     c.SetQuestions(2020, 7, qnDB)
//     c.SetDate("Feb 26 2018")
// 
//     c.AddArticleToDB(Database)
// 
//     d, err := db.NewArticle() 
//     if err != nil {
//         log.Fatal(err)
//     }
//     d.Title = "How to Train Your Duck"
//     d.URL = "http://www.httyd.com"
//     d.SetTopics("Duck", "Birds")
//     d.SetQuestions(2005, 6, qnDB)
//     d.SetQuestions(2010, 7, qnDB)
//     d.SetDate("May 26 2020")
// 
//     d.AddArticleToDB(Database)
// 
//     term := "what"
//     results, err := web.Search(term, Database)
//     if err != nil {
//         fmt.Println(err)
//     } else {
//         for _, j := range *results {
//             fmt.Printf("%s\n%s\n%s\n%v\n%v\n\n", j.DisplayDate, j.Title, j.URL, j.Topics, j.Questions)
//         }
//     }
// }
