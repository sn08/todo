package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// postgres driver
	_ "github.com/lib/pq" //just for side-effects (initialization)
)

//Todo is todo
type Todo struct {
	ID     int
	Name   string
	Status string
}

var db *sql.DB

func init() {
	// postgresql init
	psqlInfo := fmt.Sprintf("host=localhost port=5432 user=postgres " +
		"password=snehaa dbname=todolist sslmode=disable")
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully connected!")
}

func main() {
	http.HandleFunc("/newtodo", newTodo)
	http.HandleFunc("/showtodo", showTodos)
	http.HandleFunc("/deletetodo", deleteTodo)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	err := http.ListenAndServe(":8080", nil) //as it's dynamic we have to declare a separate handle
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}

func newTodo(w http.ResponseWriter, r *http.Request) {
	// JSON format(unmarshal)
	// validate todo
	// save it to DB
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(t)
	fmt.Println(t.Name)

	stmt, err := db.Prepare("INSERT INTO todo(Name,Status) VALUES ($1, $2)")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := stmt.Exec(t.Name, t.Status)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(t.Name, t.Status)
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Rows affected = %d\n", rowCnt)
}

func showTodos(w http.ResponseWriter, r *http.Request) {
	var todos []Todo
	todo := Todo{}

	rows, err := db.Query("SELECT * FROM todo")
	log.Println(err)

	for rows.Next() {
		var ID int
		var Name string
		var Status string
		err = rows.Scan(&ID, &Name, &Status)
		log.Println(err)
		todo.ID = ID
		todo.Name = Name
		todo.Status = Status
		todos = append(todos, todo)
	}
	//marshal
	todolistBytes, err := json.Marshal(todos)
	if err != nil {
		fmt.Println(fmt.Errorf("Error: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//write onto header
	w.Write(todolistBytes)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	showTodos(w, r)
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		return
	}
	stmt, err := db.Prepare("DELETE FROM todo WHERE id=$1")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := stmt.Exec(t.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	count, err := res.RowsAffected()
	if err == nil {
		fmt.Printf("No of todos deleted is %d and id is %d", count, t.ID)
	}
}
