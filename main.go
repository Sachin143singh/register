package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	// "os"

	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/context"

	// "github.com/joho/godotenv"
	// _ "github.com/joho/godotenv" //new
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	firstname string
	lastname  string
	email     string
	password  string
	// createdDate time.Time `json:"createdDate"`
}

func dbConn() (db *sql.DB) {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root"
	dbName := "register"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("templates/*.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db := dbConn()
		firstname := r.FormValue("FirstName")
		lastname := r.FormValue("LastName")
		email := r.FormValue("email")
		fmt.Printf("%s, %s, %s\n", firstname, lastname, email)

		password, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			fmt.Println(err)
			tmpl.ExecuteTemplate(w, "Register", err)
		}

		dt := time.Now()

		createdDateString := dt.Format("2006-01-02 15:04:05")

		// Convert the time before inserting into the database
		createdDate, err := time.Parse("2006-01-02 15:04:05", createdDateString)
		if err != nil {
			log.Fatal("Error converting the time:", err)
		}

		_, err = db.Exec("INSERT INTO user(firstname, lastname,email,password,createdDate) VALUES(?,?,?,?,?)", firstname, lastname, email, password, createdDate)
		if err != nil {
			fmt.Println("Error when inserting: ", err.Error())
			panic(err.Error())
		}
		log.Println("=> Inserted: First Name: " + firstname + " | Last Name: " + lastname)

		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	} else if r.Method == "GET" {
		tmpl.ExecuteTemplate(w, "register.html", nil)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db := dbConn()
		email := r.FormValue("email")
		password := r.FormValue("password")
		fmt.Printf("%s, %s\n", email, password)

		if strings.Trim(email, " ") == "" || strings.Trim(password, " ") == "" {
			fmt.Println("Parameter's can't be empty")
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
			return
		}

		checkUser, err := db.Query("SELECT id, createdDate, password, firstname, lastname, email FROM user WHERE email=?", email)
		if err != nil {
			panic(err.Error())
		}
		user := &User{}
		for checkUser.Next() {
			var id int
			var password, firstname, lastname, email, createdDate string
			// var createdDate time.Time
			// err = checkUser.Scan(&id, &createdDate, &password, &firstname, &lastname, &email)
			err = checkUser.Scan(&id, &password, &firstname, &lastname, &email, &createdDate)
			if err != nil {
				panic(err.Error())
			}
			user.ID = id
			user.firstname = firstname
			user.lastname = lastname
			user.email = email
			user.password = password
			// user.createdDate = createdDate
		}

		errf := bcrypt.CompareHashAndPassword([]byte(user.password), []byte(password))
		if errf != nil && errf == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
			fmt.Println(errf)
			http.Redirect(w, r, "/login", http.StatusMovedPermanently)
		} else {
			tmpl.ExecuteTemplate(w, "dashboard.html", user)
			return
		}
	} else if r.Method == "GET" {
		tmpl.ExecuteTemplate(w, "login.html", nil)
	}
}

// func logoutHandler(w http.ResponseWriter, r *http.Request) {
// 	http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
// }

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "dashboard.html", nil)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	// http.HandleFunc("/logouth", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started on: http://localhost:7996")
	err := http.ListenAndServe(":9999", context.ClearHandler(http.DefaultServeMux)) // context to prevent memory leak
	if err != nil {
		log.Fatal(err)
	}
}
