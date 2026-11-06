package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Task struct {
	ID        int
	Title     string
	Completed bool
}
type PageData struct {
	Tasks                                    []Task
	TotalTasks, CompletedTasks, PendingTasks int
}

var tasks []Task
var nextID int
var mutex sync.Mutex
var page = template.Must(template.ParseFiles("templates/index.html"))

// Передаём список и статистику в HTML-шаблон.
func home(responseWriter http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(responseWriter, request)
		return
	}
	if request.Method != http.MethodGet {
		http.Error(responseWriter, "Method not allowed", 405)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	data := PageData{Tasks: tasks, TotalTasks: len(tasks)}
	for _, task := range tasks {
		if task.Completed {
			data.CompletedTasks = data.CompletedTasks + 1
		}
	}
	data.PendingTasks = data.TotalTasks - data.CompletedTasks
	page.Execute(responseWriter, data)
}

func add(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "Method not allowed", 405)
		return
	}
	title := strings.TrimSpace(request.FormValue("title"))
	if title == "" {
		http.Error(responseWriter, "Введите название задачи", 400)
		return
	}
	mutex.Lock()
	nextID = nextID + 1
	tasks = append(tasks, Task{ID: nextID, Title: title})
	mutex.Unlock()
	http.Redirect(responseWriter, request, "/", http.StatusSeeOther)
}

// Находим задачу по ID, затем меняем статус или удаляем её.
func change(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "Method not allowed", 405)
		return
	}
	id, inputError := strconv.Atoi(request.FormValue("id"))
	if inputError != nil {
		http.Error(responseWriter, "Некорректный ID", 400)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	for index := range tasks {
		if tasks[index].ID == id {
			if request.URL.Path == "/tasks/delete" {
				tasks = append(tasks[:index], tasks[index+1:]...)
			} else {
				tasks[index].Completed = !tasks[index].Completed
			}
			http.Redirect(responseWriter, request, "/", http.StatusSeeOther)
			return
		}
	}
	http.Error(responseWriter, "Задача не найдена", 404)
}

func about(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(responseWriter, "Method not allowed", 405)
		return
	}
	template.Must(template.ParseFiles("templates/about.html")).Execute(responseWriter, nil)
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/tasks", add)
	http.HandleFunc("/tasks/complete", change)
	http.HandleFunc("/tasks/delete", change)
	http.HandleFunc("/about", about)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
