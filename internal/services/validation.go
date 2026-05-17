package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// ValidationResult represents the result of AI output validation
type ValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// AIOutputValidator validates AI-generated content for hallucinations
type AIOutputValidator struct {
	userProfile string
	jobContext  string
}

// NewAIOutputValidator creates a new validator instance
func NewAIOutputValidator(userProfile, jobContext string) *AIOutputValidator {
	return &AIOutputValidator{
		userProfile: userProfile,
		jobContext:  jobContext,
	}
}

// ValidateCV validates generated CV content
func (v *AIOutputValidator) ValidateCV(cvContent string) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Parse CV content
	var cv map[string]interface{}
	if err := json.Unmarshal([]byte(cvContent), &cv); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid CV format")
		return result
	}

	// Check for placeholder text
	v.checkPlaceholders(cv, &result)

	// Check for unrealistic claims
	v.checkUnrealisticClaims(cv, &result)

	// Check for required sections
	v.checkRequiredSections(cv, &result)

	// Check for consistent dates
	v.checkDateConsistency(cv, &result)

	return result
}

// checkPlaceholders checks for placeholder text that indicates incomplete generation
func (v *AIOutputValidator) checkPlaceholders(cv map[string]interface{}, result *ValidationResult) {
	content, _ := json.Marshal(cv)
	contentStr := string(content)

	placeholders := []string{
		"[Your Name]",
		"[Your Email]",
		"[Phone Number]",
		"[Company Name]",
		"[Position]",
		"[University]",
		"[Degree]",
		"TODO:",
		"PLACEHOLDER",
		"Example",
		"Sample",
	}

	for _, placeholder := range placeholders {
		if strings.Contains(contentStr, placeholder) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Found placeholder: %s", placeholder))
		}
	}
}

// checkUnrealisticClaims checks for unrealistic or exaggerated claims
func (v *AIOutputValidator) checkUnrealisticClaims(cv map[string]interface{}, result *ValidationResult) {
	content, _ := json.Marshal(cv)
	contentStr := strings.ToLower(string(content))

	unrealisticPhrases := []string{
		"10+ years of experience in everything",
		"expert in all programming languages",
		"perfect track record",
		"never failed",
		"best in the world",
		"revolutionary",
		"game-changing",
	}

	for _, phrase := range unrealisticPhrases {
		if strings.Contains(contentStr, phrase) {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Potentially unrealistic claim detected"))
		}
	}

	// Check for too many skills (potential hallucination)
	if skills, ok := cv["skills"].(map[string]interface{}); ok {
		totalSkills := 0
		for _, v := range skills {
			if arr, ok := v.([]interface{}); ok {
				totalSkills += len(arr)
			}
		}
		if totalSkills > 30 {
			result.Warnings = append(result.Warnings, "Unusually high number of skills listed")
		}
	}
}

// checkRequiredSections checks for presence of required CV sections
func (v *AIOutputValidator) checkRequiredSections(cv map[string]interface{}, result *ValidationResult) {
	requiredSections := []string{"contact", "summary", "skills", "experience"}

	for _, section := range requiredSections {
		if _, exists := cv[section]; !exists {
			if section == "contact" {
				// Contact might be nested differently
				if _, exists := cv["name"]; !exists {
					result.Errors = append(result.Errors, fmt.Sprintf("Missing required section: %s", section))
					result.IsValid = false
				}
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Missing required section: %s", section))
				result.IsValid = false
			}
		}
	}
}

// checkDateConsistency checks for date consistency in experience entries
func (v *AIOutputValidator) checkDateConsistency(cv map[string]interface{}, result *ValidationResult) {
	// Check experiences for date issues
	if experiences, ok := cv["experience"].([]interface{}); ok {
		for i, exp := range experiences {
			if expMap, ok := exp.(map[string]interface{}); ok {
				// Check for future dates
				if startDate, ok := expMap["start_date"].(string); ok {
					if strings.Contains(startDate, "2027") || strings.Contains(startDate, "2028") {
						result.Warnings = append(result.Warnings, fmt.Sprintf("Experience %d has future start date", i+1))
					}
				}
			}
		}
	}
}

// SanitizeInput sanitizes user input before sending to AI
func SanitizeInput(input string) string {
	// Remove potential prompt injection attempts
	dangerousPatterns := []string{
		"ignore previous instructions",
		"ignore all above",
		"disregard",
		"system:",
		"assistant:",
	}

	sanitized := input
	for _, pattern := range patternsToRegex(dangerousPatterns) {
		sanitized = pattern.ReplaceAllString(sanitized, "[filtered]")
	}

	return sanitized
}

func patternsToRegex(patterns []string) []*regexp.Regexp {
	regexes := make([]*regexp.Regexp, len(patterns))
	for i, p := range patterns {
		regexes[i], _ = regexp.Compile("(?i)" + regexp.QuoteMeta(p))
	}
	return regexes
}

// ValidateGapAnalysis validates gap analysis output
func (v *AIOutputValidator) ValidateGapAnalysis(analysis string) ValidationResult {
	result := ValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	var gapAnalysis map[string]interface{}
	if err := json.Unmarshal([]byte(analysis), &gapAnalysis); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "Invalid gap analysis format")
		return result
	}

	// Check match score is reasonable
	if score, ok := gapAnalysis["match_score"].(float64); ok {
		if score < 0 || score > 100 {
			result.Warnings = append(result.Warnings, "Match score out of expected range")
		}
	}

	// Check for empty skills arrays
	if skills, ok := gapAnalysis["matching_skills"].([]interface{}); ok {
		if len(skills) == 0 {
			result.Warnings = append(result.Warnings, "No matching skills identified")
		}
	}

	return result
}

// ValidateInterviewResponse validates interview chat responses
func ValidateInterviewResponse(response string) bool {
	// Check response is not empty
	if strings.TrimSpace(response) == "" {
		return false
	}

	// Check response is not too short (potential error)
	if len(response) < 10 {
		return false
	}

	// Check for error indicators
	errorIndicators := []string{
		"i cannot",
		"i'm unable",
		"error:",
		"failed to",
	}

	responseLower := strings.ToLower(response)
	for _, indicator := range errorIndicators {
		if strings.Contains(responseLower, indicator) {
			return false
		}
	}

	return true
}
