package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// Engine evaluates resources against OPA policies.
type Engine struct {
	modules     map[string]*ast.Module
	compiler    *ast.Compiler
	policyPaths []string
}

// NewEngine creates a new policy engine by loading policies from the given paths.
func NewEngine(policyPaths ...string) (*Engine, error) {
	e := &Engine{
		modules:     make(map[string]*ast.Module),
		policyPaths: policyPaths,
	}

	// Load all policies
	for _, path := range policyPaths {
		if err := e.loadPolicies(path); err != nil {
			return nil, fmt.Errorf("failed to load policies from %s: %w", path, err)
		}
	}

	// Compile all modules
	if len(e.modules) > 0 {
		e.compiler = ast.NewCompiler()
		if e.compiler.Compile(e.modules); e.compiler.Failed() {
			return nil, fmt.Errorf("failed to compile policies: %v", e.compiler.Errors)
		}
	}

	return e, nil
}

// loadPolicies loads all .rego files from a path (file or directory).
func (e *Engine) loadPolicies(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(p, ".rego") {
				return e.loadPolicyFile(p)
			}
			return nil
		})
	}

	return e.loadPolicyFile(path)
}

// loadPolicyFile loads a single .rego file.
func (e *Engine) loadPolicyFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read policy file: %w", err)
	}

	module, err := ast.ParseModule(path, string(content))
	if err != nil {
		return fmt.Errorf("failed to parse policy %s: %w", path, err)
	}

	e.modules[path] = module
	return nil
}

// Evaluate runs all loaded policies against the given input.
func (e *Engine) Evaluate(ctx context.Context, input *PolicyInput) (*EvaluationResult, error) {
	if len(e.modules) == 0 {
		return &EvaluationResult{}, nil
	}

	result := &EvaluationResult{
		Violations: make([]Violation, 0),
		Warnings:   make([]Violation, 0),
	}

	// Convert input to map for OPA
	inputData, err := toMap(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input: %w", err)
	}

	// Query for deny rules
	denyViolations, err := e.queryDeny(ctx, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to query deny rules: %w", err)
	}
	result.Violations = append(result.Violations, denyViolations...)
	result.Failed = len(denyViolations)

	// Query for warn rules
	warnViolations, err := e.queryWarn(ctx, inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to query warn rules: %w", err)
	}
	result.Warnings = append(result.Warnings, warnViolations...)

	return result, nil
}

// EvaluateAll runs policies against multiple inputs.
func (e *Engine) EvaluateAll(ctx context.Context, inputs []*PolicyInput) (*EvaluationResult, error) {
	combined := &EvaluationResult{
		Violations: make([]Violation, 0),
		Warnings:   make([]Violation, 0),
	}

	for _, input := range inputs {
		result, err := e.Evaluate(ctx, input)
		if err != nil {
			return nil, err
		}
		combined.Violations = append(combined.Violations, result.Violations...)
		combined.Warnings = append(combined.Warnings, result.Warnings...)
		combined.Failed += result.Failed
		combined.Passed += result.Passed
	}

	return combined, nil
}

// queryDeny queries for "deny" rule violations.
func (e *Engine) queryDeny(ctx context.Context, input map[string]interface{}) ([]Violation, error) {
	return e.queryRule(ctx, input, "deny", SeverityHigh)
}

// queryWarn queries for "warn" rule violations.
func (e *Engine) queryWarn(ctx context.Context, input map[string]interface{}) ([]Violation, error) {
	return e.queryRule(ctx, input, "warn", SeverityMedium)
}

// queryRule queries a specific rule type across all packages.
func (e *Engine) queryRule(ctx context.Context, input map[string]interface{}, ruleName string, defaultSeverity Severity) ([]Violation, error) {
	violations := make([]Violation, 0)

	// Get all unique packages
	packages := make(map[string]bool)
	for _, module := range e.modules {
		// Package path already includes "data." prefix in OPA
		// We need to get the path without it for the query
		pkg := module.Package.Path.String()
		// Remove "data." prefix if present
		pkg = strings.TrimPrefix(pkg, "data.")
		packages[pkg] = true
	}

	// Query each package for the rule
	for pkg := range packages {
		query := fmt.Sprintf("data.%s.%s", pkg, ruleName)

		r := rego.New(
			rego.Query(query),
			rego.Compiler(e.compiler),
			rego.Input(input),
		)

		rs, err := r.Eval(ctx)
		if err != nil {
			continue // Rule doesn't exist in this package
		}

		for _, result := range rs {
			for _, expr := range result.Expressions {
				// Handle different return types from OPA
				switch v := expr.Value.(type) {
				case []interface{}:
					// Array of violations
					for _, item := range v {
						violation := e.parseViolation(item, pkg, defaultSeverity, input)
						if violation != nil {
							violations = append(violations, *violation)
						}
					}
				case map[string]interface{}:
					// Single violation returned as object
					violation := e.parseViolation(v, pkg, defaultSeverity, input)
					if violation != nil {
						violations = append(violations, *violation)
					}
				case bool:
					// Skip boolean results (e.g., when deny is empty set evaluating to false)
				default:
					// For sets and other types, try to iterate
					if set, ok := v.(interface{ Iter(func(interface{}) error) error }); ok {
						set.Iter(func(item interface{}) error {
							violation := e.parseViolation(item, pkg, defaultSeverity, input)
							if violation != nil {
								violations = append(violations, *violation)
							}
							return nil
						})
					}
				}
			}
		}
	}

	return violations, nil
}

// parseViolation converts a policy result into a Violation struct.
func (e *Engine) parseViolation(item interface{}, pkg string, defaultSeverity Severity, input map[string]interface{}) *Violation {
	v := &Violation{
		PolicyID: pkg,
		Severity: defaultSeverity,
		Metadata: make(map[string]interface{}),
	}

	// Extract resource info from input
	if resource, ok := input["resource"].(map[string]interface{}); ok {
		if rt, ok := resource["type"].(string); ok {
			v.ResourceType = rt
		}
		if addr, ok := resource["address"].(string); ok {
			v.ResourceAddress = addr
		}
	}

	switch msg := item.(type) {
	case string:
		v.Message = msg
	case map[string]interface{}:
		// Rich violation object
		if m, ok := msg["msg"].(string); ok {
			v.Message = m
		} else if m, ok := msg["message"].(string); ok {
			v.Message = m
		}
		if id, ok := msg["policy_id"].(string); ok {
			v.PolicyID = id
		}
		if name, ok := msg["policy_name"].(string); ok {
			v.PolicyName = name
		}
		if sev, ok := msg["severity"].(string); ok {
			v.Severity = Severity(sev)
		}
		if rem, ok := msg["remediation"].(string); ok {
			v.Remediation = rem
		}
		if meta, ok := msg["metadata"].(map[string]interface{}); ok {
			v.Metadata = meta
		}
	default:
		return nil
	}

	if v.Message == "" {
		return nil
	}

	return v
}

// PolicyCount returns the number of loaded policies.
func (e *Engine) PolicyCount() int {
	return len(e.modules)
}

// toMap converts a struct to map[string]interface{}.
func toMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
