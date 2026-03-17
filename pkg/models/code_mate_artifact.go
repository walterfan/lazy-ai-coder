package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// InputType constants for Code Mate Agent classification
const (
	InputTypeResearchSolution = "research_solution"
	InputTypeLearnTech        = "learn_tech"
	InputTypeTechDesign       = "tech_design"
)

// Legacy input types (for tests and migration; do not use in new agent logic)
const (
	InputTypeWord     = "word"
	InputTypeSentence = "sentence"
	InputTypeQuestion = "question"
	InputTypeIdea     = "idea"
	InputTypeTopic    = "topic"
)

// ValidInputTypes returns all valid input types for Code Mate
func ValidInputTypesCodeMate() []string {
	return []string{InputTypeResearchSolution, InputTypeLearnTech, InputTypeTechDesign}
}

// IsValidInputTypeCodeMate checks if the given type is valid for Code Mate
func IsValidInputTypeCodeMate(t string) bool {
	for _, valid := range ValidInputTypesCodeMate() {
		if t == valid {
			return true
		}
	}
	return false
}

// ValidInputTypes returns all valid input types (Code Mate + legacy) for validation/filtering
func ValidInputTypes() []string {
	return append(ValidInputTypesCodeMate(), InputTypeWord, InputTypeSentence, InputTypeQuestion, InputTypeIdea, InputTypeTopic)
}

// IsValidInputType returns true for Code Mate or legacy types (allows migrated data and tests)
func IsValidInputType(t string) bool {
	return IsValidInputTypeCodeMate(t) || t == InputTypeWord || t == InputTypeSentence || t == InputTypeQuestion || t == InputTypeIdea || t == InputTypeTopic
}

// ResponsePayloadData is an alias for CodeMateResponsePayload for compatibility during migration
type ResponsePayloadData = CodeMateResponsePayload

// chatrecord is an alias for CodeMateArtifact for compatibility during migration
type ChatRecord = CodeMateArtifact

// chatrecordSummary is an alias for CodeMateArtifactSummary for compatibility during migration
type ChatRecordSummary = CodeMateArtifactSummary

// CodeMateArtifact represents a user's code-mate artifact (research note, learning plan, or design) in the database
type CodeMateArtifact struct {
	ID              string         `json:"id" gorm:"primaryKey;type:text"`
	InputType       string         `json:"input_type" gorm:"not null;type:text;size:32;index"`
	UserInput       string         `json:"user_input" gorm:"not null;type:text"`
	ResponsePayload string         `json:"-" gorm:"type:text"` // JSON stored as string
	UserID          string         `json:"user_id" gorm:"not null;type:text;index"`
	RealmID         string         `json:"realm_id" gorm:"type:text;index"`
	CreatedBy       string         `json:"created_by" gorm:"type:text"`
	CreatedTime     time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy       string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime     time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for CodeMateArtifact
func (CodeMateArtifact) TableName() string {
	return "code_mate_artifacts"
}

// CodeMateResponsePayload represents the structured response from the Code Mate agent
// Supports research_solution, learn_tech, and tech_design response shapes
type CodeMateResponsePayload struct {
	// Common
	Explanation string `json:"explanation,omitempty"`
	Example     string `json:"example,omitempty"`

	// research_solution: summary, options, trade-offs, recommendation
	Summary        string       `json:"summary,omitempty"`
	Options        []OptionItem `json:"options,omitempty"`
	TradeOffs      string       `json:"trade_offs,omitempty"`
	Recommendation string       `json:"recommendation,omitempty"`
	References     []string     `json:"references,omitempty"`

	// learn_tech: intro, concepts, learning path, resources, prerequisites, time estimate
	Introduction  string         `json:"introduction,omitempty"`
	KeyConcepts   []ConceptItem  `json:"key_concepts,omitempty"`
	LearningPath  []LearningStep `json:"learning_path,omitempty"`
	Resources     []ResourceItem `json:"resources,omitempty"`
	Prerequisites []string       `json:"prerequisites,omitempty"`
	TimeEstimate  string         `json:"time_estimate,omitempty"`

	// tech_design: problem, options, chosen approach, components, risks
	ProblemStatement string   `json:"problem_statement,omitempty"`
	ApproachOptions  []string `json:"approach_options,omitempty"`
	ChosenApproach   string   `json:"chosen_approach,omitempty"`
	Components       []string `json:"components,omitempty"`
	Risks            string   `json:"risks,omitempty"`

	// Legacy/fallback
	Answer        string   `json:"answer,omitempty"`
	Plan          []string `json:"plan,omitempty"`
	Pronunciation string   `json:"pronunciation,omitempty"`
}

// OptionItem for research_solution pros/cons
type OptionItem struct {
	Name    string `json:"name"`
	Pros    string `json:"pros,omitempty"`
	Cons    string `json:"cons,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// ConceptItem represents a key concept (learn_tech)
type ConceptItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Importance  string `json:"importance,omitempty"`
}

// LearningStep represents a step in the learning path
type LearningStep struct {
	Order       int      `json:"order"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Duration    string   `json:"duration,omitempty"`
	Objectives  []string `json:"objectives,omitempty"`
}

// ResourceItem represents a learning resource
type ResourceItem struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	Difficulty  string `json:"difficulty,omitempty"`
}

// GetResponsePayload parses and returns the CodeMateResponsePayload
func (a *CodeMateArtifact) GetResponsePayload() (*CodeMateResponsePayload, error) {
	if a.ResponsePayload == "" {
		return &CodeMateResponsePayload{}, nil
	}
	var payload CodeMateResponsePayload
	if err := json.Unmarshal([]byte(a.ResponsePayload), &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// SetResponsePayload marshals and sets the CodeMateResponsePayload
func (a *CodeMateArtifact) SetResponsePayload(payload *CodeMateResponsePayload) error {
	if payload == nil {
		a.ResponsePayload = ""
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	a.ResponsePayload = string(data)
	return nil
}

// MarshalJSON implements custom JSON marshaling for CodeMateArtifact
func (a CodeMateArtifact) MarshalJSON() ([]byte, error) {
	type Alias CodeMateArtifact
	payload := &CodeMateResponsePayload{}
	if a.ResponsePayload != "" {
		_ = json.Unmarshal([]byte(a.ResponsePayload), payload)
	}
	return json.Marshal(&struct {
		*Alias
		ResponsePayload *CodeMateResponsePayload `json:"response_payload"`
	}{
		Alias:           (*Alias)(&a),
		ResponsePayload: payload,
	})
}

// CodeMateArtifactSummary is a lightweight version for list responses
type CodeMateArtifactSummary struct {
	ID              string    `json:"id"`
	InputType       string    `json:"input_type"`
	UserInput       string    `json:"user_input"`
	ResponseSummary string    `json:"response_summary"`
	CreatedTime     time.Time `json:"created_time"`
}

// ToSummary converts a CodeMateArtifact to CodeMateArtifactSummary
func (a *CodeMateArtifact) ToSummary(maxLen int) CodeMateArtifactSummary {
	summary := a.UserInput
	if len(summary) > maxLen {
		summary = summary[:maxLen] + "..."
	}
	responseSummary := ""
	if payload, err := a.GetResponsePayload(); err == nil {
		if payload.Summary != "" {
			responseSummary = payload.Summary
		} else if payload.Introduction != "" {
			responseSummary = payload.Introduction
		} else if payload.ProblemStatement != "" {
			responseSummary = payload.ProblemStatement
		} else if payload.Explanation != "" {
			responseSummary = payload.Explanation
		} else if payload.Answer != "" {
			responseSummary = payload.Answer
		} else if len(payload.Plan) > 0 {
			responseSummary = payload.Plan[0]
		} else if len(payload.LearningPath) > 0 {
			responseSummary = payload.LearningPath[0].Title
		}
	}
	if len(responseSummary) > maxLen {
		responseSummary = responseSummary[:maxLen] + "..."
	}
	return CodeMateArtifactSummary{
		ID:              a.ID,
		InputType:       a.InputType,
		UserInput:       summary,
		ResponseSummary: responseSummary,
		CreatedTime:     a.CreatedTime,
	}
}
