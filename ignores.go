package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var usage = "Usage: ignores <path>\n"

func fatalError(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Print(usage)
	os.Exit(1)
}

func prepareIgnorePath(s string) string {
	return strings.ToLower(strings.TrimSuffix(s, ".gitignore"))
}

func serveIgnore(w http.ResponseWriter, r *http.Request, s string) {
	log.Printf("Serving gitignore %s", s)
	http.ServeFile(w, r, s)
}

type Ignore struct {
	Lang string `json:"lang"`
	Ref  string `json:"$ref"`
}

func serveIgnores(dir string) {
	log.Printf("Serving directory %s", dir)

	files, _ := ioutil.ReadDir(dir)
	ignores := make([]Ignore, 0)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gitignore") {
			ignores = append(ignores, Ignore{
				strings.TrimSuffix(f.Name(), ".gitignore"),
				fmt.Sprintf("/%s", prepareIgnorePath(f.Name())),
			})

			spath := prepareIgnorePath(f.Name())
			route := fmt.Sprintf("/%s", spath)
			s := path.Join(dir, f.Name())
			http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
				serveIgnore(w, r, s)
			})
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if bytes, err := json.Marshal(ignores); err == nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprint(w, string(bytes))
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	if len(os.Args) != 2 {
		fatalError(usage)
	}

	arg := os.Args[1]

	if arg == "--help" || arg == "-h" {
		fmt.Print(usage)
		os.Exit(0)
	}

	if info, err := os.Stat(arg); err == nil {
		if info.IsDir() {
			serveIgnores(arg)
		} else {
			fatalError(fmt.Sprintf("ignores: \"%s\" is not a directory\n", arg))
		}
	} else {
		fatalError(fmt.Sprintf("ignores: \"%s\" not found\n", arg))
	}
}
