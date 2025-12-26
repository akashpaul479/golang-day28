package tabledriventesting

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Person struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type MySQLInstance struct {
	DB *sql.DB
}
type RedisInstance struct {
	Client *redis.Client
}

type HybridHandler struct {
	MySQL *MySQLInstance
	Redis *RedisInstance
	Ctx   context.Context
}

func ValidatePerson(person Person) error {
	if person.Email == "" {
		return fmt.Errorf("Email is invalid and empty")
	}
	if strings.TrimSpace(person.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	prefix := strings.TrimSuffix(person.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a suffix before @gmail.com")
	}
	if !strings.HasSuffix(person.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and  must contains @gmail.com")
	}
	if person.Age <= 0 {
		return fmt.Errorf("age must be graten than 0")
	}
	if person.Age >= 100 {
		return fmt.Errorf("Age must be smaller than 100")
	}
	return nil
}

func (h *HybridHandler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	var persons Person
	if err := json.NewDecoder(r.Body).Decode(&persons); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}
	if err := ValidatePerson(persons); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	res, err := h.MySQL.DB.Exec("INSERT INTO persons (name , age , email) VALUES(? , ? , ?)", persons.Name, persons.Age, persons.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	persons.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(persons)
}
func (h *HybridHandler) ReadPerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	value, err := h.Redis.Client.Get(h.Ctx, id).Result()
	if err == nil {
		log.Println("Cache hit!")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}
	fmt.Println("Cache miss Quering MySQL ...")
	row := h.MySQL.DB.QueryRow("SELECT id , name , age , email FROM persons WHERE  id=?", id)

	var persons Person
	if err := row.Scan(&persons.ID, &persons.Name, &persons.Age, &persons.Email); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	jsondata, err := json.Marshal(persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Second)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}
func (h *HybridHandler) UpdatePerson(w http.ResponseWriter, r *http.Request) {
	var persons Person
	if err := json.NewDecoder(r.Body).Decode(&persons); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidatePerson(persons); err != nil {
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}

	res, err := h.MySQL.DB.Exec("UPDATE persons SET name=?,Age=?,email=? WHERE id=?", persons.Name, persons.Age, persons.Email, persons.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, err := json.Marshal(persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Redis.Client.Set(h.Ctx, fmt.Sprint(persons.ID), jsonData, 10*time.Second)

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
func (h *HybridHandler) DeletePerson(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idINT, _ := strconv.Atoi(id)

	res, err := h.MySQL.DB.Exec("DELETE FROM persons WHERE id=?", idINT)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	h.Redis.Client.Del(h.Ctx, id)

	w.Header().Set("content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("person deleted"))

}
func ConnectReddis() (*RedisInstance, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return &RedisInstance{Client: rdb}, nil
}
func ConnectMySQL() (*MySQLInstance, error) {
	db, err := sql.Open("mysql", os.Getenv("MySQL_DSN"))
	if err != nil {
		return nil, err
	}
	return &MySQLInstance{DB: db}, nil
}
func Crudoperation() {
	godotenv.Load()

	redisInstance, err := ConnectReddis()
	if err != nil {
		panic(err)
	}
	mySQLInstance, err := ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handle := &HybridHandler{Redis: redisInstance, MySQL: mySQLInstance, Ctx: context.Background()}

	r := mux.NewRouter()

	r.HandleFunc("/persons", handle.CreatePerson).Methods("POST")
	r.HandleFunc("/persons/{id}", handle.ReadPerson).Methods("GET")
	r.HandleFunc("/persons/{id}", handle.UpdatePerson).Methods("PUT")
	r.HandleFunc("/persons/{id}", handle.DeletePerson).Methods("DELETE")

	fmt.Println("Server running on port:8080")
	http.ListenAndServe(":8080", r)

}
