package services

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct{}

func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

// CVContent represents the structured CV data for PDF generation
type CVContent struct {
	Contact struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Location string `json:"location"`
		LinkedIn string `json:"linkedin"`
		GitHub   string `json:"github"`
	} `json:"contact"`
	Summary string `json:"summary"`
	Skills  struct {
		Technical []string `json:"technical"`
		Tools     []string `json:"tools"`
		Languages []string `json:"languages"`
		Other     []string `json:"other"`
	} `json:"skills"`
	Experience []struct {
		Company    string   `json:"company"`
		Position   string   `json:"position"`
		Location   string   `json:"location"`
		Duration   string   `json:"duration"`
		Highlights []string `json:"highlights"`
	} `json:"experience"`
	Education []struct {
		Institution string `json:"institution"`
		Degree      string `json:"degree"`
		Field       string `json:"field"`
		Year        string `json:"year"`
	} `json:"education"`
	Projects []struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Technologies []string `json:"technologies"`
	} `json:"projects"`
	Certifications []struct {
		Name   string `json:"name"`
		Issuer string `json:"issuer"`
		Year   string `json:"year"`
	} `json:"certifications"`
}

// Generate generates an ATS-friendly PDF from CV content
func (p *PDFGenerator) Generate(content CVContent) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetMargins(15, 15, 15)

	// Use standard fonts for ATS compatibility
	pdf.SetFont("Arial", "B", 16)

	// Header - Name
	if content.Contact.Name != "" {
		pdf.Cell(0, 10, content.Contact.Name)
		pdf.Ln(8)
	}

	// Contact line
	pdf.SetFont("Arial", "", 10)
	contactParts := []string{}
	if content.Contact.Email != "" {
		contactParts = append(contactParts, content.Contact.Email)
	}
	if content.Contact.Phone != "" {
		contactParts = append(contactParts, content.Contact.Phone)
	}
	if content.Contact.Location != "" {
		contactParts = append(contactParts, content.Contact.Location)
	}
	if len(contactParts) > 0 {
		pdf.Cell(0, 5, joinString(contactParts, " | "))
		pdf.Ln(5)
	}

	// Links
	linkParts := []string{}
	if content.Contact.LinkedIn != "" {
		linkParts = append(linkParts, "LinkedIn: "+content.Contact.LinkedIn)
	}
	if content.Contact.GitHub != "" {
		linkParts = append(linkParts, "GitHub: "+content.Contact.GitHub)
	}
	if len(linkParts) > 0 {
		pdf.SetFont("Arial", "I", 9)
		pdf.Cell(0, 5, joinString(linkParts, " | "))
		pdf.Ln(8)
	}

	// Horizontal line
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)

	// Summary
	if content.Summary != "" {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "PROFESSIONAL SUMMARY")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(0, 5, content.Summary, "", "L", false)
		pdf.Ln(3)
	}

	// Skills
	if len(content.Skills.Technical) > 0 || len(content.Skills.Tools) > 0 ||
		len(content.Skills.Languages) > 0 || len(content.Skills.Other) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "SKILLS")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)

		if len(content.Skills.Technical) > 0 {
			pdf.Cell(30, 5, "Technical:")
			pdf.MultiCell(0, 5, joinString(content.Skills.Technical, ", "), "", "L", false)
		}
		if len(content.Skills.Tools) > 0 {
			pdf.Cell(30, 5, "Tools:")
			pdf.MultiCell(0, 5, joinString(content.Skills.Tools, ", "), "", "L", false)
		}
		if len(content.Skills.Languages) > 0 {
			pdf.Cell(30, 5, "Languages:")
			pdf.MultiCell(0, 5, joinString(content.Skills.Languages, ", "), "", "L", false)
		}
		if len(content.Skills.Other) > 0 {
			pdf.Cell(30, 5, "Other:")
			pdf.MultiCell(0, 5, joinString(content.Skills.Other, ", "), "", "L", false)
		}
		pdf.Ln(3)
	}

	// Experience
	if len(content.Experience) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "PROFESSIONAL EXPERIENCE")
		pdf.Ln(6)

		for _, exp := range content.Experience {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(0, 6, exp.Company)
			pdf.Ln(5)

			pdf.SetFont("Arial", "", 10)
			line := exp.Position
			if exp.Duration != "" {
				line += " | " + exp.Duration
			}
			if exp.Location != "" {
				line += " | " + exp.Location
			}
			pdf.Cell(0, 5, line)
			pdf.Ln(5)

			for _, highlight := range exp.Highlights {
				pdf.Cell(5, 5, "")
				pdf.MultiCell(0, 5, "- "+highlight, "", "L", false)
			}
			pdf.Ln(2)
		}
	}

	// Education
	if len(content.Education) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "EDUCATION")
		pdf.Ln(6)

		for _, edu := range content.Education {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(0, 5, edu.Institution)
			pdf.Ln(5)
			pdf.SetFont("Arial", "", 10)
			line := edu.Degree
			if edu.Field != "" {
				line += " in " + edu.Field
			}
			if edu.Year != "" {
				line += " | " + edu.Year
			}
			pdf.Cell(0, 5, line)
			pdf.Ln(5)
		}
		pdf.Ln(2)
	}

	// Projects
	if len(content.Projects) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "PROJECTS")
		pdf.Ln(6)

		for _, proj := range content.Projects {
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(0, 5, proj.Name)
			pdf.Ln(5)
			if proj.Description != "" {
				pdf.SetFont("Arial", "", 10)
				pdf.MultiCell(0, 5, proj.Description, "", "L", false)
			}
			if len(proj.Technologies) > 0 {
				pdf.SetFont("Arial", "I", 9)
				pdf.Cell(0, 5, "Technologies: "+joinString(proj.Technologies, ", "))
				pdf.Ln(5)
			}
			pdf.Ln(2)
		}
	}

	// Certifications
	if len(content.Certifications) > 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, "CERTIFICATIONS")
		pdf.Ln(6)

		for _, cert := range content.Certifications {
			pdf.SetFont("Arial", "", 10)
			line := cert.Name
			if cert.Issuer != "" {
				line += " - " + cert.Issuer
			}
			if cert.Year != "" {
				line += " (" + cert.Year + ")"
			}
			pdf.Cell(0, 5, line)
			pdf.Ln(5)
		}
	}

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

func joinString(items []string, sep string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += sep
		}
		result += item
	}
	return result
}
