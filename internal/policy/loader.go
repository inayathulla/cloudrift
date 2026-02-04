package policy

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// BuiltinPolicies contains the embedded built-in policy files.
//
//go:embed policies/*
var BuiltinPolicies embed.FS

// LoadBuiltinPolicies creates an Engine with the built-in policies.
func LoadBuiltinPolicies() (*Engine, error) {
	// Create a temp directory to extract embedded policies
	tmpDir, err := os.MkdirTemp("", "cloudrift-policies-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded policies to temp directory
	err = fs.WalkDir(BuiltinPolicies, "policies", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get the relative path
		relPath := strings.TrimPrefix(path, "policies/")
		if relPath == "" || relPath == "policies" {
			return nil
		}

		targetPath := filepath.Join(tmpDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Read and write file
		content, err := BuiltinPolicies.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return os.WriteFile(targetPath, content, 0644)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract built-in policies: %w", err)
	}

	return NewEngine(tmpDir)
}

// LoadPoliciesWithBuiltins creates an Engine with both built-in and custom policies.
func LoadPoliciesWithBuiltins(customPaths ...string) (*Engine, error) {
	// First extract built-in policies
	tmpDir, err := os.MkdirTemp("", "cloudrift-policies-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Extract embedded policies
	err = fs.WalkDir(BuiltinPolicies, "policies", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, "policies/")
		if relPath == "" || relPath == "policies" {
			return nil
		}

		targetPath := filepath.Join(tmpDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		content, err := BuiltinPolicies.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		return os.WriteFile(targetPath, content, 0644)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to extract built-in policies: %w", err)
	}

	// Combine built-in and custom paths
	allPaths := append([]string{tmpDir}, customPaths...)

	return NewEngine(allPaths...)
}

// ListBuiltinPolicies returns a list of built-in policy files.
func ListBuiltinPolicies() ([]string, error) {
	var policies []string

	err := fs.WalkDir(BuiltinPolicies, "policies", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".rego") {
			policies = append(policies, path)
		}
		return nil
	})

	return policies, err
}

// ListPolicyCategories returns the available policy categories.
func ListPolicyCategories() ([]string, error) {
	var categories []string
	seen := make(map[string]bool)

	entries, err := BuiltinPolicies.ReadDir("policies")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && !seen[entry.Name()] {
			categories = append(categories, entry.Name())
			seen[entry.Name()] = true
		}
	}

	return categories, nil
}
