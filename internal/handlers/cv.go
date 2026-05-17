package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"smartcv-backend/internal/database"
	"smartcv-backend/internal/models"
	"smartcv-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var aiService *services.AIService

func InitAI(ai *services.AIService) {
	aiService = ai
}

func UploadCV(c *gin.Context) {
	userID := getUserID(c)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Read file content
	content := make([]byte, header.Size)
	_, err = file.Read(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Parse file and extract CV data using AI
	extractedData, err := aiService.ParseAndExtract(content, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CV: " + err.Error()})
		return
	}

	// Save to database
	var profileData map[string]any
	json.Unmarshal([]byte(extractedData), &profileData)

	// Update or create profile
	_, err = database.DB.Exec(`
		INSERT INTO user_profiles (user_id, full_name, email, phone, address, summary)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			summary = EXCLUDED.summary,
			updated_at = CURRENT_TIMESTAMP
	`, userID, profileData["name"], profileData["email"], profileData["phone"], profileData["address"], profileData["summary"])

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "CV uploaded successfully",
		"data":    profileData,
	})
}

func AnalyzeGap(c *gin.Context) {
	userID := getUserID(c)
	var req models.JobInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user profile data
	userProfile := getUserProfileString(userID)
	jobDescription := fmt.Sprintf("Title: %s\nCompany: %s\nType: %s\n\nDescription: %s\n\nQualifications: %s",
		req.JobTitle, req.Company, req.JobType, req.JobDescription, req.Qualifications)

	result, err := aiService.AnalyzeGap(userProfile, jobDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to analyze gap: " + err.Error()})
		return
	}

	var response models.GapAnalysisResponse
	json.Unmarshal([]byte(result), &response)

	c.JSON(http.StatusOK, response)
}

func GenerateCV(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		JobApplicationID int `json:"job_application_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check credits
	var credits int
	err := database.DB.QueryRow("SELECT credits FROM users WHERE id = $1", userID).Scan(&credits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check credits"})
		return
	}

	if credits < 1 {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient credits"})
		return
	}

	// Get job application
	var job models.JobApplication
	err = database.DB.QueryRow(`SELECT id, user_id, job_title, company, job_type, job_description, qualifications 
		FROM job_applications WHERE id = $1 AND user_id = $2`, req.JobApplicationID, userID).Scan(
		&job.ID, &job.UserID, &job.JobTitle, &job.Company, &job.JobType, &job.JobDescription, &job.Qualifications)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job application not found"})
		return
	}

	// Get user profile
	userProfile := getUserProfileString(userID)
	jobDescription := fmt.Sprintf("Title: %s\nCompany: %s\nType: %s\n\nDescription: %s\n\nQualifications: %s",
		job.JobTitle, job.Company, job.JobType, job.JobDescription, job.Qualifications)

	// Generate CV
	cvContent, err := aiService.GenerateCV(userProfile, jobDescription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CV: " + err.Error()})
		return
	}

	var content map[string]any
	json.Unmarshal([]byte(cvContent), &content)

	// Save CV
	var cvID int
	err = database.DB.QueryRow(`
		INSERT INTO generated_cvs (user_id, job_application_id, title, content, ats_score, version)
		VALUES ($1, $2, $3, $4, 85, 1) RETURNING id
	`, userID, req.JobApplicationID, job.JobTitle+" - "+job.Company, content).Scan(&cvID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save CV"})
		return
	}

	// Deduct credit
	database.DB.Exec("UPDATE users SET credits = credits - 1 WHERE id = $1", userID)
	database.DB.Exec(`INSERT INTO credit_transactions (user_id, amount, type, description) VALUES ($1, -1, 'usage', 'CV Generation')`, userID)

	c.JSON(http.StatusOK, gin.H{
		"id":      cvID,
		"content": content,
		"message": "CV generated successfully",
	})
}

func GetCVs(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT c.id, c.user_id, c.job_application_id, j.job_title, c.title, c.content, c.ats_score, c.version, c.created_at, c.updated_at 
	          FROM generated_cvs c 
			  LEFT JOIN job_applications j ON c.job_application_id = j.id
			  WHERE c.user_id = $1 ORDER BY c.created_at DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var cvs []models.GeneratedCV
	for rows.Next() {
		var cv models.GeneratedCV
		// Use sql.NullString for job_title since it's from a LEFT JOIN
		var jobTitle sql.NullString
		err := rows.Scan(&cv.ID, &cv.UserID, &cv.JobApplicationID, &jobTitle, &cv.Title, &cv.Content, &cv.ATSScore, &cv.Version, &cv.CreatedAt, &cv.UpdatedAt)
		if err == nil && jobTitle.Valid {
			cv.JobTitle = jobTitle.String
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		cvs = append(cvs, cv)
	}

	c.JSON(http.StatusOK, cvs)
}

func GetCV(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	var cv models.GeneratedCV
	query := `SELECT c.id, c.user_id, c.job_application_id, j.job_title, c.title, c.content, c.ats_score, c.version, c.created_at, c.updated_at 
	          FROM generated_cvs c
			  LEFT JOIN job_applications j ON c.job_application_id = j.id
			  WHERE c.id = $1 AND c.user_id = $2`

	var jobTitle sql.NullString
	err := database.DB.QueryRow(query, id, userID).Scan(
		&cv.ID, &cv.UserID, &cv.JobApplicationID, &jobTitle, &cv.Title, &cv.Content, &cv.ATSScore, &cv.Version, &cv.CreatedAt, &cv.UpdatedAt,
	)
	if err == nil && jobTitle.Valid {
		cv.JobTitle = jobTitle.String
	}

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, cv)
}

func UpdateCV(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	var req struct {
		Content map[string]any `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current CV version
	var currentVersion int
	err := database.DB.QueryRow("SELECT version FROM generated_cvs WHERE id = $1 AND user_id = $2", id, userID).Scan(&currentVersion)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
		return
	}

	// Update CV with new version
	_, err = database.DB.Exec(`
		UPDATE generated_cvs SET content = $1, version = $2, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $3 AND user_id = $4
	`, req.Content, currentVersion+1, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update CV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CV updated successfully", "version": currentVersion + 1})
}

func GetComments(c *gin.Context) {
	cvID := c.Param("id")
	userID := getUserID(c)

	query := `SELECT id, cv_id, user_id, section, content, is_resolved, created_at 
	          FROM cv_comments WHERE cv_id = $1 AND user_id = $2 ORDER BY created_at DESC`

	rows, err := database.DB.Query(query, cvID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var comments []models.CVComment
	for rows.Next() {
		var comment models.CVComment
		err := rows.Scan(&comment.ID, &comment.CVID, &comment.UserID, &comment.Section, &comment.Content, &comment.IsResolved, &comment.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		comments = append(comments, comment)
	}

	c.JSON(http.StatusOK, comments)
}

func CreateComment(c *gin.Context) {
	cvID := c.Param("id")
	userID := getUserID(c)

	var req struct {
		Section string `json:"section"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cvIDInt, _ := strconv.Atoi(cvID)
	var commentID int
	err := database.DB.QueryRow(`
		INSERT INTO cv_comments (cv_id, user_id, section, content) 
		VALUES ($1, $2, $3, $4) RETURNING id
	`, cvIDInt, userID, req.Section, req.Content).Scan(&commentID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Apply revision with AI
	var cvContent map[string]any
	database.DB.QueryRow("SELECT content FROM generated_cvs WHERE id = $1", cvIDInt).Scan(&cvContent)

	cvJSON, _ := json.Marshal(cvContent)
	revisedCV, err := aiService.ReviseCV(string(cvJSON), req.Content)
	if err == nil {
		var updatedContent map[string]any
		json.Unmarshal([]byte(revisedCV), &updatedContent)
		database.DB.Exec("UPDATE generated_cvs SET content = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", updatedContent, cvIDInt)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      commentID,
		"message": "Comment created and revision applied",
	})
}

func ResolveComment(c *gin.Context) {
	commentID := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec(`
		UPDATE cv_comments SET is_resolved = TRUE 
		WHERE id = $1 AND user_id = $2
	`, commentID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment resolved"})
}

func ChatWithAI(c *gin.Context) {
	userID := getUserID(c)
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get or create chat session
	var sessionID int
	var messagesJSON []byte
	err := database.DB.QueryRow(`
		SELECT id, messages FROM ai_chat_sessions 
		WHERE user_id = $1 AND job_application_id = $2 
		ORDER BY created_at DESC LIMIT 1
	`, userID, req.JobApplicationID).Scan(&sessionID, &messagesJSON)

	if err == sql.ErrNoRows {
		// Create new session
		err = database.DB.QueryRow(`
			INSERT INTO ai_chat_sessions (user_id, job_application_id, messages) 
			VALUES ($1, $2, '[]') RETURNING id
		`, userID, req.JobApplicationID).Scan(&sessionID)
		messagesJSON = []byte("[]")
	}

	var history []models.ChatMessage
	json.Unmarshal(messagesJSON, &history)

	// Get job and user data
	userProfile := getUserProfileString(userID)
	var job models.JobApplication
	database.DB.QueryRow(`SELECT job_title, job_description FROM job_applications WHERE id = $1`, req.JobApplicationID).Scan(&job.JobTitle, &job.JobDescription)

	// Convert history to services.ChatMessage
	servicesHistory := make([]services.ChatMessage, len(history))
	for i, msg := range history {
		servicesHistory[i] = services.ChatMessage{Role: msg.Role, Content: msg.Content}
	}

	// Chat with AI
	response, err := aiService.ChatInterview(userProfile, job.JobDescription, req.Message, servicesHistory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI error: " + err.Error()})
		return
	}

	// Save to history
	history = append(history, models.ChatMessage{Role: "user", Content: req.Message})
	history = append(history, models.ChatMessage{Role: "assistant", Content: response})

	historyJSON, _ := json.Marshal(history)
	database.DB.Exec(`UPDATE ai_chat_sessions SET messages = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, historyJSON, sessionID)

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"history":  history,
	})
}

func GetChatHistory(c *gin.Context) {
	jobID := c.Param("job_id")
	userID := getUserID(c)

	var messages []models.ChatMessage
	var messagesJSON []byte
	err := database.DB.QueryRow(`
		SELECT messages FROM ai_chat_sessions 
		WHERE user_id = $1 AND job_application_id = $2 
		ORDER BY created_at DESC LIMIT 1
	`, userID, jobID).Scan(&messagesJSON)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, []models.ChatMessage{})
		return
	}

	json.Unmarshal(messagesJSON, &messages)
	c.JSON(http.StatusOK, messages)
}

func GetCredits(c *gin.Context) {
	userID := getUserID(c)

	var credits int
	err := database.DB.QueryRow("SELECT credits FROM users WHERE id = $1", userID).Scan(&credits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get credits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"credits": credits})
}

func PurchaseCredits(c *gin.Context) {
	userID := getUserID(c)
	var req struct {
		Amount int `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate order ID
	orderID := fmt.Sprintf("SMTCV-%d-%d", userID, time.Now().UnixNano())

	// Here you would integrate with Midtrans
	// For now, return the order ID for frontend to handle

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"amount":   req.Amount,
		"message":  "Proceed to payment",
	})
}

func MidtransCallback(c *gin.Context) {
	var notification struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		TransactionStatus string `json:"transaction_status"`
		SignatureKey      string `json:"signature_key"`
	}

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature (in production)
	// Update transaction and add credits if successful
	if notification.TransactionStatus == "capture" || notification.TransactionStatus == "settlement" {
		// Extract user ID from order ID
		var userID int
		database.DB.QueryRow("SELECT user_id FROM credit_transactions WHERE reference_id = $1", notification.OrderID).Scan(&userID)

		// Add credits
		credits := 0
		amount := 0
		fmt.Sscanf(notification.GrossAmount, "%d", &amount)
		credits = amount / 10000 // 1 credit = Rp 10,000

		database.DB.Exec("UPDATE users SET credits = credits + $1 WHERE id = $2", credits, userID)
		database.DB.Exec(`INSERT INTO credit_transactions (user_id, amount, type, description, reference_id) VALUES ($1, $2, 'purchase', 'Credit purchase', $3)`, userID, credits, notification.OrderID)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func DownloadCV(c *gin.Context) {
	cvID := c.Param("id")
	userID := getUserID(c)

	var cv models.GeneratedCV
	var title string
	err := database.DB.QueryRow("SELECT content, title FROM generated_cvs WHERE id = $1 AND user_id = $2", cvID, userID).Scan(&cv.Content, &title)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CV not found"})
		return
	}

	// Parse CV content to PDFContent struct
	contentBytes, err := json.Marshal(cv.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CV content"})
		return
	}

	var pdfContent services.CVContent
	if err := json.Unmarshal(contentBytes, &pdfContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CV content structure"})
		return
	}

	// Generate PDF
	pdfGenerator := services.NewPDFGenerator()
	pdfBytes, err := pdfGenerator.Generate(pdfContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	// Set response headers for file download
	filename := fmt.Sprintf("SmartCV_%s.pdf", sanitizeFilename(title))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(name string) string {
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result += string(r)
		} else if r == ' ' {
			result += "_"
		}
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func getUserProfileString(userID int) string {
	var profile struct {
		FullName string
		Email    string
		Phone    string
		Address  string
		Summary  string
	}

	database.DB.QueryRow(`SELECT COALESCE(full_name, ''), COALESCE(email, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(summary, '') 
		FROM user_profiles WHERE user_id = $1`, userID).Scan(&profile.FullName, &profile.Email, &profile.Phone, &profile.Address, &profile.Summary)

	// Get experiences
	rows, _ := database.DB.Query(`SELECT company, position, start_date, end_date, description FROM experiences WHERE user_id = $1`, userID)
	defer rows.Close()

	experiences := ""
	for rows.Next() {
		var company, position, description string
		var startDate, endDate string
		rows.Scan(&company, &position, &startDate, &endDate, &description)
		experiences += fmt.Sprintf("- %s at %s (%s - %s): %s\n", position, company, startDate, endDate, description)
	}

	// Get skills
	rows2, _ := database.DB.Query("SELECT name, category FROM skills WHERE user_id = $1", userID)
	defer rows2.Close()

	skills := ""
	for rows2.Next() {
		var name, category string
		rows2.Scan(&name, &category)
		skills += fmt.Sprintf("- %s (%s)\n", name, category)
	}

	// Get education
	rows3, _ := database.DB.Query(`SELECT institution, degree, field_of_study FROM educations WHERE user_id = $1`, userID)
	defer rows3.Close()

	education := ""
	for rows3.Next() {
		var institution, degree, field string
		rows3.Scan(&institution, &degree, &field)
		education += fmt.Sprintf("- %s in %s at %s\n", degree, field, institution)
	}

	return fmt.Sprintf(`NAME: %s
EMAIL: %s
PHONE: %s
ADDRESS: %s

SUMMARY:
%s

EXPERIENCE:
%s

EDUCATION:
%s

SKILLS:
%s`, profile.FullName, profile.Email, profile.Phone, profile.Address, profile.Summary, experiences, education, skills)
}
