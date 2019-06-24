package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func postTodosHandler(c *gin.Context) {
	t := Todo{}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	fmt.Println(t)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can't open", err.Error())
	}
	defer db.Close()

	query := `INSERT INTO todos (title, status) VALUES ($1,$2) RETURNING id`
	var id int
	row := db.QueryRow(query, t.Title, t.Status)
	err = row.Scan(&id)
	if err != nil {
		log.Fatal("Can't scan id ", id)
	}
	t.ID = id
	fmt.Println("Insert success id :", id)
	c.JSON(201, t)
}

func getTodosHandler(c *gin.Context) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can't open", err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, title, status FROM todos WHERE id = $1")
	if err != nil {
		log.Fatal("Can't SELECT", err.Error())
	}

	idst := c.Param("id")

	row := stmt.QueryRow(idst)
	t := Todo{}

	err = row.Scan(&t.ID, &t.Title, &t.Status)
	if err != nil {
		log.Fatal("Can't scan", err.Error())
	}
	fmt.Println("Select one row is", t.ID, t.Title, t.Status)
	c.JSON(200, t)
}

func getlistTodosHandler(c *gin.Context) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can't open", err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id, title, status FROM todos")
	if err != nil {
		log.Fatal("Can't select", err.Error())
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Fatal("Can't query", err.Error())
	}

	todos := []Todo{}
	for rows.Next() {
		t := Todo{}
		err := rows.Scan(&t.ID, &t.Title, &t.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		todos = append(todos, t)
	}
	c.JSON(200, todos)
}

func putupdateTodosHandler(c *gin.Context) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can't open", err.Error())
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE todos SET status=$2, title=$3 WHERE id=$1")
	if err != nil {
		log.Fatal("Prepare error ", err.Error())
	}

	idin := c.Param("id")

	t := Todo{}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	t.ID, err = strconv.Atoi(idin)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if _, err := stmt.Exec(idin, t.Status, t.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"exec error": err.Error()})
		return
	}
	fmt.Println("Update success", t.ID, t.Title, t.Status)
	c.JSON(200, t)
}

func deleteTodosByIdHandler(c *gin.Context) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Can't open", err.Error())
	}
	defer db.Close()

	todos := []Todo{}
	query := `DELETE FROM todos WHERE id=$1`
	var id int
	db.QueryRow(query, id)

	fmt.Println("Delete success")
	fmt.Println(todos)
	c.JSON(200, gin.H{
		"status": "success",
	})
}

func main() {
	r := gin.Default()
	r.POST("api/todos", postTodosHandler)
	r.GET("api/todos/:id", getTodosHandler)
	r.GET("api/todos", getlistTodosHandler)
	r.PUT("api/todos/:id", putupdateTodosHandler)
	r.DELETE("api/todos/:id", deleteTodosByIdHandler)
	r.Run(":1234")
}
