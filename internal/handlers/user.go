package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"smartcv-backend/internal/database"
	"smartcv-backend/internal/middleware"
	"smartcv-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	clerkID := middleware.GetUserFromContext(c)

	var user models.User
	query := `SELECT id, clerk_id, email, full_name, credits, created_at, updated_at FROM users WHERE clerk_id = $1`
	err := database.DB.QueryRow(query, clerkID).Scan(
		&user.ID, &user.ClerkID, &user.Email, &user.FullName, &user.Credits, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create new user
		insertQuery := `INSERT INTO users (clerk_id, email, full_name, credits) VALUES ($1, $2, $3, 3) RETURNING id, credits, created_at, updated_at`
		email := c.GetHeader("X-User-Email")
		name := c.GetHeader("X-User-Name")

		err = database.DB.QueryRow(insertQuery, clerkID, email, name).Scan(&user.ID, &user.Credits, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		user.ClerkID = clerkID
		user.Email = email
		user.FullName = name
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	var req struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	clerkID := middleware.GetUserFromContext(c)
	query := `UPDATE users SET full_name = $1, email = $2, updated_at = CURRENT_TIMESTAMP WHERE clerk_id = $3`

	_, err := database.DB.Exec(query, req.FullName, req.Email, clerkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func GetProfile(c *gin.Context) {
	userID := getUserID(c)

	var profile models.UserProfile
	query := `SELECT id, user_id, full_name, email, phone, address, summary, linked_in, github, website, created_at, updated_at 
	          FROM user_profiles WHERE user_id = $1`

	err := database.DB.QueryRow(query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.FullName, &profile.Email, &profile.Phone,
		&profile.Address, &profile.Summary, &profile.LinkedIn, &profile.GitHub, &profile.Website,
		&profile.CreatedAt, &profile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusOK, nil)
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func UpdateProfile(c *gin.Context) {
	var profile models.UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)

	query := `
		INSERT INTO user_profiles (user_id, full_name, email, phone, address, summary, linked_in, github, website)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			address = EXCLUDED.address,
			summary = EXCLUDED.summary,
			linked_in = EXCLUDED.linked_in,
			github = EXCLUDED.github,
			website = EXCLUDED.website,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := database.DB.Exec(query, userID, profile.FullName, profile.Email, profile.Phone,
		profile.Address, profile.Summary, profile.LinkedIn, profile.GitHub, profile.Website)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func GetExperiences(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, company, position, location, start_date, end_date, is_current, description, achievements, created_at 
	          FROM experiences WHERE user_id = $1 ORDER BY start_date DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var experiences []models.Experience
	for rows.Next() {
		var exp models.Experience
		var achievements []byte
		err := rows.Scan(
			&exp.ID, &exp.UserID, &exp.Company, &exp.Position, &exp.Location,
			&exp.StartDate, &exp.EndDate, &exp.IsCurrent, &exp.Description, &achievements, &exp.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		json.Unmarshal(achievements, &exp.Achievements)
		experiences = append(experiences, exp)
	}

	c.JSON(http.StatusOK, experiences)
}

func CreateExperience(c *gin.Context) {
	var exp models.Experience
	if err := c.ShouldBindJSON(&exp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO experiences (user_id, company, position, location, start_date, end_date, is_current, description, achievements)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := database.DB.QueryRow(query, userID, exp.Company, exp.Position, exp.Location,
		exp.StartDate, exp.EndDate, exp.IsCurrent, exp.Description, exp.Achievements).Scan(&exp.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create experience"})
		return
	}

	c.JSON(http.StatusCreated, exp)
}

func UpdateExperience(c *gin.Context) {
	id := c.Param("id")
	var exp models.Experience
	if err := c.ShouldBindJSON(&exp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `UPDATE experiences SET company = $1, position = $2, location = $3, start_date = $4, end_date = $5, is_current = $6, description = $7, achievements = $8
	          WHERE id = $9 AND user_id = $10`

	_, err := database.DB.Exec(query, exp.Company, exp.Position, exp.Location,
		exp.StartDate, exp.EndDate, exp.IsCurrent, exp.Description, exp.Achievements, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update experience"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Experience updated successfully"})
}

func DeleteExperience(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM experiences WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete experience"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Experience deleted successfully"})
}

func GetEducations(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, institution, degree, field_of_study, location, start_date, end_date, gpa, achievements, created_at 
	          FROM educations WHERE user_id = $1 ORDER BY start_date DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var educations []models.Education
	for rows.Next() {
		var edu models.Education
		var achievements []byte
		err := rows.Scan(
			&edu.ID, &edu.UserID, &edu.Institution, &edu.Degree, &edu.FieldOfStudy, &edu.Location,
			&edu.StartDate, &edu.EndDate, &edu.GPA, &achievements, &edu.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		json.Unmarshal(achievements, &edu.Achievements)
		educations = append(educations, edu)
	}

	c.JSON(http.StatusOK, educations)
}

func CreateEducation(c *gin.Context) {
	var edu models.Education
	if err := c.ShouldBindJSON(&edu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO educations (user_id, institution, degree, field_of_study, location, start_date, end_date, gpa, achievements)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	err := database.DB.QueryRow(query, userID, edu.Institution, edu.Degree, edu.FieldOfStudy,
		edu.Location, edu.StartDate, edu.EndDate, edu.GPA, edu.Achievements).Scan(&edu.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create education"})
		return
	}

	c.JSON(http.StatusCreated, edu)
}

func UpdateEducation(c *gin.Context) {
	id := c.Param("id")
	var edu models.Education
	if err := c.ShouldBindJSON(&edu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `UPDATE educations SET institution = $1, degree = $2, field_of_study = $3, location = $4, start_date = $5, end_date = $6, gpa = $7, achievements = $8
	          WHERE id = $9 AND user_id = $10`

	_, err := database.DB.Exec(query, edu.Institution, edu.Degree, edu.FieldOfStudy,
		edu.Location, edu.StartDate, edu.EndDate, edu.GPA, edu.Achievements, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update education"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Education updated successfully"})
}

func DeleteEducation(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM educations WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete education"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Education deleted successfully"})
}

func GetSkills(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, name, category, proficiency, created_at FROM skills WHERE user_id = $1 ORDER BY category, name`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var skills []models.Skill
	for rows.Next() {
		var s models.Skill
		err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Category, &s.Proficiency, &s.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		skills = append(skills, s)
	}

	c.JSON(http.StatusOK, skills)
}

func CreateSkill(c *gin.Context) {
	var s models.Skill
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO skills (user_id, name, category, proficiency) VALUES ($1, $2, $3, $4) RETURNING id`

	err := database.DB.QueryRow(query, userID, s.Name, s.Category, s.Proficiency).Scan(&s.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create skill"})
		return
	}

	c.JSON(http.StatusCreated, s)
}

func UpdateSkill(c *gin.Context) {
	id := c.Param("id")
	var s models.Skill
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `UPDATE skills SET name = $1, category = $2, proficiency = $3 WHERE id = $4 AND user_id = $5`

	_, err := database.DB.Exec(query, s.Name, s.Category, s.Proficiency, id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill updated successfully"})
}

func DeleteSkill(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM skills WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill deleted successfully"})
}

func GetCertifications(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, name, issuer, issue_date, expiry_date, credential_id, credential_url, created_at 
	          FROM certifications WHERE user_id = $1 ORDER BY issue_date DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var certifications []models.Certification
	for rows.Next() {
		var cert models.Certification
		err := rows.Scan(&cert.ID, &cert.UserID, &cert.Name, &cert.Issuer, &cert.IssueDate,
			&cert.ExpiryDate, &cert.CredentialID, &cert.CredentialURL, &cert.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		certifications = append(certifications, cert)
	}

	c.JSON(http.StatusOK, certifications)
}

func CreateCertification(c *gin.Context) {
	var cert models.Certification
	if err := c.ShouldBindJSON(&cert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO certifications (user_id, name, issuer, issue_date, expiry_date, credential_id, credential_url)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := database.DB.QueryRow(query, userID, cert.Name, cert.Issuer, cert.IssueDate,
		cert.ExpiryDate, cert.CredentialID, cert.CredentialURL).Scan(&cert.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create certification"})
		return
	}

	c.JSON(http.StatusCreated, cert)
}

func DeleteCertification(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM certifications WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete certification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certification deleted successfully"})
}

func GetProjects(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, name, description, technologies, url, start_date, end_date, created_at 
	          FROM projects WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var technologies []byte
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &technologies,
			&p.URL, &p.StartDate, &p.EndDate, &p.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		json.Unmarshal(technologies, &p.Technologies)
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, projects)
}

func CreateProject(c *gin.Context) {
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO projects (user_id, name, description, technologies, url, start_date, end_date)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := database.DB.QueryRow(query, userID, p.Name, p.Description, p.Technologies,
		p.URL, p.StartDate, p.EndDate).Scan(&p.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `UPDATE projects SET name = $1, description = $2, technologies = $3, url = $4, start_date = $5, end_date = $6
	          WHERE id = $7 AND user_id = $8`

	_, err := database.DB.Exec(query, p.Name, p.Description, p.Technologies, p.URL,
		p.StartDate, p.EndDate, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully"})
}

func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM projects WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func getUserID(c *gin.Context) int {
	clerkID := middleware.GetUserFromContext(c)
	var userID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE clerk_id = $1", clerkID).Scan(&userID)
	if err != nil {
		return 0
	}
	return userID
}
