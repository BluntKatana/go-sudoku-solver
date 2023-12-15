package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

type Sudoku struct {
	Id    string
	Board [9][9]int // 0: empy, 1-9: number
	Flag  [9][9]int // 0: not set, 1: set by user, 2: set by program
}

func (s *Sudoku) String() {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Printf("%d ", s.Board[i][j])
		}
		fmt.Println()
	}
}

func (p *Page) save() error {
	filename := "./data/" + strings.ToLower(p.Title) + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "./data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func loadSudoku(title string) (*Sudoku, error) {
	filename := "./sudoku/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// parse a string of 81 characters into a Sudoku
	var board [9][9]int
	for i, c := range body {
		board[i/9][i%9] = int(c - '0')
	}

	var flag [9][9]int
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			flag[i][j] = 0
			if board[i][j] != 0 {
				flag[i][j] = 1
			}
		}
	}

	return &Sudoku{Board: board, Flag: flag, Id: title}, nil
}

func loadAllPages() ([]*Page, error) {
	files, err := os.ReadDir("./data")
	if err != nil {
		return nil, err
	}
	pages := make([]*Page, len(files))
	for i, file := range files {
		title := strings.TrimSuffix(file.Name(), ".txt")
		pages[i], err = loadPage(title)
		if err != nil {
			return nil, err
		}
	}
	return pages, nil
}

func loadAllSudokus() ([]*Sudoku, error) {
	files, err := os.ReadDir("./sudoku")
	if err != nil {
		return nil, err
	}
	sudokus := make([]*Sudoku, len(files))
	for i, file := range files {
		title := strings.TrimSuffix(file.Name(), ".txt")
		sudokus[i], err = loadSudoku(title)
		if err != nil {
			return nil, err
		}
	}
	return sudokus, nil
}

var validPath = regexp.MustCompile("^/(edit|save|view|sudoku)/([a-zA-Z0-9]+)$")

var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html", "./tmpl/all.html", "./tmpl/all-sudokus.html", "./tmpl/sudoku.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("rootHandler")
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewAllHandler(w http.ResponseWriter, r *http.Request) {
	pages, err := loadAllPages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "all.html", pages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewAllSudokuHandler(w http.ResponseWriter, r *http.Request) {
	sudokus, err := loadAllSudokus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "all-sudokus.html", sudokus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewSudokuHandler(w http.ResponseWriter, r *http.Request, title string) {
	s, err := loadSudoku(title)
	fmt.Println("VIEW, title:", title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "sudoku.html", s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		fmt.Println("makeHandler, m:", m, "r.URL.Path:", r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/sudoku/all", viewAllSudokuHandler)
	http.HandleFunc("/sudoku/", makeHandler(viewSudokuHandler))

	http.HandleFunc("/all/", viewAllHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
