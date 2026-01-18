package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)


var cacheMutex sync.RWMutex
var tableName string


func main(){
	tableName = "users"
	initDBFile("userDB")
	createTable(tableName, `id INTEGER PRIMARY KEY , name TEXT NOT NULL, email TEXT NOT NULL UNIQUE, password TEXT NOT NULL`);
	port := ":" + os.Args[1]
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users", getUsers)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)
	mux.HandleFunc("PUT /users/{id}", updateUser)
	mux.HandleFunc("POST /login", login)
	mux.HandleFunc("GET /verify/{token}", verify)

	fmt.Printf("Listening on port %s...\n", port)
	http.ListenAndServe(port, mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hello");
}
func getUser(w http.ResponseWriter, r *http.Request){
	id, err := strconv.Atoi(r.PathValue("id"));
	if err != nil {
		http.Error(w, "ID must be an integer", http.StatusBadRequest)
	}
	sqlStmt := fmt.Sprintf("SELECT id, name, email FROM %s where id = %d;", tableName, id)
	res, err := DB.Query(sqlStmt);
	var user User
	if err != nil{
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	for res.Next(){
		if err := res.Scan(&user.ID, &user.Name, &user.Email); err != nil{
		  http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
    w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusFound)
	w.Write(userJson)
}
func getUsers(w http.ResponseWriter, r *http.Request){

	sqlStmt := fmt.Sprintf("SELECT id, email, name FROM %s;", tableName)
	res, err := DB.Query(sqlStmt);
	users := []User{}


	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
    for res.Next(){
		var user User
		if err := res.Scan(&user.ID, &user.Name, &user.Email); err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	
		users = append(users, user)
	}
	usersJson, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
    w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(usersJson)
}

func createUser(w http.ResponseWriter,r *http.Request){
	var user User
	err := json.NewDecoder(r.Body).Decode(&user);
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest);
	}
	if user.Name == ""{
		http.Error(w, "Name is required.", http.StatusBadRequest)
	}
	if user.Email == ""{
		http.Error(w, "Email is required.", http.StatusBadRequest)
	}

  	hashed, err := hashPassword(user.Password)

	if err != nil{
		http.Error(w, "Email is required.", http.StatusInternalServerError)
	}

    sqlStmt := "insert into users(name, email, password) values(?, ?, ?);"
	
	cacheMutex.Lock()
    _, err = DB.Exec(sqlStmt, user.Name, user.Email, hashed) 
	cacheMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request){
	id, err := strconv.Atoi(r.PathValue("id"))
	
    if err != nil {
		http.Error(w, "ID must be an integer", http.StatusBadRequest)
	}
  	sqlStmt := "delete from users where id = ? or email = ?;"
	
	cacheMutex.Lock()
    _, err = DB.Exec(sqlStmt, id, id) 
	cacheMutex.Unlock()

	fmt.Fprintf(w, "User with (id=%d) deleted successfully.", id)
}

func updateUser(w http.ResponseWriter, r *http.Request){
	id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil {
		http.Error(w, "ID must be an integer", http.StatusBadRequest)
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest);
	}

  	sqlStmt := fmt.Sprintf("update %s set name = ?, email = ? where id = ?", tableName)
	
	cacheMutex.Lock()
    _, err = DB.Exec(sqlStmt, user.Name, user.Email, id) 
	if err != nil{
		log.Panic(err.Error())
	}
	cacheMutex.Unlock()

	fmt.Fprintf(w, "User with (id=%d) updated successfully.", id)
}

func login(w http.ResponseWriter, r *http.Request){
	var reqUser User
	 err := json.NewDecoder(r.Body).Decode(&reqUser)
	 if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	 }
	 if reqUser.Email == ""{
		  http.Error(w, "Email is required", http.StatusBadRequest)
		  return
	 }
	 if reqUser.Password == ""{
		  http.Error(w, "Password is required", http.StatusBadRequest)
		  return
	 }
	sqlStmt := fmt.Sprintf("SELECT email, password FROM %s where email = ?;", tableName)
	res, err := DB.Query(sqlStmt, reqUser.Email);
	var user User
	if err != nil{
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	for res.Next(){
		if err := res.Scan(&user.Email, &user.Password); err != nil{
		  http.Error(w, err.Error(), http.StatusInternalServerError)
		  return
		}
	}
	if !comparePassword(user.Password, reqUser.Password){
		http.Error(w, "The password you entered is incorrect!!", http.StatusUnauthorized)
		return
	}
    
	token, err := createToken(user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError);
		return
	}
	w.Write([]byte(token))
}

func verify(w http.ResponseWriter, r *http.Request){
	token := r.PathValue("token")
	err := verifyToken(token) 
	w.Write([]byte(strconv.FormatBool(err == nil)))
}

