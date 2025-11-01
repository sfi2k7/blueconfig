package models

import (
	"testing"
)

func TestFlattenMapSimple(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"age":  30,
		"city": "NYC",
	}

	row := NewRow(data)

	if row["name"].AsString() != "John" {
		t.Errorf("Expected 'John', got '%s'", row["name"].AsString())
	}

	age, _ := row["age"].AsInt64()
	if age != 30 {
		t.Errorf("Expected 30, got %d", age)
	}

	if row["city"].AsString() != "NYC" {
		t.Errorf("Expected 'NYC', got '%s'", row["city"].AsString())
	}
}

func TestFlattenMapNested(t *testing.T) {
	data := map[string]any{
		"name": "Alice",
		"address": map[string]any{
			"city":   "Boston",
			"street": "Main St",
			"zip":    "02101",
		},
	}

	row := NewRow(data)

	// Should have flattened keys
	if row["name"].AsString() != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", row["name"].AsString())
	}

	if row["address.city"].AsString() != "Boston" {
		t.Errorf("Expected 'Boston', got '%s'", row["address.city"].AsString())
	}

	if row["address.street"].AsString() != "Main St" {
		t.Errorf("Expected 'Main St', got '%s'", row["address.street"].AsString())
	}

	if row["address.zip"].AsString() != "02101" {
		t.Errorf("Expected '02101', got '%s'", row["address.zip"].AsString())
	}

	// Original nested key should not exist
	if _, exists := row["address"]; exists {
		t.Error("Nested 'address' key should not exist in flattened row")
	}
}

func TestFlattenMapDeeplyNested(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"profile": map[string]any{
				"personal": map[string]any{
					"name": "Bob",
					"age":  25,
				},
			},
		},
	}

	row := NewRow(data)

	if row["user.profile.personal.name"].AsString() != "Bob" {
		t.Errorf("Expected 'Bob', got '%s'", row["user.profile.personal.name"].AsString())
	}

	age, _ := row["user.profile.personal.age"].AsInt64()
	if age != 25 {
		t.Errorf("Expected 25, got %d", age)
	}
}

func TestFlattenMapWithArray(t *testing.T) {
	data := map[string]any{
		"name": "Charlie",
		"tags": []string{"developer", "golang", "database"},
	}

	row := NewRow(data)

	if row["name"].AsString() != "Charlie" {
		t.Errorf("Expected 'Charlie', got '%s'", row["name"].AsString())
	}

	// Array should be JSON-encoded
	expected := `["developer","golang","database"]`
	if row["tags"].AsString() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, row["tags"].AsString())
	}
}

func TestFlattenMapWithNil(t *testing.T) {
	data := map[string]any{
		"name":  "Dave",
		"email": nil,
	}

	row := NewRow(data)

	if row["name"].AsString() != "Dave" {
		t.Errorf("Expected 'Dave', got '%s'", row["name"].AsString())
	}

	if !row["email"].IsNull() {
		t.Error("Expected email to be null")
	}
}

func TestFlattenStructSimple(t *testing.T) {
	type Person struct {
		Name string
		Age  int
		City string
	}

	person := Person{
		Name: "Emma",
		Age:  28,
		City: "Seattle",
	}

	row := NewRowFromStruct(person)

	if row["Name"].AsString() != "Emma" {
		t.Errorf("Expected 'Emma', got '%s'", row["Name"].AsString())
	}

	age, _ := row["Age"].AsInt64()
	if age != 28 {
		t.Errorf("Expected 28, got %d", age)
	}

	if row["City"].AsString() != "Seattle" {
		t.Errorf("Expected 'Seattle', got '%s'", row["City"].AsString())
	}
}

func TestFlattenStructWithJSONTags(t *testing.T) {
	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	person := Person{
		Name:  "Frank",
		Age:   35,
		Email: "frank@example.com",
	}

	row := NewRowFromStruct(person)

	// Should use JSON tag names
	if row["name"].AsString() != "Frank" {
		t.Errorf("Expected 'Frank', got '%s'", row["name"].AsString())
	}

	age, _ := row["age"].AsInt64()
	if age != 35 {
		t.Errorf("Expected 35, got %d", age)
	}

	if row["email"].AsString() != "frank@example.com" {
		t.Errorf("Expected 'frank@example.com', got '%s'", row["email"].AsString())
	}

	// Original field names should not exist
	if _, exists := row["Name"]; exists {
		t.Error("Field 'Name' should not exist, should use JSON tag 'name'")
	}
}

func TestFlattenStructWithSkipTag(t *testing.T) {
	type User struct {
		Name     string `json:"name"`
		Password string `json:"-"`
		Email    string `json:"email"`
	}

	user := User{
		Name:     "Grace",
		Password: "secret123",
		Email:    "grace@example.com",
	}

	row := NewRowFromStruct(user)

	// Password should be skipped
	if _, exists := row["Password"]; exists {
		t.Error("Password field should be skipped")
	}

	if _, exists := row["password"]; exists {
		t.Error("Password field should be skipped")
	}

	if row["name"].AsString() != "Grace" {
		t.Errorf("Expected 'Grace', got '%s'", row["name"].AsString())
	}

	if row["email"].AsString() != "grace@example.com" {
		t.Errorf("Expected 'grace@example.com', got '%s'", row["email"].AsString())
	}
}

func TestFlattenStructWithOmitEmpty(t *testing.T) {
	type Profile struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Age         int    `json:"age"`
		Title       string `json:"title,omitempty"`
	}

	profile := Profile{
		Name:        "Henry",
		Description: "", // Empty, should be omitted
		Age:         40,
		Title:       "", // Empty, should be omitted
	}

	row := NewRowFromStruct(profile)

	if row["name"].AsString() != "Henry" {
		t.Errorf("Expected 'Henry', got '%s'", row["name"].AsString())
	}

	age, _ := row["age"].AsInt64()
	if age != 40 {
		t.Errorf("Expected 40, got %d", age)
	}

	// Empty fields with omitempty should be skipped
	if _, exists := row["description"]; exists {
		t.Error("Empty description with omitempty should be skipped")
	}

	if _, exists := row["title"]; exists {
		t.Error("Empty title with omitempty should be skipped")
	}
}

func TestFlattenStructNested(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
		Zip    string `json:"zip"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	person := Person{
		Name: "Ivy",
		Age:  32,
		Address: Address{
			Street: "Oak Ave",
			City:   "Portland",
			Zip:    "97201",
		},
	}

	row := NewRowFromStruct(person)

	if row["name"].AsString() != "Ivy" {
		t.Errorf("Expected 'Ivy', got '%s'", row["name"].AsString())
	}

	// Nested struct should be flattened with dot notation
	if row["address.street"].AsString() != "Oak Ave" {
		t.Errorf("Expected 'Oak Ave', got '%s'", row["address.street"].AsString())
	}

	if row["address.city"].AsString() != "Portland" {
		t.Errorf("Expected 'Portland', got '%s'", row["address.city"].AsString())
	}

	if row["address.zip"].AsString() != "97201" {
		t.Errorf("Expected '97201', got '%s'", row["address.zip"].AsString())
	}
}

func TestFlattenStructDeeplyNested(t *testing.T) {
	type Location struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}

	type Address struct {
		Street   string   `json:"street"`
		City     string   `json:"city"`
		Location Location `json:"location"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	person := Person{
		Name: "Jack",
		Address: Address{
			Street: "Pine St",
			City:   "Denver",
			Location: Location{
				Lat: 39.7392,
				Lon: -104.9903,
			},
		},
	}

	row := NewRowFromStruct(person)

	if row["name"].AsString() != "Jack" {
		t.Errorf("Expected 'Jack', got '%s'", row["name"].AsString())
	}

	if row["address.street"].AsString() != "Pine St" {
		t.Errorf("Expected 'Pine St', got '%s'", row["address.street"].AsString())
	}

	if row["address.city"].AsString() != "Denver" {
		t.Errorf("Expected 'Denver', got '%s'", row["address.city"].AsString())
	}

	lat, _ := row["address.location.lat"].AsFloat64()
	if lat != 39.7392 {
		t.Errorf("Expected 39.7392, got %f", lat)
	}

	lon, _ := row["address.location.lon"].AsFloat64()
	if lon != -104.9903 {
		t.Errorf("Expected -104.9903, got %f", lon)
	}
}

func TestFlattenStructWithPointer(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}

	type Person struct {
		Name    string   `json:"name"`
		Address *Address `json:"address,omitempty"`
	}

	// Test with non-nil pointer
	person1 := Person{
		Name: "Kate",
		Address: &Address{
			City: "Austin",
		},
	}

	row1 := NewRowFromStruct(person1)

	if row1["name"].AsString() != "Kate" {
		t.Errorf("Expected 'Kate', got '%s'", row1["name"].AsString())
	}

	if row1["address.city"].AsString() != "Austin" {
		t.Errorf("Expected 'Austin', got '%s'", row1["address.city"].AsString())
	}

	// Test with nil pointer
	person2 := Person{
		Name:    "Leo",
		Address: nil,
	}

	row2 := NewRowFromStruct(person2)

	if row2["name"].AsString() != "Leo" {
		t.Errorf("Expected 'Leo', got '%s'", row2["name"].AsString())
	}

	// Nil pointer with omitempty should be skipped
	if _, exists := row2["address"]; exists {
		t.Error("Nil pointer with omitempty should be skipped")
	}
}

func TestFlattenStructWithSlice(t *testing.T) {
	type Person struct {
		Name  string   `json:"name"`
		Tags  []string `json:"tags"`
		Roles []string `json:"roles,omitempty"`
	}

	person := Person{
		Name:  "Mia",
		Tags:  []string{"admin", "developer"},
		Roles: []string{}, // Empty slice with omitempty
	}

	row := NewRowFromStruct(person)

	if row["name"].AsString() != "Mia" {
		t.Errorf("Expected 'Mia', got '%s'", row["name"].AsString())
	}

	// Slice should be JSON-encoded
	expected := `["admin","developer"]`
	if row["tags"].AsString() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, row["tags"].AsString())
	}

	// Empty slice with omitempty should be skipped
	if _, exists := row["roles"]; exists {
		t.Error("Empty slice with omitempty should be skipped")
	}
}

func TestNewObjectFromMap(t *testing.T) {
	data := map[string]any{
		"name": "Nina",
		"contact": map[string]any{
			"email": "nina@example.com",
			"phone": "555-1234",
		},
	}

	obj := NewObjectFromMap(data)

	name, ok := obj.Get("name")
	if !ok || name != "Nina" {
		t.Errorf("Expected 'Nina', got '%v'", name)
	}

	email, ok := obj.Get("contact.email")
	if !ok || email != "nina@example.com" {
		t.Errorf("Expected 'nina@example.com', got '%v'", email)
	}

	phone, ok := obj.Get("contact.phone")
	if !ok || phone != "555-1234" {
		t.Errorf("Expected '555-1234', got '%v'", phone)
	}
}

func TestNewObjectFromStruct(t *testing.T) {
	type Contact struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Contact Contact `json:"contact"`
	}

	person := Person{
		Name: "Oscar",
		Age:  45,
		Contact: Contact{
			Email: "oscar@example.com",
			Phone: "555-5678",
		},
	}

	obj := NewObjectFromStruct(person)

	name, ok := obj.Get("name")
	if !ok || name != "Oscar" {
		t.Errorf("Expected 'Oscar', got '%v'", name)
	}

	age, ok := obj.Get("age")
	if !ok {
		t.Error("Expected to get age")
	}
	// Age could be int or int64 depending on platform
	var ageInt int64
	switch v := age.(type) {
	case int:
		ageInt = int64(v)
	case int64:
		ageInt = v
	}
	if ageInt != 45 {
		t.Errorf("Expected 45, got %d", ageInt)
	}

	email, ok := obj.Get("contact.email")
	if !ok || email != "oscar@example.com" {
		t.Errorf("Expected 'oscar@example.com', got '%v'", email)
	}

	phone, ok := obj.Get("contact.phone")
	if !ok || phone != "555-5678" {
		t.Errorf("Expected '555-5678', got '%v'", phone)
	}
}

func TestSetDataFromMap(t *testing.T) {
	obj := NewObject(make(Row))

	data := map[string]any{
		"name": "Paul",
		"info": map[string]any{
			"age":  50,
			"city": "Miami",
		},
	}

	obj.SetDataFromMap(data)

	name, ok := obj.Get("name")
	if !ok || name != "Paul" {
		t.Errorf("Expected 'Paul', got '%v'", name)
	}

	age, ok := obj.Get("info.age")
	if !ok {
		t.Error("Expected to get info.age")
	}
	// Age could be int or int64 depending on platform
	var ageInt int64
	switch v := age.(type) {
	case int:
		ageInt = int64(v)
	case int64:
		ageInt = v
	}
	if ageInt != 50 {
		t.Errorf("Expected 50, got %d", ageInt)
	}

	city, ok := obj.Get("info.city")
	if !ok || city != "Miami" {
		t.Errorf("Expected 'Miami', got '%v'", city)
	}
}

func TestSetDataFromStruct(t *testing.T) {
	type Info struct {
		Age  int    `json:"age"`
		City string `json:"city"`
	}

	type Person struct {
		Name string `json:"name"`
		Info Info   `json:"info"`
	}

	obj := NewObject(make(Row))

	person := Person{
		Name: "Quinn",
		Info: Info{
			Age:  55,
			City: "Phoenix",
		},
	}

	obj.SetDataFromStruct(person)

	name, ok := obj.Get("name")
	if !ok || name != "Quinn" {
		t.Errorf("Expected 'Quinn', got '%v'", name)
	}

	age, ok := obj.Get("info.age")
	if !ok {
		t.Error("Expected to get info.age")
	}
	// Age could be int or int64 depending on platform
	var ageInt int64
	switch v := age.(type) {
	case int:
		ageInt = int64(v)
	case int64:
		ageInt = v
	}
	if ageInt != 55 {
		t.Errorf("Expected 55, got %d", ageInt)
	}

	city, ok := obj.Get("info.city")
	if !ok || city != "Phoenix" {
		t.Errorf("Expected 'Phoenix', got '%v'", city)
	}
}

func TestFlattenMixedStructAndMap(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}

	type Person struct {
		Name    string         `json:"name"`
		Address Address        `json:"address"`
		Meta    map[string]any `json:"meta"`
	}

	person := Person{
		Name: "Rita",
		Address: Address{
			City: "Chicago",
		},
		Meta: map[string]any{
			"verified": true,
			"score":    95,
		},
	}

	row := NewRowFromStruct(person)

	if row["name"].AsString() != "Rita" {
		t.Errorf("Expected 'Rita', got '%s'", row["name"].AsString())
	}

	if row["address.city"].AsString() != "Chicago" {
		t.Errorf("Expected 'Chicago', got '%s'", row["address.city"].AsString())
	}

	verified, _ := row["meta.verified"].AsBool()
	if !verified {
		t.Error("Expected meta.verified to be true")
	}

	score, _ := row["meta.score"].AsInt64()
	if score != 95 {
		t.Errorf("Expected 95, got %d", score)
	}
}
