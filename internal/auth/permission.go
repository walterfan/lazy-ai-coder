package auth

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// PermissionEngine handles permission evaluation using AWS-style policies
type PermissionEngine struct {
	policyService PolicyService
}

// NewPermissionEngine creates a new permission engine
func NewPermissionEngine(policyService PolicyService) *PermissionEngine {
	return &PermissionEngine{
		policyService: policyService,
	}
}

// CheckPermission evaluates if a user has permission to perform an action on a resource
func (p *PermissionEngine) CheckPermission(check PermissionCheck) (bool, error) {
	// Get all policies for the user (direct, role-based, and resource-based)
	policies, err := p.policyService.GetUserPolicies(check.UserID, check.RealmID)
	if err != nil {
		return false, fmt.Errorf("failed to get user policies: %w", err)
	}

	// Evaluate each policy
	for _, policy := range policies {
		allowed, err := p.evaluatePolicy(policy, check)
		if err != nil {
			return false, fmt.Errorf("failed to evaluate policy %s: %w", policy.ID, err)
		}

		// If any policy explicitly denies, return false
		if allowed != nil && *allowed == false {
			return false, nil
		}

		// If any policy explicitly allows, continue checking for explicit denies
		if allowed != nil && *allowed == true {
			continue
		}
	}

	// Check if we found any explicit allows
	hasExplicitAllow := false
	for _, policy := range policies {
		if p.hasExplicitAllow(policy, check) {
			hasExplicitAllow = true
			break
		}
	}

	// Default deny if no explicit allow found
	return hasExplicitAllow, nil
}

// evaluatePolicy evaluates a single policy against the permission check
func (p *PermissionEngine) evaluatePolicy(policy *Policy, check PermissionCheck) (*bool, error) {
	statements, err := p.policyService.GetPolicyStatements(policy.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy statements: %w", err)
	}

	for _, statement := range statements {
		// Check if action matches
		if !p.actionMatches(check.Action, statement.Actions) {
			continue
		}

		// Check if resource matches
		if !p.resourceMatches(check.Resource, statement.Resources) {
			continue
		}

		// Check conditions
		if !p.conditionsMatch(statement.Conditions, check.Context) {
			continue
		}

		// Return the effect (Allow/Deny)
		allowed := statement.Effect == "Allow"
		return &allowed, nil
	}

	// No matching statements
	return nil, nil
}

// actionMatches checks if an action matches any of the action patterns
func (p *PermissionEngine) actionMatches(action string, patterns []string) bool {
	for _, pattern := range patterns {
		if p.patternMatches(action, pattern) {
			return true
		}
	}
	return false
}

// resourceMatches checks if a resource matches any of the resource patterns
func (p *PermissionEngine) resourceMatches(resource string, patterns []string) bool {
	for _, pattern := range patterns {
		if p.patternMatches(resource, pattern) {
			return true
		}
	}
	return false
}

// patternMatches checks if a value matches a pattern (supports wildcards and variable substitution)
func (p *PermissionEngine) patternMatches(value, pattern string) bool {
	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		matched, _ := regexp.MatchString(regexPattern, value)
		return matched
	}

	// Handle variable substitution patterns (e.g., "user:${user:id}")
	if strings.Contains(pattern, "${") {
		// For now, return false - variable substitution would need context
		return false
	}

	// Exact match
	return value == pattern
}

// conditionsMatch checks if conditions are satisfied given the context
func (p *PermissionEngine) conditionsMatch(conditions map[string]interface{}, context map[string]interface{}) bool {
	if conditions == nil || len(conditions) == 0 {
		return true
	}

	for conditionType, conditionValue := range conditions {
		switch conditionType {
		case "StringEquals":
			if !p.checkStringEquals(conditionValue, context) {
				return false
			}
		case "StringNotEquals":
			if p.checkStringEquals(conditionValue, context) {
				return false
			}
		case "StringLike":
			if !p.checkStringLike(conditionValue, context) {
				return false
			}
		case "NumericEquals":
			if !p.checkNumericEquals(conditionValue, context) {
				return false
			}
		case "Bool":
			if !p.checkBool(conditionValue, context) {
				return false
			}
		// Add more condition types as needed
		default:
			// Unknown condition type, deny by default
			return false
		}
	}

	return true
}

// checkStringEquals checks StringEquals conditions
func (p *PermissionEngine) checkStringEquals(condition interface{}, context map[string]interface{}) bool {
	conditionMap, ok := condition.(map[string]interface{})
	if !ok {
		return false
	}

	for key, expectedValue := range conditionMap {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
			return false
		}
	}

	return true
}

// checkStringLike checks StringLike conditions (supports wildcards)
func (p *PermissionEngine) checkStringLike(condition interface{}, context map[string]interface{}) bool {
	conditionMap, ok := condition.(map[string]interface{})
	if !ok {
		return false
	}

	for key, expectedValue := range conditionMap {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		expectedStr := fmt.Sprintf("%v", expectedValue)
		actualStr := fmt.Sprintf("%v", actualValue)

		if strings.Contains(expectedStr, "*") {
			regexPattern := strings.ReplaceAll(expectedStr, "*", ".*")
			matched, _ := regexp.MatchString(regexPattern, actualStr)
			if !matched {
				return false
			}
		} else if expectedStr != actualStr {
			return false
		}
	}

	return true
}

// checkNumericEquals checks NumericEquals conditions
func (p *PermissionEngine) checkNumericEquals(condition interface{}, context map[string]interface{}) bool {
	conditionMap, ok := condition.(map[string]interface{})
	if !ok {
		return false
	}

	for key, expectedValue := range conditionMap {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		// Simple numeric comparison - in production, you'd want more robust type checking
		if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
			return false
		}
	}

	return true
}

// checkBool checks Bool conditions
func (p *PermissionEngine) checkBool(condition interface{}, context map[string]interface{}) bool {
	conditionMap, ok := condition.(map[string]interface{})
	if !ok {
		return false
	}

	for key, expectedValue := range conditionMap {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		expectedBool, ok1 := expectedValue.(bool)
		actualBool, ok2 := actualValue.(bool)
		if !ok1 || !ok2 {
			return false
		}

		if expectedBool != actualBool {
			return false
		}
	}

	return true
}

// hasExplicitAllow checks if a policy has an explicit allow statement
func (p *PermissionEngine) hasExplicitAllow(policy *Policy, check PermissionCheck) bool {
	statements, err := p.policyService.GetPolicyStatements(policy.ID)
	if err != nil {
		return false
	}

	for _, statement := range statements {
		if statement.Effect == "Allow" &&
			p.actionMatches(check.Action, statement.Actions) &&
			p.resourceMatches(check.Resource, statement.Resources) {
			return true
		}
	}

	return false
}

// PolicyService interface for dependency injection
type PolicyService interface {
	GetUserPolicies(userID, realmID uuid.UUID) ([]*Policy, error)
	GetPolicyStatements(policyID uuid.UUID) ([]*Statement, error)
}
