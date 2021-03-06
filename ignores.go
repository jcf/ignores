package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

var usage = "Usage: ignores <path>\n"

type Ignore struct {
	Lang string `json:"lang"`
	Ref  string `json:"$ref"`
}

func fatalError(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Print(usage)
	os.Exit(1)
}

func serveIgnore(w http.ResponseWriter, r *http.Request, s string) {
	log.Printf("Serving gitignore %s", s)
	http.ServeFile(w, r, s)
}

func serveIgnores(dir string) {
	log.Printf("Serving directory %s", dir)

	files, _ := ioutil.ReadDir(dir)
	ignores := make([]Ignore, 0)

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gitignore") {
			lang := strings.TrimSuffix(f.Name(), ".gitignore")
			route := fmt.Sprintf("/%s", strings.ToLower(lang))
			s := path.Join(dir, f.Name())

			ignores = append(ignores, Ignore{
				lang,
				route,
			})

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
	runtime.GOMAXPROCS(runtime.NumCPU())

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
