package models

import (
	"testing"
)

func TestUnflattenMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "simple flat map",
			input: map[string]any{
				"name": "John",
				"age":  30,
			},
			expected: map[string]any{
				"name": "John",
				"age":  30,
			},
		},
		{
			name: "nested object",
			input: map[string]any{
				"name":         "John",
				"address.city": "New York",
				"address.zip":  "10001",
			},
			expected: map[string]any{
				"name": "John",
				"address": map[string]any{
					"city": "New York",
					"zip":  "10001",
				},
			},
		},
		{
			name: "deeply nested",
			input: map[string]any{
				"user.profile.name":    "Alice",
				"user.profile.age":     25,
				"user.settings.theme":  "dark",
				"user.settings.notify": true,
			},
			expected: map[string]any{
				"user": map[string]any{
					"profile": map[string]any{
						"name": "Alice",
						"age":  25,
					},
					"settings": map[string]any{
						"theme":  "dark",
						"notify": true,
					},
				},
			},
		},
		{
			name: "mixed nested and flat",
			input: map[string]any{
				"id":            123,
				"name":          "Product",
				"details.color": "blue",
				"details.size":  "large",
				"price":         99.99,
			},
			expected: map[string]any{
				"id":   123,
				"name": "Product",
				"details": map[string]any{
					"color": "blue",
					"size":  "large",
				},
				"price": 99.99,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unflattenMap(tt.input)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("unflattenMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRowToMap(t *testing.T) {
	tests := []struct {
		name     string
		input    Row
		expected map[string]any
	}{
		{
			name: "simple row",
			input: Row{
				"name": NewValue("John"),
				"age":  NewValue(30),
			},
			expected: map[string]any{
				"name": "John",
				"age":  30,
			},
		},
		{
			name: "nested row",
			input: Row{
				"name":         NewValue("John"),
				"address.city": NewValue("New York"),
				"address.zip":  NewValue("10001"),
			},
			expected: map[string]any{
				"name": "John",
				"address": map[string]any{
					"city": "New York",
					"zip":  "10001",
				},
			},
		},
		{
			name: "row with null values",
			input: Row{
				"name":  NewValue("Alice"),
				"email": NewValue(nil),
				"age":   NewValue(25),
			},
			expected: map[string]any{
				"name":  "Alice",
				"email": nil,
				"age":   25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RowToMap(tt.input)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("RowToMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRowToStruct(t *testing.T) {
	type Address struct {
		City    string `json:"city"`
		ZipCode string `json:"zip"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	t.Run("basic struct", func(t *testing.T) {
		row := Row{
			"name":         NewValue("John"),
			"age":          NewValue(30),
			"address.city": NewValue("New York"),
			"address.zip":  NewValue("10001"),
		}

		var person Person
		err := RowToStruct(row, &person)
		if err != nil {
			t.Fatalf("RowToStruct() error = %v", err)
		}

		if person.Name != "John" {
			t.Errorf("Name = %s, want John", person.Name)
		}
		if person.Age != 30 {
			t.Errorf("Age = %d, want 30", person.Age)
		}
		if person.Address.City != "New York" {
			t.Errorf("Address.City = %s, want New York", person.Address.City)
		}
		if person.Address.ZipCode != "10001" {
			t.Errorf("Address.ZipCode = %s, want 10001", person.Address.ZipCode)
		}
	})

	t.Run("struct with pointer fields", func(t *testing.T) {
		type PersonWithPointers struct {
			Name    *string  `json:"name"`
			Age     *int     `json:"age"`
			Address *Address `json:"address"`
		}

		row := Row{
			"name":         NewValue("Alice"),
			"age":          NewValue(25),
			"address.city": NewValue("Boston"),
			"address.zip":  NewValue("02101"),
		}

		var person PersonWithPointers
		err := RowToStruct(row, &person)
		if err != nil {
			t.Fatalf("RowToStruct() error = %v", err)
		}

		if person.Name == nil || *person.Name != "Alice" {
			t.Errorf("Name = %v, want Alice", person.Name)
		}
		if person.Age == nil || *person.Age != 25 {
			t.Errorf("Age = %v, want 25", person.Age)
		}
		if person.Address == nil {
			t.Fatal("Address is nil")
		}
		if person.Address.City != "Boston" {
			t.Errorf("Address.City = %s, want Boston", person.Address.City)
		}
	})

	t.Run("struct with array fields", func(t *testing.T) {
		type Product struct {
			Name  string   `json:"name"`
			Tags  []string `json:"tags"`
			Price float64  `json:"price"`
		}

		row := Row{
			"name":  NewValue("Widget"),
			"tags":  NewValue(`["electronics","gadget","popular"]`),
			"price": NewValue(99.99),
		}

		var product Product
		err := RowToStruct(row, &product)
		if err != nil {
			t.Fatalf("RowToStruct() error = %v", err)
		}

		if product.Name != "Widget" {
			t.Errorf("Name = %s, want Widget", product.Name)
		}
		if len(product.Tags) != 3 {
			t.Errorf("Tags length = %d, want 3", len(product.Tags))
		}
		if product.Price != 99.99 {
			t.Errorf("Price = %f, want 99.99", product.Price)
		}
	})

	t.Run("struct with omitempty", func(t *testing.T) {
		type User struct {
			Name     string `json:"name"`
			Email    string `json:"email,omitempty"`
			Nickname string `json:"nickname,omitempty"`
		}

		row := Row{
			"name":  NewValue("Bob"),
			"email": NewValue("bob@example.com"),
		}

		var user User
		err := RowToStruct(row, &user)
		if err != nil {
			t.Fatalf("RowToStruct() error = %v", err)
		}

		if user.Name != "Bob" {
			t.Errorf("Name = %s, want Bob", user.Name)
		}
		if user.Email != "bob@example.com" {
			t.Errorf("Email = %s, want bob@example.com", user.Email)
		}
		if user.Nickname != "" {
			t.Errorf("Nickname = %s, want empty", user.Nickname)
		}
	})
}

func TestObjectToMap(t *testing.T) {
	t.Run("simple object", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}

		obj := NewObjectFromMap(data)
		result := obj.ToMap()

		if !mapsEqual(result, data) {
			t.Errorf("ToMap() = %v, want %v", result, data)
		}
	})

	t.Run("nested object", func(t *testing.T) {
		data := map[string]any{
			"name": "Alice",
			"address": map[string]any{
				"city": "Boston",
				"zip":  "02101",
			},
		}

		obj := NewObjectFromMap(data)
		result := obj.ToMap()

		if !mapsEqual(result, data) {
			t.Errorf("ToMap() = %v, want %v", result, data)
		}
	})

	t.Run("deeply nested object", func(t *testing.T) {
		data := map[string]any{
			"user": map[string]any{
				"profile": map[string]any{
					"name": "Bob",
					"age":  40,
				},
				"settings": map[string]any{
					"theme":  "light",
					"notify": false,
				},
			},
		}

		obj := NewObjectFromMap(data)
		result := obj.ToMap()

		if !mapsEqual(result, data) {
			t.Errorf("ToMap() = %v, want %v", result, data)
		}
	})
}

func TestObjectToStruct(t *testing.T) {
	type Address struct {
		City string `json:"city"`
		Zip  string `json:"zip"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	t.Run("from map", func(t *testing.T) {
		data := map[string]any{
			"name": "Charlie",
			"age":  35,
			"address": map[string]any{
				"city": "Seattle",
				"zip":  "98101",
			},
		}

		obj := NewObjectFromMap(data)

		var person Person
		err := obj.ToStruct(&person)
		if err != nil {
			t.Fatalf("ToStruct() error = %v", err)
		}

		if person.Name != "Charlie" {
			t.Errorf("Name = %s, want Charlie", person.Name)
		}
		if person.Age != 35 {
			t.Errorf("Age = %d, want 35", person.Age)
		}
		if person.Address.City != "Seattle" {
			t.Errorf("Address.City = %s, want Seattle", person.Address.City)
		}
		if person.Address.Zip != "98101" {
			t.Errorf("Address.Zip = %s, want 98101", person.Address.Zip)
		}
	})

	t.Run("from struct", func(t *testing.T) {
		original := Person{
			Name: "Diana",
			Age:  28,
			Address: Address{
				City: "Portland",
				Zip:  "97201",
			},
		}

		obj := NewObjectFromStruct(original)

		var person Person
		err := obj.ToStruct(&person)
		if err != nil {
			t.Fatalf("ToStruct() error = %v", err)
		}

		if person.Name != original.Name {
			t.Errorf("Name = %s, want %s", person.Name, original.Name)
		}
		if person.Age != original.Age {
			t.Errorf("Age = %d, want %d", person.Age, original.Age)
		}
		if person.Address.City != original.Address.City {
			t.Errorf("Address.City = %s, want %s", person.Address.City, original.Address.City)
		}
		if person.Address.Zip != original.Address.Zip {
			t.Errorf("Address.Zip = %s, want %s", person.Address.Zip, original.Address.Zip)
		}
	})
}

func TestRoundTripConversion(t *testing.T) {
	type Profile struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	type User struct {
		ID      int     `json:"id"`
		Profile Profile `json:"profile"`
		Active  bool    `json:"active"`
	}

	t.Run("struct -> object -> struct", func(t *testing.T) {
		original := User{
			ID: 42,
			Profile: Profile{
				Username: "testuser",
				Email:    "test@example.com",
			},
			Active: true,
		}

		// Convert to Object
		obj := NewObjectFromStruct(original)

		// Convert back to struct
		var result User
		err := obj.ToStruct(&result)
		if err != nil {
			t.Fatalf("ToStruct() error = %v", err)
		}

		// Verify
		if result.ID != original.ID {
			t.Errorf("ID = %d, want %d", result.ID, original.ID)
		}
		if result.Profile.Username != original.Profile.Username {
			t.Errorf("Username = %s, want %s", result.Profile.Username, original.Profile.Username)
		}
		if result.Profile.Email != original.Profile.Email {
			t.Errorf("Email = %s, want %s", result.Profile.Email, original.Profile.Email)
		}
		if result.Active != original.Active {
			t.Errorf("Active = %v, want %v", result.Active, original.Active)
		}
	})

	t.Run("map -> object -> map", func(t *testing.T) {
		original := map[string]any{
			"id": 42,
			"profile": map[string]any{
				"username": "testuser",
				"email":    "test@example.com",
			},
			"active": true,
		}

		// Convert to Object
		obj := NewObjectFromMap(original)

		// Convert back to map
		result := obj.ToMap()

		// Verify
		if !mapsEqual(result, original) {
			t.Errorf("Round trip failed: got %v, want %v", result, original)
		}
	})
}

// Helper function to compare maps deeply
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valA := range a {
		valB, exists := b[key]
		if !exists {
			return false
		}

		// Handle nested maps
		if mapA, okA := valA.(map[string]any); okA {
			if mapB, okB := valB.(map[string]any); okB {
				if !mapsEqual(mapA, mapB) {
					return false
				}
			} else {
				return false
			}
		} else if valA != valB {
			return false
		}
	}

	return true
}
