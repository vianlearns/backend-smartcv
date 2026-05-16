package services

func (s *AIService) ExtractCVData(content string) (string, error) {
	systemPrompt := `You are a CV/Resume parser. Extract all relevant information from the provided text.

Extract the following information and return as JSON:
{
  "name": "Full Name",
  "email": "email@example.com",
  "phone": "phone number",
  "address": "city, country",
  "summary": "professional summary",
  "experiences": [
    {
      "company": "Company Name",
      "position": "Job Title",
      "location": "City, Country",
      "start_date": "YYYY-MM-DD",
      "end_date": "YYYY-MM-DD or null if current",
      "is_current": true/false,
      "description": "job description",
      "achievements": ["achievement 1", "achievement 2"]
    }
  ],
  "education": [
    {
      "institution": "University Name",
      "degree": "Degree Type",
      "field_of_study": "Major/Field",
      "start_date": "YYYY-MM-DD",
      "end_date": "YYYY-MM-DD",
      "gpa": 3.5
    }
  ],
  "skills": [
    {
      "name": "Skill Name",
      "category": "Technical/Soft/Language/etc",
      "proficiency": "Expert/Intermediate/Beginner"
    }
  ],
  "certifications": [
    {
      "name": "Certification Name",
      "issuer": "Issuing Organization",
      "issue_date": "YYYY-MM-DD",
      "credential_id": "ID if available"
    }
  ],
  "projects": [
    {
      "name": "Project Name",
      "description": "Brief description",
      "technologies": ["tech1", "tech2"]
    }
  ]
}

Return ONLY the JSON object. If information is not available, use empty string or empty array.`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: content},
	}

	return s.Chat(messages)
}
