package models

import "time"

type User struct {
	ID        int       `json:"id"`
	ClerkID   string    `json:"clerk_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Credits   int       `json:"credits"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserProfile struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Address    string    `json:"address"`
	Summary    string    `json:"summary"`
	LinkedIn   string    `json:"linked_in"`
	GitHub     string    `json:"github"`
	Website    string    `json:"website"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Experience struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Company      string    `json:"company"`
	Position     string    `json:"position"`
	Location     string    `json:"location"`
	StartDate    time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	IsCurrent    bool      `json:"is_current"`
	Description  string    `json:"description"`
	Achievements []string  `json:"achievements"`
	CreatedAt    time.Time `json:"created_at"`
}

type Education struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Institution   string    `json:"institution"`
	Degree        string    `json:"degree"`
	FieldOfStudy  string    `json:"field_of_study"`
	Location      string    `json:"location"`
	StartDate     time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	GPA           float64   `json:"gpa"`
	Achievements  []string  `json:"achievements"`
	CreatedAt     time.Time `json:"created_at"`
}

type Skill struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Proficiency string    `json:"proficiency"`
	CreatedAt   time.Time `json:"created_at"`
}

type Certification struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Name          string     `json:"name"`
	Issuer        string     `json:"issuer"`
	IssueDate     *time.Time `json:"issue_date"`
	ExpiryDate    *time.Time `json:"expiry_date"`
	CredentialID  string     `json:"credential_id"`
	CredentialURL string     `json:"credential_url"`
	CreatedAt     time.Time  `json:"created_at"`
}

type Project struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Technologies  []string   `json:"technologies"`
	URL           string     `json:"url"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	CreatedAt     time.Time  `json:"created_at"`
}

type JobApplication struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	JobTitle        string    `json:"job_title"`
	Company         string    `json:"company"`
	JobType         string    `json:"job_type"`
	JobDescription  string    `json:"job_description"`
	Qualifications  string    `json:"qualifications"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type GeneratedCV struct {
	ID               int            `json:"id"`
	UserID           int            `json:"user_id"`
	JobApplicationID int            `json:"job_application_id"`
	Title            string         `json:"title"`
	Content          map[string]any `json:"content"`
	ATSScore         int            `json:"ats_score"`
	Version          int            `json:"version"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type CVComment struct {
	ID          int       `json:"id"`
	CVID        int       `json:"cv_id"`
	UserID      int       `json:"user_id"`
	Section     string    `json:"section"`
	Content     string    `json:"content"`
	IsResolved  bool      `json:"is_resolved"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreditTransaction struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Amount      int       `json:"amount"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	ReferenceID string    `json:"reference_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type AIChatSession struct {
	ID               int              `json:"id"`
	UserID           int              `json:"user_id"`
	JobApplicationID int              `json:"job_application_id"`
	Messages         []map[string]any `json:"messages"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// Request/Response DTOs

type CVUploadRequest struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

type JobInputRequest struct {
	JobTitle       string `json:"job_title" binding:"required"`
	Company        string `json:"company"`
	JobType        string `json:"job_type"`
	JobDescription string `json:"job_description" binding:"required"`
	Qualifications string `json:"qualifications"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	JobApplicationID int    `json:"job_application_id"`
	Message          string `json:"message"`
}

type CommentRequest struct {
	CVID    int    `json:"cv_id" binding:"required"`
	Section string `json:"section"`
	Content string `json:"content" binding:"required"`
}

type MidtransRequest struct {
	OrderID string `json:"order_id"`
	Amount  int    `json:"amount"`
}

type GapAnalysisResponse struct {
	MatchScore     int                  `json:"match_score"`
	MatchingSkills []string             `json:"matching_skills"`
	MissingSkills  []string             `json:"missing_skills"`
	Recommendations []string            `json:"recommendations"`
	Questions      []string             `json:"questions"`
}
