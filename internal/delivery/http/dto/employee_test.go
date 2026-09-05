package dto

import "testing"

func TestCreateEmployeeRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateEmployeeRequest
		expected bool
	}{
		{
			name: "valid employee",
			request: CreateEmployeeRequest{
				Name:       "Ali",
				Email:      "ali@example.com",
				Department: stringPtr("IT"),
			},
			expected: true,
		},
		{
			name: "empty name",
			request: CreateEmployeeRequest{
				Name:  "",
				Email: "ali@example.com",
			},
			expected: false,
		},
		{
			name: "empty email",
			request: CreateEmployeeRequest{
				Name:  "Ali",
				Email: "",
			},
			expected: false,
		},
		{
			name: "name contains only spaces",
			request: CreateEmployeeRequest{
				Name:  "   ",
				Email: "ali@example.com",
			},
			expected: false,
		},
		{
			name: "email contains only spaces",
			request: CreateEmployeeRequest{
				Name:  "Ali",
				Email: "   ",
			},
			expected: false,
		},
		{
			name: "department is optional",
			request: CreateEmployeeRequest{
				Name:  "Ali",
				Email: "ali@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()

			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}

func TestUpdateEmployeeRequest_Validate(t *testing.T) {
	tests := []struct {
		name     string
		request  UpdateEmployeeRequest
		expected bool
	}{
		{
			name: "valid employee",
			request: UpdateEmployeeRequest{
				Name:       "Ali",
				Email:      "ali@example.com",
				Department: stringPtr("IT"),
			},
			expected: true,
		},
		{
			name: "empty name",
			request: UpdateEmployeeRequest{
				Name:  "",
				Email: "ali@example.com",
			},
			expected: false,
		},
		{
			name: "empty email",
			request: UpdateEmployeeRequest{
				Name:  "Ali",
				Email: "",
			},
			expected: false,
		},
		{
			name: "name contains only spaces",
			request: UpdateEmployeeRequest{
				Name:  "   ",
				Email: "ali@example.com",
			},
			expected: false,
		},
		{
			name: "email contains only spaces",
			request: UpdateEmployeeRequest{
				Name:  "Ali",
				Email: "   ",
			},
			expected: false,
		},
		{
			name: "department is optional",
			request: UpdateEmployeeRequest{
				Name:  "Ali",
				Email: "ali@example.com",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Validate()

			if result != tt.expected {
				t.Fatalf(
					"expected %v, got %v",
					tt.expected,
					result,
				)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func BenchmarkCreateEmployeeRequestValidate(b *testing.B) {
	req := CreateEmployeeRequest{
		Name:  "Ali",
		Email: "ali@example.com",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req.Validate()
	}
}