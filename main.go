package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	// postgres driver
	_ "github.com/lib/pq" //just for side-effects (initialization)
)

//Todo is todo
type Todo struct {
	ID           int
	Name         string
	Status       string
	LastModified time.Time
	Image        string
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
	http.HandleFunc("/updatestatus", updateStatus)
	http.HandleFunc("/updatetodo", updateTodo)
	http.HandleFunc("/clrcomp", clearCompleted)
	http.HandleFunc("/newtodoup", newTodoup)
	http.HandleFunc("/uploadimg", uploadImg)
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
	fmt.Println("newtodo called")
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

	rows, err := db.Query("SELECT * FROM todo order by ID ASC")
	log.Println(err)

	for rows.Next() {
		var ID int
		var Name string
		var Status string
		var LastModified time.Time
		var Image string
		err = rows.Scan(&ID, &Name, &Status, &LastModified, &Image)
		log.Println(err)
		todo.ID = ID
		todo.Name = Name
		todo.Status = Status
		todo.LastModified = LastModified
		todo.Image = Image
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

func updateStatus(w http.ResponseWriter, r *http.Request) {
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("decode done")
	stmt, err := db.Prepare("UPDATE todo SET Status=$1 WHERE ID=$2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("update prepare done")
	_, err = stmt.Exec(t.Status, t.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("status", t.Status)
	fmt.Printf("Id of updated todo is %d", t.ID)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	var t Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("decode done")
	stmt, err := db.Prepare("UPDATE todo SET Name=$1 WHERE ID=$2")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("update prepare done")
	_, err = stmt.Exec(t.Name, t.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Id of updated todo is %d", t.ID)
}

func clearCompleted(w http.ResponseWriter, r *http.Request) {
	showTodos(w, r)
	stmt, err := db.Prepare("DELETE FROM todo WHERE status=$1")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := stmt.Exec("Completed")
	if err != nil {
		fmt.Println(err)
		return
	}
	count, err := res.RowsAffected()
	if err == nil {
		fmt.Printf("No of todos deleted is %d", count)
	}
}

func newTodoup(w http.ResponseWriter, r *http.Request) {
	var t []Todo
	err := json.NewDecoder(r.Body).Decode(&t)
	if err != nil {
		fmt.Println(err)
		return
	}

	stmt, err := db.Prepare("INSERT INTO todo(Name,Status) VALUES ($1, $2)")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, k := range t {
		fmt.Println(k)
		res, err := stmt.Exec(k.Name, k.Status)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(k.Name, k.Status)
		rowCnt, err := res.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		log.Printf("Rows affected = %d\n", rowCnt)
	}
}

func uploadImg(w http.ResponseWriter, r *http.Request) {
	size := r.Header["Content-Length"][0]
	len, err := strconv.Atoi(size)
	if err == nil {
		if len > 20000 {
			fmt.Println("Image size should be less than 20KB")
		}
	}
	var id = r.FormValue("id")
	file, _, err := r.FormFile("pic")
	out, err := os.Create("public/image/" + id + ".jpg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("create img done")
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success!")
	Image := "public/image/" + id + ".jpg"

	stmt, err := db.Prepare("UPDATE todo SET Image=$1 WHERE ID=$2")
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := stmt.Exec(Image, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("image inserted ID", id)
	rowCnt, err := res.RowsAffected()
	if err != nil {
		log.Println(err)
	}
	log.Printf("Rows affected = %d\n", rowCnt)

}
