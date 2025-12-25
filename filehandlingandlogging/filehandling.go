package filehandlingandlogging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Employee struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var employees []Employee

const dataFile = "employees.json"

// Load employees from JSON file
func loadEmployees() {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			employees = []Employee{}
			return
		}
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&employees)
	if err != nil && err.Error() != "EOF" {
		panic(err)
	}
}

// Save employees to JSON file
func saveEmployees() {
	file, err := os.Create(dataFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(employees)
	if err != nil {
		panic(err)
	}
}

// Add employee
func addEmployee(emp Employee) {
	employees = append(employees, emp)
	saveEmployees()
}

// Get employee by ID
func getEmployee(id int) *Employee {
	for _, emp := range employees {
		if emp.ID == id {
			return &emp
		}
	}
	return nil
}

// Update employee
func updateEmployee(updated Employee) bool {
	for i, emp := range employees {
		if emp.ID == updated.ID {
			employees[i] = updated
			saveEmployees()
			return true
		}
	}
	return false
}

// Delete employee
func deleteEmployee(id int) bool {
	for i, emp := range employees {
		if emp.ID == id {
			employees = append(employees[:i], employees[i+1:]...)
			saveEmployees()
			return true
		}
	}
	return false
}
func waitforkey(reader *bufio.Reader) {
	fmt.Println("press space or enter to continue...")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" || input == " " || input == "\t" {
		return
	}
}

func FileHandling() {
	loadEmployees()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- Employee Management ---")
		fmt.Println("1. Add Employee")
		fmt.Println("2. Get Employee")
		fmt.Println("3. Update Employee")
		fmt.Println("4. Delete Employee")
		fmt.Println("5. List Employees")
		fmt.Println("6. Exit")
		fmt.Print("Choose an option: ")

		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)
		choice, _ := strconv.Atoi(choiceStr)

		switch choice {
		case 1:
			fmt.Print("Enter ID: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			id, _ := strconv.Atoi(idStr)

			fmt.Print("Enter Name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Enter email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)

			addEmployee(Employee{ID: id, Name: name, Email: email})
			fmt.Println("Employee added.")
			waitforkey(reader)

		case 2:
			fmt.Print("Enter ID to get: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			id, _ := strconv.Atoi(idStr)

			emp := getEmployee(id)
			if emp != nil {
				fmt.Printf("Found: %+v\n", *emp)
			} else {
				fmt.Println("Employee not found.")
			}
			waitforkey(reader)

		case 3:
			fmt.Print("Enter ID to update: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			id, _ := strconv.Atoi(idStr)

			fmt.Print("Enter New Name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			fmt.Print("Enter New email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)

			success := updateEmployee(Employee{ID: id, Name: name, Email: email})
			if success {
				fmt.Println("Employee updated.")
			} else {
				fmt.Println("Employee not found.")
			}
			waitforkey(reader)

		case 4:
			fmt.Print("Enter ID to delete: ")
			idStr, _ := reader.ReadString('\n')
			idStr = strings.TrimSpace(idStr)
			id, _ := strconv.Atoi(idStr)

			if deleteEmployee(id) {
				fmt.Println("Employee deleted.")
			} else {
				fmt.Println("Employee not found.")
			}
			waitforkey(reader)

		case 5:
			fmt.Println("Employees:")
			for _, emp := range employees {
				fmt.Printf("%+v\n", emp)
			}
			waitforkey(reader)

		case 6:
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("Invalid choice.")
		}
	}
}
