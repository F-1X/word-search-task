package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"word-search-in-files/pkg/searcher"
)

// запустить приложение - go run main.go

func main() {
	// хардкод os.DirFS(".") статичный маршрут, по умолчанию, где произвонится индексация, тк экземпляр Searcher и мапа индексов  зависят друг от друга,
	// как вариант синхронизации использовать БД fullPath/fileName : map индексов например redis для нескольких инстансов приложения, с Append Only File.

	searcher := searcher.NewSearcher(os.DirFS("."))
	NewServer(searcher).Run()

}

type Server struct {
	searcher *searcher.Searcher
}

func NewServer(searcher *searcher.Searcher) *Server {
	return &Server{
		searcher: searcher,
	}
}

func (s *Server) Run() {
	http.HandleFunc("/files/search", s.searchHandler)
	log.Println("Server started on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	// curl "localhost:8888/files/search?&dir=examples&word=летние" -k -vvv
	// dir указывается конкретная директория, по умолчанию все директории ".", не обязательный параметр. word слово, обязательный параметр
	dir := r.URL.Query().Get("dir")
	word := r.URL.Query().Get("word")
	recursive := r.URL.Query().Get("R")

	// рукурсивный поиск по директориям. до умолчанию True- например examples/file... , examples/a , при dir=examples, R=false вывод будет examples/file...
	if recursive == "false" {
		s.searcher.Recursive = false
	} else {
		s.searcher.Recursive = true
	}

	if word == "" {
		http.Error(w, "Missing 'word' parameter", http.StatusBadRequest)
		return
	}

	if dir == "" {
		dir = "."
	}
	s.searcher.Dir = dir

	result, err := s.searcher.Search(word)

	if err != nil {
		if err.Error() == "stat .: no such file or directory" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, "Dir not found")
			return
		}
		if len(result) == 0 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, "Word not found")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
