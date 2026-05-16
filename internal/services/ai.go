package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AIService struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewAIService(baseURL, model string) *AIService {
	return &AIService{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (s *AIService) Chat(messages []ChatMessage) (string, error) {
	req := ChatRequest{
		Model:       s.Model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   4096,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", s.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "HermesAgent/1.0")

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API error: %s - %s", resp.Status, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (s *AIService) AnalyzeGap(userProfile, jobDescription string) (string, error) {
	systemPrompt := `You are an expert career advisor and resume analyst. Your task is to analyze the gap between a user's profile and a job description.

Analyze the following:
1. Match Score (0-100): How well the profile matches the job requirements
2. Matching Skills: Skills the user has that are required for the job
3. Missing Skills: Required skills the user doesn't have
4. Recommendations: Specific suggestions to improve the CV for this job
5. Clarification Questions: Questions to ask the user to fill information gaps

IMPORTANT: 
- Never hallucinate or invent skills or experiences the user doesn't have
- If information is missing, ask clarifying questions
- Be constructive and specific in recommendations

Respond in JSON format:
{
  "match_score": <number>,
  "matching_skills": ["skill1", "skill2"],
  "missing_skills": ["skill1", "skill2"],
  "recommendations": ["rec1", "rec2"],
  "questions": ["q1", "q2"]
}`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("USER PROFILE:\n%s\n\nJOB DESCRIPTION:\n%s", userProfile, jobDescription)},
	}

	return s.Chat(messages)
}

func (s *AIService) GenerateCV(userProfile, jobDescription string) (string, error) {
	systemPrompt := `You are an expert resume writer specializing in ATS-optimized CVs. Create a tailored resume based on the user's profile for the specific job.

RULES:
1. Use only information from the user's profile - NEVER invent or hallucinate
2. Optimize keywords for ATS systems based on the job description
3. Use classic ATS-friendly format: clear sections, standard fonts, left alignment
4. Quantify achievements where possible
5. Use action verbs and concise bullet points
6. Tailor the summary to highlight relevant experience

SECTIONS TO INCLUDE:
{
  "contact": {
    "name": "Full Name",
    "email": "email",
    "phone": "phone",
    "location": "city, country",
    "linkedin": "url",
    "github": "url"
  },
  "summary": "Professional summary tailored to the job (3-4 sentences)",
  "skills": {
    "technical": ["skill1", "skill2"],
    "tools": ["tool1", "tool2"],
    "languages": ["language1"],
    "other": ["other"]
  },
  "experience": [
    {
      "company": "Company Name",
      "position": "Job Title",
      "location": "City, Country",
      "duration": "Month Year - Month Year",
      "highlights": ["bullet1", "bullet2"]
    }
  ],
  "education": [
    {
      "institution": "University Name",
      "degree": "Degree",
      "field": "Field of Study",
      "year": "Year"
    }
  ],
  "projects": [
    {
      "name": "Project Name",
      "description": "Brief description",
      "technologies": ["tech1", "tech2"]
    }
  ],
  "certifications": [
    {
      "name": "Certification Name",
      "issuer": "Issuing Organization",
      "year": "Year"
    }
  ]
}

Return ONLY the JSON object, no additional text.`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("USER PROFILE:\n%s\n\nJOB DESCRIPTION:\n%s", userProfile, jobDescription)},
	}

	return s.Chat(messages)
}

func (s *AIService) ChatInterview(userProfile, jobDescription, userMessage string, history []ChatMessage) (string, error) {
	systemPrompt := fmt.Sprintf(`You are a career counselor conducting an interactive interview to gather more information about the user's professional background for a job application.

USER PROFILE (what we know):
%s

TARGET JOB:
%s

YOUR ROLE:
1. Ask specific questions to fill gaps in the user's profile
2. Verify information that seems unclear
3. Help the user remember relevant experiences they might have forgotten
4. Be encouraging but professional
5. Ask ONE question at a time
6. Acknowledge the user's response before asking the next question

IMPORTANT:
- Never assume or invent information
- If the user mentions something new, ask for details
- If the user doesn't know something, move on to another topic
- Keep questions specific and actionable`, userProfile, jobDescription)

	messages := append([]ChatMessage{{Role: "system", Content: systemPrompt}}, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: userMessage})

	return s.Chat(messages)
}

func (s *AIService) ReviseCV(cvContent, comment string) (string, error) {
	systemPrompt := `You are a resume editor. Apply the requested revision to the CV while maintaining ATS-friendly format.

RULES:
1. Only modify what the user requested
2. Keep the rest of the CV unchanged
3. Maintain consistent formatting
4. Return the complete updated CV JSON

Return ONLY the updated JSON object.`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("CURRENT CV:\n%s\n\nREVISION REQUEST:\n%s", cvContent, comment)},
	}

	return s.Chat(messages)
}
