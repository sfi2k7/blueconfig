package models

import (
	"testing"
)

func TestNewRowValue(t *testing.T) {
	v := NewValue("test")
	if v.Base() != "test" {
		t.Errorf("Expected base to be 'test', got %v", v.Base())
	}
	if v.IsNull() {
		t.Error("Expected value to not be null")
	}

	nilValue := NewValue(nil)
	if !nilValue.IsNull() {
		t.Error("Expected nil value to be null")
	}
}

func TestNewRowValueWithSchema(t *testing.T) {
	v := NewValueWithSchema("123", "int")
	if v.SchemaType() != "int" {
		t.Errorf("Expected schema type to be 'int', got %s", v.SchemaType())
	}
	if v.Base() != "123" {
		t.Errorf("Expected base to be '123', got %v", v.Base())
	}
}

func TestRowValueAsInt64(t *testing.T) {
	tests := []struct {
		name      string
		raw       any
		expected  int64
		shouldErr bool
	}{
		{"int64", int64(42), 42, false},
		{"int", int(42), 42, false},
		{"int32", int32(42), 42, false},
		{"float64", float64(42.9), 42, false},
		{"string valid", "42", 42, false},
		{"string invalid", "abc", 0, true},
		{"nil", nil, 0, true},
		{"uint", uint(42), 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.raw)
			result, err := v.AsInt64()

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestRowValueAsFloat64(t *testing.T) {
	tests := []struct {
		name      string
		raw       any
		expected  float64
		shouldErr bool
	}{
		{"float64", float64(42.5), 42.5, false},
		{"float32", float32(42.5), 42.5, false},
		{"int", int(42), 42.0, false},
		{"int64", int64(42), 42.0, false},
		{"string valid", "42.5", 42.5, false},
		{"string invalid", "abc", 0, true},
		{"nil", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.raw)
			result, err := v.AsFloat64()

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}

func TestRowValueAsString(t *testing.T) {
	tests := []struct {
		name     string
		raw      any
		expected string
	}{
		{"string", "test", "test"},
		{"int", 42, "42"},
		{"float", 42.5, "42.5"},
		{"bool", true, "true"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.raw)
			result := v.AsString()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRowValueAsBool(t *testing.T) {
	tests := []struct {
		name      string
		raw       any
		expected  bool
		shouldErr bool
	}{
		{"bool true", true, true, false},
		{"bool false", false, false, false},
		{"int non-zero", 1, true, false},
		{"int zero", 0, false, false},
		{"string true", "true", true, false},
		{"string 1", "1", true, false},
		{"string false", "false", false, false},
		{"string 0", "0", false, false},
		{"string empty", "", false, false},
		{"string invalid", "maybe", false, true},
		{"nil", nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.raw)
			result, err := v.AsBool()

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestRowValueVal(t *testing.T) {
	t.Run("no schema returns raw", func(t *testing.T) {
		v := NewValue("test")
		if v.Val() != "test" {
			t.Errorf("Expected 'test', got %v", v.Val())
		}
	})

	t.Run("schema int converts string", func(t *testing.T) {
		v := NewValueWithSchema("42", "int")
		result := v.Val()
		if result != int64(42) {
			t.Errorf("Expected int64(42), got %v (type %T)", result, result)
		}
	})

	t.Run("schema float converts string", func(t *testing.T) {
		v := NewValueWithSchema("42.5", "float")
		result := v.Val()
		if result != 42.5 {
			t.Errorf("Expected 42.5, got %v", result)
		}
	})

	t.Run("schema bool converts string", func(t *testing.T) {
		v := NewValueWithSchema("true", "bool")
		result := v.Val()
		if result != true {
			t.Errorf("Expected true, got %v", result)
		}
	})

	t.Run("schema string returns as string", func(t *testing.T) {
		v := NewValueWithSchema(42, "string")
		result := v.Val()
		if result != "42" {
			t.Errorf("Expected '42', got %v", result)
		}
	})

	t.Run("caches converted value", func(t *testing.T) {
		v := NewValueWithSchema("42", "int")
		val1 := v.Val()
		val2 := v.Val()
		if val1 != val2 {
			t.Error("Val() should return cached value")
		}
	})
}

func TestRowValueInferredType(t *testing.T) {
	tests := []struct {
		name     string
		raw      any
		expected string
	}{
		{"string", "test", "string"},
		{"int", 42, "int"},
		{"int64", int64(42), "int"},
		{"float64", 42.5, "float"},
		{"bool", true, "bool"},
		{"nil", nil, "null"},
		{"map", map[string]any{}, "map"},
		{"slice", []int{}, "array"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValue(tt.raw)
			result := v.InferredType()
			if result != tt.expected {
				t.Errorf("Expected type '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRowValueType(t *testing.T) {
	t.Run("returns schema type if set", func(t *testing.T) {
		v := NewValueWithSchema("42", "int")
		if v.Type() != "int" {
			t.Errorf("Expected 'int', got '%s'", v.Type())
		}
	})

	t.Run("returns inferred type if no schema", func(t *testing.T) {
		v := NewValue(42)
		if v.Type() != "int" {
			t.Errorf("Expected 'int', got '%s'", v.Type())
		}
	})
}

func TestRowValueSetSchemaType(t *testing.T) {
	v := NewValue("42")
	v.SetSchemaType("int")

	if v.SchemaType() != "int" {
		t.Errorf("Expected schema type 'int', got '%s'", v.SchemaType())
	}

	result := v.Val()
	if result != int64(42) {
		t.Errorf("Expected int64(42), got %v", result)
	}
}

func TestRowValueEquals(t *testing.T) {
	tests := []struct {
		name     string
		v1       *RowValue
		v2       *RowValue
		expected bool
	}{
		{
			"equal strings",
			NewValue("test"),
			NewValue("test"),
			true,
		},
		{
			"different strings",
			NewValue("test"),
			NewValue("other"),
			false,
		},
		{
			"equal ints",
			NewValue(42),
			NewValue(42),
			true,
		},
		{
			"int and float equal",
			NewValue(42),
			NewValue(42.0),
			true,
		},
		{
			"string and int equal after conversion",
			NewValueWithSchema("42", "int"),
			NewValue(42),
			true,
		},
		{
			"both nil",
			NewValue(nil),
			NewValue(nil),
			true,
		},
		{
			"one nil",
			NewValue(nil),
			NewValue(42),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Equals(tt.v2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRowValueCompare(t *testing.T) {
	tests := []struct {
		name      string
		v1        *RowValue
		v2        *RowValue
		op        string
		expected  bool
		shouldErr bool
	}{
		{
			"int greater than",
			NewValue(42),
			NewValue(10),
			">",
			true,
			false,
		},
		{
			"int less than",
			NewValue(10),
			NewValue(42),
			"<",
			true,
			false,
		},
		{
			"string greater than",
			NewValue("b"),
			NewValue("a"),
			">",
			true,
			false,
		},
		{
			"float greater or equal",
			NewValue(42.5),
			NewValue(42.5),
			">=",
			true,
			false,
		},
		{
			"equals operator",
			NewValue(42),
			NewValue(42),
			"==",
			true,
			false,
		},
		{
			"string to int comparison",
			NewValueWithSchema("42", "int"),
			NewValue(10),
			">",
			true,
			false,
		},
		{
			"invalid operator",
			NewValue(42),
			NewValue(10),
			"!=",
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.v1.Compare(tt.v2, tt.op)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestRowValueIsZero(t *testing.T) {
	tests := []struct {
		name     string
		value    *RowValue
		expected bool
	}{
		{"nil is zero", NewValue(nil), true},
		{"zero int is zero", NewValue(0), true},
		{"non-zero int is not zero", NewValue(42), false},
		{"empty string is zero", NewValue(""), true},
		{"non-empty string is not zero", NewValue("test"), false},
		{"false bool is zero", NewValue(false), true},
		{"true bool is not zero", NewValue(true), false},
		{"zero float is zero", NewValue(0.0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.IsZero()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRowValueIsNumeric(t *testing.T) {
	tests := []struct {
		name     string
		value    *RowValue
		expected bool
	}{
		{"int is numeric", NewValue(42), true},
		{"float is numeric", NewValue(42.5), true},
		{"numeric string is numeric", NewValue("42"), true},
		{"non-numeric string is not numeric", NewValue("abc"), false},
		{"nil is not numeric", NewValue(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.IsNumeric()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRowValueClone(t *testing.T) {
	original := NewValueWithSchema("42", "int")
	_ = original.Val()

	clone := original.Clone()

	if clone.Base() != original.Base() {
		t.Error("Clone should have same base value")
	}
	if clone.SchemaType() != original.SchemaType() {
		t.Error("Clone should have same schema type")
	}
	if clone.IsNull() != original.IsNull() {
		t.Error("Clone should have same null status")
	}

	clone.SetSchemaType("string")
	if clone.SchemaType() == original.SchemaType() {
		t.Error("Modifying clone should not affect original")
	}
}

func TestRowValueString(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		v := NewValue(nil)
		str := v.String()
		if str != "<nil>" {
			t.Errorf("Expected '<nil>', got '%s'", str)
		}
	})

	t.Run("value without schema", func(t *testing.T) {
		v := NewValue("test")
		str := v.String()
		if str == "" {
			t.Error("String should not be empty")
		}
	})

	t.Run("value with schema", func(t *testing.T) {
		v := NewValueWithSchema("42", "int")
		str := v.String()
		if str == "" {
			t.Error("String should not be empty")
		}
	})
}

func TestRowValueMustInt64(t *testing.T) {
	t.Run("valid conversion", func(t *testing.T) {
		v := NewValue(42)
		result := v.MustInt64()
		if result != 42 {
			t.Errorf("Expected 42, got %d", result)
		}
	})

	t.Run("invalid conversion panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but got none")
			}
		}()
		v := NewValue("abc")
		_ = v.MustInt64()
	})
}

func TestRowValueMustFloat64(t *testing.T) {
	t.Run("valid conversion", func(t *testing.T) {
		v := NewValue(42.5)
		result := v.MustFloat64()
		if result != 42.5 {
			t.Errorf("Expected 42.5, got %f", result)
		}
	})

	t.Run("invalid conversion panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic but got none")
			}
		}()
		v := NewValue("abc")
		_ = v.MustFloat64()
	})
}

func TestRowValueDatabaseScenario(t *testing.T) {
	t.Run("string from database converted to int", func(t *testing.T) {
		rawFromDB := "12345"
		v := NewValueWithSchema(rawFromDB, "int")

		if v.Base() != rawFromDB {
			t.Error("Base should preserve raw string from database")
		}

		converted := v.Val()
		if converted != int64(12345) {
			t.Errorf("Val should convert to int64, got %v (type %T)", converted, converted)
		}

		intVal, err := v.AsInt64()
		if err != nil {
			t.Errorf("AsInt64 should succeed: %v", err)
		}
		if intVal != 12345 {
			t.Errorf("Expected 12345, got %d", intVal)
		}
	})

	t.Run("string from database converted to float", func(t *testing.T) {
		rawFromDB := "123.45"
		v := NewValueWithSchema(rawFromDB, "float")

		if v.Base() != rawFromDB {
			t.Error("Base should preserve raw string from database")
		}

		converted := v.Val()
		if converted != 123.45 {
			t.Errorf("Val should convert to float64, got %v", converted)
		}
	})

	t.Run("comparison with database strings", func(t *testing.T) {
		v1 := NewValueWithSchema("100", "int")
		v2 := NewValueWithSchema("50", "int")

		result, err := v1.Compare(v2, ">")
		if err != nil {
			t.Errorf("Comparison should succeed: %v", err)
		}
		if !result {
			t.Error("100 should be greater than 50")
		}
	})
}
