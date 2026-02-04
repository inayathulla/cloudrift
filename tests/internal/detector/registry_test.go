package detector

import (
	"testing"

	"github.com/inayathulla/cloudrift/internal/detector"
	"github.com/stretchr/testify/assert"
)

// Test DriftInfo.HasDrift method
func TestDriftInfo_HasDrift(t *testing.T) {
	tests := []struct {
		name     string
		info     detector.DriftInfo
		expected bool
	}{
		{
			name:     "no drift - empty",
			info:     detector.DriftInfo{},
			expected: false,
		},
		{
			name: "no drift - initialized empty maps",
			info: detector.DriftInfo{
				Diffs:           map[string][2]interface{}{},
				ExtraAttributes: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "drift - missing resource",
			info: detector.DriftInfo{
				Missing: true,
			},
			expected: true,
		},
		{
			name: "drift - has diffs",
			info: detector.DriftInfo{
				Diffs: map[string][2]interface{}{
					"attr": {"expected", "actual"},
				},
			},
			expected: true,
		},
		{
			name: "drift - has extra attributes",
			info: detector.DriftInfo{
				ExtraAttributes: map[string]interface{}{
					"extra": "value",
				},
			},
			expected: true,
		},
		{
			name: "drift - all conditions",
			info: detector.DriftInfo{
				Missing: true,
				Diffs: map[string][2]interface{}{
					"attr": {"a", "b"},
				},
				ExtraAttributes: map[string]interface{}{
					"extra": "value",
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.info.HasDrift())
		})
	}
}

// Test Registry functionality
func TestRegistry_NewRegistry(t *testing.T) {
	reg := detector.NewRegistry()
	assert.NotNil(t, reg)
	assert.Empty(t, reg.List())
}

func TestRegistry_Register(t *testing.T) {
	reg := detector.NewRegistry()

	// Create a mock factory
	mockFactory := func() detector.Detector {
		return nil // Mock detector
	}

	reg.Register("test-service", mockFactory)

	assert.True(t, reg.Has("test-service"))
	assert.False(t, reg.Has("unknown-service"))
}

func TestRegistry_Get(t *testing.T) {
	reg := detector.NewRegistry()

	// Register a mock detector
	called := false
	mockFactory := func() detector.Detector {
		called = true
		return nil
	}

	reg.Register("mock", mockFactory)

	// Get should call the factory
	_, err := reg.Get("mock")
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestRegistry_Get_NotFound(t *testing.T) {
	reg := detector.NewRegistry()

	_, err := reg.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no detector registered")
}

func TestRegistry_List(t *testing.T) {
	reg := detector.NewRegistry()

	reg.Register("s3", func() detector.Detector { return nil })
	reg.Register("ec2", func() detector.Detector { return nil })
	reg.Register("iam", func() detector.Detector { return nil })

	services := reg.List()
	assert.Len(t, services, 3)
	assert.Contains(t, services, "s3")
	assert.Contains(t, services, "ec2")
	assert.Contains(t, services, "iam")
}

func TestRegistry_Has(t *testing.T) {
	reg := detector.NewRegistry()

	reg.Register("exists", func() detector.Detector { return nil })

	assert.True(t, reg.Has("exists"))
	assert.False(t, reg.Has("does-not-exist"))
}

// Test default registry functions
func TestDefaultRegistry_Register(t *testing.T) {
	// Note: This modifies the global default registry
	// In real tests, you might want to skip or isolate this

	// Just verify the functions don't panic
	assert.NotPanics(t, func() {
		detector.Has("some-service")
	})

	assert.NotPanics(t, func() {
		detector.List()
	})
}

// Test concurrent access (basic)
func TestRegistry_ConcurrentAccess(t *testing.T) {
	reg := detector.NewRegistry()

	// Register in one goroutine
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			reg.Register("service", func() detector.Detector { return nil })
		}
		done <- true
	}()

	// Read in another goroutine
	go func() {
		for i := 0; i < 100; i++ {
			reg.List()
			reg.Has("service")
		}
		done <- true
	}()

	<-done
	<-done
}
