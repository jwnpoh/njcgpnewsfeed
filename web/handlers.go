package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

)

func index(w http.ResponseWriter, r *http.Request) {
    data := *s.Articles
    data = data[0:15]

	err := tpl.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Fatal("unable to execute template - ", err)
	}
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	defer os.Exit(0)
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	filename := r.Form.Get("download")
	rmdir := r.Form.Get("remove")

	defer os.RemoveAll(rmdir)

	fmt.Printf("File ready for download. Cleaning up temporary files....\n==> ")
	fmt.Println("Done!")

	filenamebase := filepath.Base(filename)
	w.Header().Set("Content-Disposition", "attachment; filename="+filenamebase)
	http.ServeFile(w, r, filename)
}
