package services

import (
	"fmt"
	"strings"

	//"strings"
	"time"

	"github.com/NinePTH/GO_MVC-S/src/models/patients"
)

func isValidString(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && strings.ToLower(s) != "undefined" && strings.ToLower(s) != "null"
}

func GetPatientSearch(id string, first_name string, last_name string) ([]patients.GetPatientResponse, error) {
	table := "Patient"
	fields := []string{"*"}

	results, err := SelectData(table, fields, true, "($1 = '' OR patient_id = $1) AND ($2 = '' OR first_name ILIKE '%' || $2 || '%') AND ($3 = '' OR last_name ILIKE '%' || $3 || '%')", []interface{}{id, first_name, last_name}, false, "", "", "ORDER BY patient_id DESC")

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("Patient not found")
	}

	var patientResponses []patients.GetPatientResponse

	for _, row := range results {
		patient_id := row["patient_id"].(string)
		patient := patients.GeneralPatientInformation{
			Patient_id:        patient_id,
			First_name:        row["first_name"].(string),
			Last_name:         row["last_name"].(string),
			Age:               int(row["age"].(int64)),
			Date_of_birth:     row["date_of_birth"].(time.Time).Format("02-01-2006"),
			Gender:            string(row["gender"].([]uint8)),
			Blood_type:        string(row["blood_type"].([]uint8)),
			Email:             row["email"].(string),
			Health_insurance:  string(row["health_insurance"].([]uint8)),
			Address:           row["address"].(string),
			Phone_number:      row["phone_number"].(string),
			Id_card_number:    row["id_card_number"].(string),
			Ongoing_treatment: row["ongoing_treatment"].(string),
			Unhealthy_habits:  row["unhealthy_habits"].(string),
		}

		// Medical History
		table = "Medical_history"
		fields = []string{"*"}
		whereCondition := "patient_id = $1"
		args := []interface{}{patient_id}
		medicalResults, err := SelectData(table, fields, true, whereCondition, args, false, "", "", "")
		if err != nil {
			return nil, err
		}
		var medical_history []patients.MedicalHistory
		for _, row := range medicalResults {
			details := row["detail"].(string)
			date := row["date"].(time.Time).Format("02-01-2006")
			time := row["time"].(time.Time).Format("15:04:05")

			medical_history = append(medical_history, patients.MedicalHistory{
				Details: details,
				Date:    date,
				Time:    time,
			})
		}

		// Chronic diseases (With JOIN)
		table = "patient_chronic_disease"
		jointables := "disease ON patient_chronic_disease.disease_id = disease.disease_id"
		fields = []string{"disease_name"}
		whereCondition = "patient_id = $1"
		args = []interface{}{patient_id}
		chronicResults, err := SelectData(table, fields, true, whereCondition, args, true, jointables, "", "")
		if err != nil {
			return nil, err
		}
		var chronicDiseases []patients.ChronicDiseaseName
		for _, row := range chronicResults {
			chronicDiseases = append(chronicDiseases, patients.ChronicDiseaseName{
				DiseaseID: row["disease_name"].(string),
			})
		}

		// Drug allergies
		table = "patient_drug_allergy"
		jointables = "drug ON patient_drug_allergy.drug_id = drug.drug_id"
		whereCondition = "patient_id = $1"
		fields = []string{"drug_name"}
		args = []interface{}{patient_id}
		allergyResults, err := SelectData(table,
			fields,
			true,
			whereCondition,
			args,
			true,
			jointables,
			"",
			"")
		if err != nil {
			return nil, err
		}
		var drugAllergies []patients.DrugAllergyName
		for _, row := range allergyResults {
			drugAllergies = append(drugAllergies, patients.DrugAllergyName{
				DrugID: row["drug_name"].(string),
			})
		}

		// Patient_appointment (Select only 1 latest appointment)
		table = "patient_appointment"
		fields = []string{"*"}
		args = []interface{}{patient_id}
		appointmentResults, err := SelectData(
			table,
			fields,
			true,
			"patient_id = $1",
			args,
			false,
			"",
			"",
			"ORDER BY date DESC, time DESC LIMIT 1")

		if err != nil {
			return nil, err
		}

		var patient_appointment patients.PatientAppointment

		if len(appointmentResults) == 0 {
			patient_appointment = patients.PatientAppointment{}
		} else {
			patient_appointment = patients.PatientAppointment{
				Time:  appointmentResults[0]["time"].(time.Time).Format("15:04:05"),
				Date:  appointmentResults[0]["date"].(time.Time).Format("02-01-2006"),
				Topic: appointmentResults[0]["topic"].(string),
			}
		}

		// รวมร่าง json response = patient_model + medical_history + Patient_appointment + patient_chronicdisease + patientdrug_allerygy
		response := patients.GetPatientResponse{
			PatientGeneralInfo:    patient,
			PatientAppointment:    patient_appointment,
			PatientMedicalHistory: medical_history,
			PatientChronicDisease: chronicDiseases,
			PatientDrugAllergy:    drugAllergies,
		}
		patientResponses = append(patientResponses, response)
	}

	return patientResponses, nil
}

func AddPatientAppointment(req patients.AddPatientAppointment) error {
	// log ข้อมูลที่รับเข้ามา
	fmt.Printf("Received AddPatientRequest: %+v\n", req)

	patientMap := map[string]interface{}{
		"patient_id": req.Patient_id,
		"time":       req.Time,
		"date":       req.Date,
		"topic":      req.Topic,
	}

	fmt.Printf("Inserting patient: %+v\n", patientMap)

	// Insert to patient table
	table := "patient_appointment"
	_, err := InsertData(table, patientMap)
	if err != nil {
		return fmt.Errorf("insert patient failed: %w", err)
	}
	return nil
}
func AddPatientHistory(req patients.AddPatientHistory) error {
	// log ข้อมูลที่รับเข้ามา
	fmt.Printf("Received AddPatientRequest: %+v\n", req)

	patientMap := map[string]interface{}{
		"patient_id": req.Patient_id,
		"detail":     req.Detail,
		"time":       req.Time,
		"date":       req.Date,
	}

	fmt.Printf("Inserting patient: %+v\n", patientMap)

	// Insert to patient table
	table := "Medical_history"
	_, err := InsertData(table, patientMap)
	if err != nil {
		return fmt.Errorf("insert patient failed: %w", err)
	}
	return nil
}

func DeleteByPatientID(table string, patientID string) error {
	condition := "patient_id = $1"
	conditionValues := []interface{}{patientID}
	rowsAffected, err := DeleteData(table, condition, conditionValues)
	if err != nil {
		return fmt.Errorf("failed to delete from %s: %w", table, err)
	}
	fmt.Printf("Deleted %d rows from %s where patient_id = %s\n", rowsAffected, table, patientID)
	return nil
}

func UpdatePatient(req *patients.AddPatientRequest) (int64, error) {
	patientID := req.Patient.Patient_id
	if patientID == "" {
		return 0, fmt.Errorf("missing patient_id")
	}

	// เตรียมข้อมูลที่จะ update
	data := make(map[string]interface{})
	addIfNotEmpty := func(key, value string) {
		if value != "" {
			data[key] = value
		}
	}

	addIfNotEmpty("first_name", req.Patient.First_name)
	addIfNotEmpty("last_name", req.Patient.Last_name)
	addIfNotEmpty("health_insurance", req.Patient.Health_insurance)
	addIfNotEmpty("date_of_birth", req.Patient.Date_of_birth)
	addIfNotEmpty("gender", req.Patient.Gender)
	addIfNotEmpty("blood_type", req.Patient.Blood_type)
	addIfNotEmpty("email", req.Patient.Email)
	addIfNotEmpty("address", req.Patient.Address)
	addIfNotEmpty("phone_number", req.Patient.Phone_number)
	addIfNotEmpty("id_card_number", req.Patient.Id_card_number)
	addIfNotEmpty("ongoing_treatment", req.Patient.Ongoing_treatment)
	addIfNotEmpty("unhealthy_habits", req.Patient.Unhealthy_habits)

	// เช็กค่า 'age' ว่ามีการส่งมาหรือไม่
	if req.Patient.Age != 0 { // ใช้ค่า default 0 เช็กว่ามีการส่งมาหรือไม่
		data["age"] = fmt.Sprintf("%v", req.Patient.Age)
	}

	table := "Patient"
	condition := "patient_id = $1"
	conditionValues := []interface{}{patientID}

	var totalRowsAffected int64 = 0

	// อัปเดตข้อมูล Patient
	if len(data) > 0 {
		rowsAffected, err := UpdateData(table, data, condition, conditionValues)
		if err != nil {
			return 0, err
		}
		totalRowsAffected += rowsAffected
	}

	// ============ Chronic Diseases ============
	if len(req.PatientChronicDisease) > 0 {
		table := "patient_chronic_disease"
		// ลบของเก่า
		if err := DeleteByPatientID(table, patientID); err != nil {
			return totalRowsAffected, fmt.Errorf("failed to delete chronic diseases: %v", err)
		}

		// Insert ใหม่
		for _, chronic := range req.PatientChronicDisease {
			chronicMap := map[string]interface{}{
				"patient_id": patientID,
				"disease_id": chronic.DiseaseID,
			}
			_, err := InsertData(table, chronicMap)
			if err != nil {
				return totalRowsAffected, fmt.Errorf("insert chronic disease failed: %v", err)
			}
			totalRowsAffected++ // นับเพิ่มทีละ insert
		}
	} else if len(req.PatientChronicDisease) == 0 {
		table := "patient_chronic_disease"
		// ลบของเก่า
		if err := DeleteByPatientID(table, patientID); err != nil {
			return totalRowsAffected, fmt.Errorf("failed to delete chronic diseases: %v", err)
		}
	}

	// ============ Drug allergy ============
	if len(req.PatientDrugAllergy) > 0 {
		table := "patient_drug_allergy"
		// ลบของเก่า
		if err := DeleteByPatientID(table, patientID); err != nil {
			return totalRowsAffected, fmt.Errorf("failed to delete drug allergy: %v", err)
		}

		// Insert ใหม่
		for _, drug := range req.PatientDrugAllergy {
			drugMap := map[string]interface{}{
				"patient_id": patientID,
				"drug_id":    drug.DrugID,
			}
			_, err := InsertData(table, drugMap)
			if err != nil {
				return totalRowsAffected, fmt.Errorf("insert drug allergy failed: %v", err)
			}
			totalRowsAffected++ // นับเพิ่มทีละ insert
		}
	} else if len(req.PatientDrugAllergy) == 0 {
		table := "patient_drug_allergy"
		// ลบของเก่า
		if err := DeleteByPatientID(table, patientID); err != nil {
			return totalRowsAffected, fmt.Errorf("failed to delete drug allergy: %v", err)
		}
	}
	return totalRowsAffected, nil
}

func AddPatient(req patients.AddPatientRequest) error {
	fmt.Printf("Received AddPatientRequest: %+v\n", req)

	p := req.Patient
	patientMap := map[string]interface{}{
		"patient_id":        p.Patient_id,
		"first_name":        p.First_name,
		"last_name":         p.Last_name,
		"age":               p.Age,
		"gender":            p.Gender,
		"date_of_birth":     p.Date_of_birth,
		"blood_type":        p.Blood_type,
		"email":             p.Email,
		"health_insurance":  p.Health_insurance,
		"address":           p.Address,
		"phone_number":      p.Phone_number,
		"id_card_number":    p.Id_card_number,
		"ongoing_treatment": p.Ongoing_treatment,
		"unhealthy_habits":  p.Unhealthy_habits,
	}

	fmt.Printf("Inserting patient: %+v\n", patientMap)

	// Insert to patient table
	table := "patient"
	_, err := InsertData(table, patientMap)
	if err != nil {
		return fmt.Errorf("insert patient failed: %w", err)
	}

	// Insert to chronic diseases table
	for _, chronic := range req.PatientChronicDisease {
		if !isValidString(chronic.DiseaseID) {
			continue // ข้ามถ้า disease_id ว่าง, undefined, หรือ null
		}
		chronicMap := map[string]interface{}{
			"patient_id": p.Patient_id,
			"disease_id": chronic.DiseaseID,
		}
		_, err := InsertData("patient_chronic_disease", chronicMap)
		if err != nil {
			return fmt.Errorf("insert chronic disease failed: %w", err)
		}
	}

	// Insert to drug allergies table
	for _, allergy := range req.PatientDrugAllergy {
		if !isValidString(allergy.DrugID) {
			continue // ข้ามถ้า disease_id ว่าง, undefined, หรือ null
		}
		allergyMap := map[string]interface{}{
			"patient_id": p.Patient_id,
			"drug_id":    allergy.DrugID,
		}
		_, err := InsertData("patient_drug_allergy", allergyMap)
		if err != nil {
			return fmt.Errorf("insert drug allergy failed: %w", err)
		}
	}

	return nil
}

func GetPatient(id string) (*patients.GetPatientResponse, error) {
	table := "Patient"
	fields := []string{"*"}

	result, err := SelectData(table, fields, true, "patient_id = $1", []interface{}{id}, false, "", "", "")

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("Patient not found")
	}

	patient_id := result[0]["patient_id"].(string)
	first_name := result[0]["first_name"].(string)
	last_name := result[0]["last_name"].(string)
	age := int(result[0]["age"].(int64))
	gender := string(result[0]["gender"].([]uint8))
	date_of_birth := string(result[0]["date_of_birth"].(time.Time).Format("02-01-2006"))
	blood_type := string(result[0]["blood_type"].([]uint8))
	email := result[0]["email"].(string)
	health_insurance := string(result[0]["health_insurance"].([]uint8))
	address := result[0]["address"].(string)
	phone_number := result[0]["phone_number"].(string)
	id_card_number := result[0]["id_card_number"].(string)
	ongoing_treatment := result[0]["ongoing_treatment"].(string)
	unhealthy_habits := result[0]["unhealthy_habits"].(string)

	var patient = patients.GeneralPatientInformation{
		Patient_id:        patient_id,
		First_name:        first_name,
		Last_name:         last_name,
		Age:               age,
		Gender:            gender,
		Date_of_birth:     date_of_birth,
		Blood_type:        blood_type,
		Email:             email,
		Health_insurance:  health_insurance,
		Address:           address,
		Phone_number:      phone_number,
		Id_card_number:    id_card_number,
		Ongoing_treatment: ongoing_treatment,
		Unhealthy_habits:  unhealthy_habits,
	}

	// Medical History
	table = "Medical_history"
	fields = []string{"*"}
	whereCondition := "patient_id = $1"
	args := []interface{}{patient_id}
	medicalResults, err := SelectData(table, fields, true, whereCondition, args, false, "", "", "")
	if err != nil {
		return nil, err
	}
	var medical_history []patients.MedicalHistory
	for _, row := range medicalResults {
		details := row["detail"].(string)
		date := row["date"].(time.Time).Format("02-01-2006")
		time := row["time"].(time.Time).Format("15:04:05")

		medical_history = append(medical_history, patients.MedicalHistory{
			Details: details,
			Date:    date,
			Time:    time,
		})
	}

	// Chronic diseases (With INNER JOIN)
	table = "patient_chronic_disease"
	jointables := "disease ON patient_chronic_disease.disease_id = disease.disease_id"
	fields = []string{"disease_name"}
	whereCondition = "patient_id = $1"
	args = []interface{}{patient_id}
	chronicResults, err := SelectData(table, fields, true, whereCondition, args, true, jointables, "", "")
	if err != nil {
		return nil, err
	}
	var chronicDiseases []patients.ChronicDiseaseName
	for _, row := range chronicResults {
		chronicDiseases = append(chronicDiseases, patients.ChronicDiseaseName{
			DiseaseID: row["disease_name"].(string),
		})
	}

	// Drug allergies
	table = "patient_drug_allergy"
	jointables = "drug ON patient_drug_allergy.drug_id = drug.drug_id"
	whereCondition = "patient_id = $1"
	fields = []string{"drug_name"}
	args = []interface{}{patient_id}
	allergyResults, err := SelectData(table,
		fields,
		true,
		whereCondition,
		args,
		true,
		jointables,
		"",
		"")
	if err != nil {
		return nil, err
	}
	var drugAllergies []patients.DrugAllergyName
	for _, row := range allergyResults {
		drugAllergies = append(drugAllergies, patients.DrugAllergyName{
			DrugID: row["drug_name"].(string),
		})
	}

	// Patient_appointment (Select only 1 latest appointment)
	table = "patient_appointment"
	fields = []string{"*"}
	args = []interface{}{patient_id}
	appointmentResults, err := SelectData(
		table,
		fields,
		true,
		"patient_id = $1",
		args,
		false,
		"",
		"",
		"ORDER BY date DESC, time DESC LIMIT 1")

	if err != nil {
		return nil, err
	}

	var patient_appointment patients.PatientAppointment

	if len(appointmentResults) == 0 {
		patient_appointment = patients.PatientAppointment{}
	} else {
		patient_appointment = patients.PatientAppointment{
			Time:  appointmentResults[0]["time"].(time.Time).Format("15:04:05"),
			Date:  appointmentResults[0]["date"].(time.Time).Format("02-01-2006"),
			Topic: appointmentResults[0]["topic"].(string),
		}
	}

	// รวมร่าง json response = patient_model + medical_history + patient_chronicdisease + patientdrug_allerygy
	response := patients.GetPatientResponse{
		PatientGeneralInfo:    patient,
		PatientAppointment:    patient_appointment,
		PatientMedicalHistory: medical_history,
		PatientChronicDisease: chronicDiseases,
		PatientDrugAllergy:    drugAllergies,
	}
	return &response, nil
}

func GetAllPatients() ([]patients.GetPatientResponse, error) {
	table := "patient"
	fields := []string{"*"}
	results, err := SelectData(table, fields, false, "", nil, false, "", "", "ORDER BY patient_id DESC")
	if err != nil {
		return nil, err
	}

	var patientResponses []patients.GetPatientResponse

	for _, row := range results {
		patient_id := row["patient_id"].(string)
		patient := patients.GeneralPatientInformation{
			Patient_id:        patient_id,
			First_name:        row["first_name"].(string),
			Last_name:         row["last_name"].(string),
			Age:               int(row["age"].(int64)),
			Date_of_birth:     row["date_of_birth"].(time.Time).Format("02-01-2006"),
			Gender:            string(row["gender"].([]uint8)),
			Blood_type:        string(row["blood_type"].([]uint8)),
			Email:             row["email"].(string),
			Health_insurance:  string(row["health_insurance"].([]uint8)),
			Address:           row["address"].(string),
			Phone_number:      row["phone_number"].(string),
			Id_card_number:    row["id_card_number"].(string),
			Ongoing_treatment: row["ongoing_treatment"].(string),
			Unhealthy_habits:  row["unhealthy_habits"].(string),
		}

		// Medical History
		table = "Medical_history"
		fields = []string{"*"}
		whereCondition := "patient_id = $1"
		args := []interface{}{patient_id}
		medicalResults, err := SelectData(table, fields, true, whereCondition, args, false, "", "", "")
		if err != nil {
			return nil, err
		}
		var medical_history []patients.MedicalHistory
		for _, row := range medicalResults {
			details := row["detail"].(string)
			date := row["date"].(time.Time).Format("02-01-2006")
			time := row["time"].(time.Time).Format("15:04:05")

			medical_history = append(medical_history, patients.MedicalHistory{
				Details: details,
				Date:    date,
				Time:    time,
			})
		}

		// Chronic diseases (With JOIN)
		table = "patient_chronic_disease"
		jointables := "disease ON patient_chronic_disease.disease_id = disease.disease_id"
		fields = []string{"disease_name"}
		whereCondition = "patient_id = $1"
		args = []interface{}{patient_id}
		chronicResults, err := SelectData(table, fields, true, whereCondition, args, true, jointables, "", "")
		if err != nil {
			return nil, err
		}
		var chronicDiseases []patients.ChronicDiseaseName
		for _, row := range chronicResults {
			chronicDiseases = append(chronicDiseases, patients.ChronicDiseaseName{
				DiseaseID: row["disease_name"].(string),
			})
		}

		// Drug allergies
		table = "patient_drug_allergy"
		jointables = "drug ON patient_drug_allergy.drug_id = drug.drug_id"
		whereCondition = "patient_id = $1"
		fields = []string{"drug_name"}
		args = []interface{}{patient_id}
		allergyResults, err := SelectData(table,
			fields,
			true,
			whereCondition,
			args,
			true,
			jointables,
			"",
			"")
		if err != nil {
			return nil, err
		}
		var drugAllergies []patients.DrugAllergyName
		for _, row := range allergyResults {
			drugAllergies = append(drugAllergies, patients.DrugAllergyName{
				DrugID: row["drug_name"].(string),
			})
		}

		// Patient_appointment (Select only 1 latest appointment)
		table = "patient_appointment"
		fields = []string{"*"}
		args = []interface{}{patient_id}
		appointmentResults, err := SelectData(
			table,
			fields,
			true,
			"patient_id = $1",
			args,
			false,
			"",
			"",
			"ORDER BY date DESC, time DESC LIMIT 1")

		if err != nil {
			return nil, err
		}
		
		var patient_appointment patients.PatientAppointment

		if len(appointmentResults) == 0 {
			patient_appointment = patients.PatientAppointment{}
		} else {
			patient_appointment = patients.PatientAppointment{
				Time:  appointmentResults[0]["time"].(time.Time).Format("15:04:05"),
				Date:  appointmentResults[0]["date"].(time.Time).Format("02-01-2006"),
				Topic: appointmentResults[0]["topic"].(string),
			}
		}

		// รวมร่าง json response = patient_model + medical_history + Patient_appointment + patient_chronicdisease + patientdrug_allerygy
		response := patients.GetPatientResponse{
			PatientGeneralInfo:    patient,
			PatientAppointment:    patient_appointment,
			PatientMedicalHistory: medical_history,
			PatientChronicDisease: chronicDiseases,
			PatientDrugAllergy:    drugAllergies,
		}
		patientResponses = append(patientResponses, response)
	}

	return patientResponses, nil
}
