package filehandlingandlogging

import (
	"os"
	"testing"
)

func setupTest() {
	// Clean up test file before each run
	_ = os.Remove(dataFile)
	employees = []Employee{}
}

func TestAddEmployee(t *testing.T) {
	setupTest()

	emp := Employee{ID: 1, Name: "Akash", Email: "akash@gmail.com"}
	addEmployee(emp)

	if len(employees) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(employees))
	}
	if employees[0].Name != "Akash" {
		t.Fatalf("expected name Akash, got %s", employees[0].Name)
	}
}

func TestGetEmployee(t *testing.T) {
	setupTest()

	addEmployee(Employee{ID: 2, Name: "John", Email: "john@gmail.com"})
	emp := getEmployee(2)

	if emp == nil {
		t.Fatal("expected employee, got nil")
	}
	if emp.Email != "john@gmail.com" {
		t.Fatalf("expected email john@gmail.com, got %s", emp.Email)
	}
	if emp.Name != "John" {
		t.Fatalf("expected name john, got %s", emp.Name)
	}
}

func TestUpdateEmployee(t *testing.T) {
	setupTest()

	addEmployee(Employee{ID: 3, Name: "abhi", Email: "abhi@gmail.com"})

	updated := Employee{ID: 3, Name: "abhi", Email: "abhibiswas@gmail.com"}
	ok := updateEmployee(updated)

	if !ok {
		t.Fatal("expected true on update")
	}

	emp := getEmployee(3)
	if emp.Name != "abhi" {
		t.Fatalf("expected abhi, got %s", emp.Name)
	}
	if emp.Email != "abhibiswas@gmail.com" {
		t.Fatalf("expected abhibiswas@gmail.com, got %s", emp.Email)
	}
}

func TestDeleteEmployee(t *testing.T) {
	setupTest()

	addEmployee(Employee{ID: 4, Name: "Akash", Email: "akash@gmail.com"})

	ok := deleteEmployee(4)
	if !ok {
		t.Fatal("expected true on delete")
	}
	if len(employees) != 0 {
		t.Fatalf("expected 0 employees, got %d", len(employees))
	}
}

func TestSaveLoadEmployees(t *testing.T) {
	setupTest()

	addEmployee(Employee{ID: 5, Name: "Akash", Email: "akash@gmail.com"})

	// Clear memory
	employees = []Employee{}

	// Load from file
	loadEmployees()

	if len(employees) != 1 {
		t.Fatalf("expected 1 employee after load, got %d", len(employees))
	}
	if employees[0].Name != "Akash" {
		t.Fatalf("expected name Akash, got %s", employees[0].Name)
	}
}
