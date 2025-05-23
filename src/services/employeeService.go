package services

import (
	"database/sql" // เพิ่มการ import
	"fmt"
	"strconv"
	"time"
	"github.com/NinePTH/GO_MVC-S/src/models"
)
func GetEmployeeSearch(id string,first_name string,last_name string) ([]models.EmployeeResponse, error) {
	selectedColumns := []string{
		"Employee.employee_id",
		"Employee.first_name",
		"Employee.last_name",
		"Position.position_name",
		"Employee.phone_number",
		"Department.department_name",
		"Employee.salary",
		"Employee.email",
		"Employee.hire_date",
		"Employee.resignation_date",
		"Employee.work_status",
	}
	joinTables := "Position ON employee.position_id = position.position_id JOIN Department ON position.department_id = department.department_id"

	results, err := SelectData(
		"Employee",      // main table
		selectedColumns, // columns to select
		true,           // ไม่มี WHERE
		"($1 = '' OR employee_id = $1) AND ($2 = '' OR first_name ILIKE '%' || $2 || '%') AND ($3 = '' OR last_name ILIKE '%' || $3 || '%')",              // เงื่อนไข WHERE ว่าง
		[]interface{}{id,first_name,last_name},             // ไม่มี args
		true,            // ใช้ JOIN
		joinTables,      // joinTables ที่รวม INNER JOIN ไว้แล้ว
		"",              // joinCondition เว้นว่าง
		"ORDER BY employee_id DESC",
	)

	if err != nil {
		return nil, err
	}

	var employees []models.EmployeeResponse

	for _, row := range results {
		// จัดการกับ resignation_date ที่อาจเป็น NULL
		var resignationDate sql.NullTime
		if row["resignation_date"] != nil {
			resignationDate = sql.NullTime{Time: row["resignation_date"].(time.Time), Valid: true}
		}

		// ตรวจสอบ resignation_date ว่ามีค่าเป็น "0001-01-01T00:00:00Z" หรือไม่
		if resignationDate.Valid && resignationDate.Time.Year() == 1 && resignationDate.Time.Month() == 1 && resignationDate.Time.Day() == 1 {
			resignationDate.Valid = false // ถ้าเป็นค่า default ให้ถือว่าเป็น NULL
		}

		// แปลง salary ให้เป็น float64
		salary := parseSalary(row["salary"])

		// แปลง work_status ให้เป็น string
		workStatus := string(row["work_status"].([]byte))

		// ตรวจสอบ resignation_date และถ้า NULL หรือค่ามาตรฐานให้เปลี่ยนเป็นข้อความ
		var resignationDateStr string
		if resignationDate.Valid {
			// ถ้า resignation_date มีค่า, format เป็น YYYY-MM-DD
			resignationDateStr = resignationDate.Time.Format("2006-01-02")
		} else {
			resignationDateStr = "Not resigned yet" // ถ้าไม่มีการลาออก
		}

		// แปลง hire_date เป็น string
		hireDate := row["hire_date"].(time.Time).Format("2006-01-02")

		// สร้าง struct ของ Employee
		employee := models.EmployeeResponse{
			Employee_id:      fmt.Sprintf("%v", row["employee_id"]),
			First_name:       fmt.Sprintf("%v", row["first_name"]),
			Last_name:        fmt.Sprintf("%v", row["last_name"]),
			Position_name:    fmt.Sprintf("%v", row["position_name"]),
			Phone_number:     fmt.Sprintf("%v", row["phone_number"]),
			Department_name:  fmt.Sprintf("%v", row["department_name"]),
			Salary:           salary, // salary ที่แปลงแล้ว
			Email:            fmt.Sprintf("%v", row["email"]),
			Hire_date:        hireDate,           // แปลง hire_date เป็นรูปแบบที่ต้องการ
			Resignation_date: resignationDateStr, // เปลี่ยนเป็นข้อความตามที่ต้องการ
			Work_status:      workStatus,         // work_status ที่แปลงแล้ว
		}

		employees = append(employees, employee)
		fmt.Println(employee)
	}

	return employees, nil
}

func UpdateEmployee(id string, data map[string]interface{}) (int64, error) {
	table := "Employee"
	condition := "employee_id = $1"
	conditionValues := []interface{}{id}

	// Call UpdateData with correct parameters
	rowsAffected, err := UpdateData(table, data, condition, conditionValues)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func AddEmployee(data map[string]interface{}) (int64, error) {
	table := "Employee"
	rowsAffected, err := InsertData(table, data)
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func GetEmployee(employeeID string) (*models.EmployeeResponse, error) {
	fields := []string{
		"Employee.employee_id",
		"Employee.first_name",
		"Employee.last_name",
		"Position.position_name",
		"Employee.phone_number",
		"Department.department_name",
		"Employee.salary",
		"Employee.email",
		"Employee.hire_date",
		"Employee.resignation_date",
		"Employee.work_status",
	}
	// กำหนด joinTables แบบ JOIN statement
	joinTables := "Position ON employee.position_id = position.position_id JOIN Department ON position.department_id = department.department_id"

	// WHERE clause + arguments
	whereCondition := "Employee.employee_id = $1"
	args := []interface{}{employeeID}

	results, err := SelectData(
		"Employee",     // main table
		fields,         // []string เช่น []{"Employee.name", "Department.name", "Position.title"}
		true,           // where
		whereCondition, // เช่น "Employee.salary > ?"
		args,           // []interface{}{50000}
		true,           // join
		joinTables,     // joinTable ที่รวม INNER JOIN ไว้
		"",             // joinCondition เว้นว่างเพราะรวมไว้แล้ว
		"",
	)

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("employee not found")
	}

	row := results[0]

	// แปลง resignation_date
	var resignationDateStr string
	if row["resignation_date"] == nil {
		resignationDateStr = "Not resigned yet"
	} else {
		date := row["resignation_date"].(time.Time)
		if date.IsZero() || date.Year() == 1 {
			resignationDateStr = "Not resigned yet"
		} else {
			resignationDateStr = date.Format("2006-01-02")
		}
	}

	salary := parseSalary(row["salary"])
	workStatus := string(row["work_status"].([]byte))
	hireDate := row["hire_date"].(time.Time).Format("2006-01-02")

	employee := &models.EmployeeResponse{
		Employee_id:      fmt.Sprintf("%v", row["employee_id"]),
		First_name:       fmt.Sprintf("%v", row["first_name"]),
		Last_name:        fmt.Sprintf("%v", row["last_name"]),
		Position_name:    fmt.Sprintf("%v", row["position_name"]),
		Phone_number:     fmt.Sprintf("%v", row["phone_number"]),
		Department_name:  fmt.Sprintf("%v", row["department_name"]),
		Salary:           salary,
		Email:            fmt.Sprintf("%v", row["email"]),
		Hire_date:        hireDate,
		Resignation_date: resignationDateStr,
		Work_status:      workStatus,
	}

	return employee, nil
}

func GetAllEmployee() ([]models.EmployeeResponse, error) {
	selectedColumns := []string{
		"Employee.employee_id",
		"Employee.first_name",
		"Employee.last_name",
		"Position.position_name",
		"Employee.phone_number",
		"Department.department_name",
		"Employee.salary",
		"Employee.email",
		"Employee.hire_date",
		"Employee.resignation_date",
		"Employee.work_status",
	}
	joinTables := "Position ON employee.position_id = position.position_id JOIN Department ON position.department_id = department.department_id"

	results, err := SelectData(
		"Employee",      // main table
		selectedColumns, // columns to select
		false,           // ไม่มี WHERE
		"",              // เงื่อนไข WHERE ว่าง
		nil,             // ไม่มี args
		true,            // ใช้ JOIN
		joinTables,      // joinTables ที่รวม INNER JOIN ไว้แล้ว
		"",              // joinCondition เว้นว่าง
		"ORDER BY employee_id DESC",
	)

	if err != nil {
		return nil, err
	}

	var employees []models.EmployeeResponse

	for _, row := range results {
		// จัดการกับ resignation_date ที่อาจเป็น NULL
		var resignationDate sql.NullTime
		if row["resignation_date"] != nil {
			resignationDate = sql.NullTime{Time: row["resignation_date"].(time.Time), Valid: true}
		}

		// ตรวจสอบ resignation_date ว่ามีค่าเป็น "0001-01-01T00:00:00Z" หรือไม่
		if resignationDate.Valid && resignationDate.Time.Year() == 1 && resignationDate.Time.Month() == 1 && resignationDate.Time.Day() == 1 {
			resignationDate.Valid = false // ถ้าเป็นค่า default ให้ถือว่าเป็น NULL
		}

		// แปลง salary ให้เป็น float64
		salary := parseSalary(row["salary"])

		// แปลง work_status ให้เป็น string
		workStatus := string(row["work_status"].([]byte))

		// ตรวจสอบ resignation_date และถ้า NULL หรือค่ามาตรฐานให้เปลี่ยนเป็นข้อความ
		var resignationDateStr string
		if resignationDate.Valid {
			// ถ้า resignation_date มีค่า, format เป็น YYYY-MM-DD
			resignationDateStr = resignationDate.Time.Format("2006-01-02")
		} else {
			resignationDateStr = "Not resigned yet" // ถ้าไม่มีการลาออก
		}

		// แปลง hire_date เป็น string
		hireDate := row["hire_date"].(time.Time).Format("2006-01-02")

		// สร้าง struct ของ Employee
		employee := models.EmployeeResponse{
			Employee_id:      fmt.Sprintf("%v", row["employee_id"]),
			First_name:       fmt.Sprintf("%v", row["first_name"]),
			Last_name:        fmt.Sprintf("%v", row["last_name"]),
			Position_name:    fmt.Sprintf("%v", row["position_name"]),
			Phone_number:     fmt.Sprintf("%v", row["phone_number"]),
			Department_name:  fmt.Sprintf("%v", row["department_name"]),
			Salary:           salary, // salary ที่แปลงแล้ว
			Email:            fmt.Sprintf("%v", row["email"]),
			Hire_date:        hireDate,           // แปลง hire_date เป็นรูปแบบที่ต้องการ
			Resignation_date: resignationDateStr, // เปลี่ยนเป็นข้อความตามที่ต้องการ
			Work_status:      workStatus,         // work_status ที่แปลงแล้ว
		}

		employees = append(employees, employee)
		fmt.Println(employee)
	}

	return employees, nil
}

// ใช้แปลง salary เป็น float64
func parseSalary(data interface{}) float64 {
	salaryStr := string(data.([]byte)) // "52000.00"
	salaryFloat, err := strconv.ParseFloat(salaryStr, 64)
	if err != nil {
		fmt.Println("Error parsing salary:", err)
		return 0
	}
	return salaryFloat
}
