package scaling

import (
	"testing"

	finopsv1 "github.com/migalsp/kubex-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseMinutes(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"00:00", 0},
		{"01:30", 90},
		{"12:00", 720},
		{"23:59", 1439},
	}

	for _, tt := range tests {
		actual := parseMinutes(tt.input)
		if actual != tt.expected {
			t.Errorf("parseMinutes(%q) = %d; want %d", tt.input, actual, tt.expected)
		}
	}
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		name       string
		exclusions []string
		expected   bool
	}{
		{"frontend", []string{"backend", "redis"}, false},
		{"frontend", []string{"frontend"}, true},
		{"frontend", []string{"front*"}, true},
		{"api-server", []string{"*"}, true},
		{"db-postgres", []string{"db-*"}, true},
		{"db-postgres", []string{"db"}, false},
		{"  spaced  ", []string{"spaced"}, true},
		{"empty-rule", []string{""}, false},
	}

	for _, tt := range tests {
		actual := isExcluded(tt.name, tt.exclusions)
		if actual != tt.expected {
			t.Errorf("isExcluded(%q, %v) = %v; want %v", tt.name, tt.exclusions, actual, tt.expected)
		}
	}
}

func TestGetSequenceIndex(t *testing.T) {
	sequence := []string{"db-*", "backend", "*", "frontend"}

	tests := []struct {
		name     string
		expected int
	}{
		{"db-postgres", 0},
		{"backend", 1},
		{"anything-else", 2},
		{"frontend-app", 2},    // Matches "*" before "frontend" since "*" is at index 2
		{"unknown-no-star", 2}, // Matches "*" at index 2
	}

	for _, tt := range tests {
		obj := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: tt.name},
		}
		actual := getSequenceIndex(obj, sequence)
		if actual != tt.expected {
			t.Errorf("getSequenceIndex(%q) = %d; want %d", tt.name, actual, tt.expected)
		}
	}

	// Test missing string
	obj2 := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "not-in-list"},
	}
	actual := getSequenceIndex(obj2, []string{"only-one"})
	if actual != 999 {
		t.Errorf("getSequenceIndex(not-in-list) = %d; want 999", actual)
	}
}

func TestIsActive(t *testing.T) {
	engine := &Engine{}

	truthy := true
	falsy := false

	tests := []struct {
		name         string
		schedules    []finopsv1.ScalingSchedule
		manualActive *bool
		expected     bool
	}{
		{
			name:         "manual override true",
			schedules:    []finopsv1.ScalingSchedule{{Days: []int{0, 1, 2, 3, 4, 5, 6}, StartTime: "00:00", EndTime: "00:01"}},
			manualActive: &truthy,
			expected:     true,
		},
		{
			name:         "manual override false ignores schedule",
			schedules:    []finopsv1.ScalingSchedule{{Days: []int{0, 1, 2, 3, 4, 5, 6}, StartTime: "00:00", EndTime: "23:59"}},
			manualActive: &falsy,
			expected:     false,
		},
		{
			name:         "no schedules, no override",
			schedules:    nil,
			manualActive: nil,
			expected:     true, // defaults to active
		},
		{
			name:         "empty schedules list, no override",
			schedules:    []finopsv1.ScalingSchedule{},
			manualActive: nil,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := engine.IsActive(tt.schedules, tt.manualActive)
			if actual != tt.expected {
				t.Errorf("IsActive() = %v; want %v", actual, tt.expected)
			}
		})
	}
}
