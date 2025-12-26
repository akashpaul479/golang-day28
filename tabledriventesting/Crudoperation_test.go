package tabledriventesting_test

import (
	"bytes"
	"encoding/json"
	"filehandlingandlogging/tabledriventesting"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
)

func TestHybridHandler_CreatePerson(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		person   tabledriventesting.Person
		willpass bool
	}{
		{
			name: "valid name , age and valid email",
			person: tabledriventesting.Person{
				Name:  "kunal",
				Age:   21,
				Email: "Kunal@gmail.com",
			},
			willpass: true,
		},
		{
			name: "invalid name  and valid age ,email",
			person: tabledriventesting.Person{
				Name:  "",
				Age:   21,
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name ,invalid age and valid email",
			person: tabledriventesting.Person{
				Name:  "Akash",
				Age:   -1,
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name , age and invalid email",
			person: tabledriventesting.Person{
				Name:  "Akash",
				Age:   21,
				Email: "",
			},
			willpass: false,
		},
		{
			name: "valid name , age and email without prefix",
			person: tabledriventesting.Person{
				Name:  "Akash",
				Age:   21,
				Email: "@gmail.com",
			},
			willpass: false,
		},
		{
			name: "withspace name and valid  age , valid email",
			person: tabledriventesting.Person{
				Name:  "   ",
				Age:   21,
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
	}
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/migrations_db")

	redisInstance, err := tabledriventesting.ConnectReddis()
	if err != nil {
		panic(err)
	}
	mySQLInstance, err := tabledriventesting.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handle := &tabledriventesting.HybridHandler{Redis: redisInstance, MySQL: mySQLInstance}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.person)
			if err != nil {
				log.Panic("Failed to marshal!")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/persons", buffer)
			w := httptest.NewRecorder()

			handle.CreatePerson(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var person tabledriventesting.Person
				if err := json.NewDecoder(w.Body).Decode(&person); err != nil {
					t.Fatalf("Failed to decode response:%v ", err)
				}
				if person.Name != tt.person.Name {
					t.Fatalf("Expected  name %s, got %s", person.Name, tt.person.Name)
				}
				if person.Age != tt.person.Age {
					t.Fatalf("Expected email %d, got %d", person.Age, tt.person.Age)
				}
				if person.Email != tt.person.Email {
					t.Fatalf("Expected email %s, got %s", person.Email, tt.person.Email)
				}
				if person.ID == 0 {
					t.Fatalf("expected non zero ID")
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler_ReadPerson(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id exists in mysql",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid id ",
			id:       6348,
			willpass: false,
		},
	}
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MySQL_DSN", "root:root@tcp(127.0.0.1:3306)/migrations_db")

	redisInstance, err := tabledriventesting.ConnectReddis()
	if err != nil {
		panic(err)
	}
	mySQLInstance, err := tabledriventesting.ConnectMySQL()
	if err != nil {
		panic(err)
	}
	handle := &tabledriventesting.HybridHandler{Redis: redisInstance, MySQL: mySQLInstance}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle.MySQL.DB.Exec("DELETE FROM persons")
			handle.Redis.Client.FlushAll(handle.Ctx)

			if tt.willpass {
				res, err := handle.MySQL.DB.Exec("INSERT INTO persons (name , age , email) VALUES (? , ? , ?)", "Akash", 21, "akash@gmail.com")
				if err != nil {
					t.Fatal(err)
				}
				id, _ := res.LastInsertId()
				tt.id = int(id)
			}

			r := httptest.NewRequest(http.MethodGet, "/persons"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			handle.ReadPerson(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				var person tabledriventesting.Person
				if err := json.NewDecoder(w.Body).Decode(&person); err != nil {
					t.Fatalf("failed to Decode response:%v", err)
				}
				if person.ID != tt.id {
					t.Fatalf("Expected %d , got %d", person.ID, tt.id)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status , got %d", w.Code)
				}
			}
		})
	}
}
