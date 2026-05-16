package handlers

import (
	"database/sql"
	"net/http"
	"smartcv-backend/internal/database"
	"smartcv-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GetJobApplications(c *gin.Context) {
	userID := getUserID(c)

	query := `SELECT id, user_id, job_title, company, job_type, job_description, qualifications, status, created_at, updated_at 
	          FROM job_applications WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var jobs []models.JobApplication
	for rows.Next() {
		var job models.JobApplication
		err := rows.Scan(&job.ID, &job.UserID, &job.JobTitle, &job.Company, &job.JobType,
			&job.JobDescription, &job.Qualifications, &job.Status, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Scan error"})
			return
		}
		jobs = append(jobs, job)
	}

	c.JSON(http.StatusOK, jobs)
}

func GetJobApplication(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	var job models.JobApplication
	query := `SELECT id, user_id, job_title, company, job_type, job_description, qualifications, status, created_at, updated_at 
	          FROM job_applications WHERE id = $1 AND user_id = $2`

	err := database.DB.QueryRow(query, id, userID).Scan(
		&job.ID, &job.UserID, &job.JobTitle, &job.Company, &job.JobType,
		&job.JobDescription, &job.Qualifications, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job application not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, job)
}

func CreateJobApplication(c *gin.Context) {
	var req models.JobInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `INSERT INTO job_applications (user_id, job_title, company, job_type, job_description, qualifications, status)
	          VALUES ($1, $2, $3, $4, $5, $6, 'draft') RETURNING id, created_at, updated_at`

	var job models.JobApplication
	job.UserID = userID
	job.JobTitle = req.JobTitle
	job.Company = req.Company
	job.JobType = req.JobType
	job.JobDescription = req.JobDescription
	job.Qualifications = req.Qualifications

	err := database.DB.QueryRow(query, userID, req.JobTitle, req.Company, req.JobType,
		req.JobDescription, req.Qualifications).Scan(&job.ID, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job application"})
		return
	}

	c.JSON(http.StatusCreated, job)
}

func UpdateJobApplication(c *gin.Context) {
	id := c.Param("id")
	var req models.JobInputRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := getUserID(c)
	query := `UPDATE job_applications SET job_title = $1, company = $2, job_type = $3, job_description = $4, qualifications = $5, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $6 AND user_id = $7`

	_, err := database.DB.Exec(query, req.JobTitle, req.Company, req.JobType,
		req.JobDescription, req.Qualifications, id, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update job application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job application updated successfully"})
}

func DeleteJobApplication(c *gin.Context) {
	id := c.Param("id")
	userID := getUserID(c)

	_, err := database.DB.Exec("DELETE FROM job_applications WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete job application"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Job application deleted successfully"})
}
