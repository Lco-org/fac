package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// EmailTemplateService handles email template CRUD operations
type EmailTemplateService struct {
	db *database.DB
}

// NewEmailTemplateService creates a new email template service
func NewEmailTemplateService(db *database.DB) *EmailTemplateService {
	return &EmailTemplateService{db: db}
}

// ListTemplates returns all email templates
func (s *EmailTemplateService) ListTemplates(ctx context.Context) ([]models.EmailTemplate, error) {
	query := `
		SELECT id, template_key, name, name_ar, subject, subject_ar, 
		       body_html, body_text, description, variables, is_active,
		       created_at, updated_at, created_by
		FROM email_templates
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	var templates []models.EmailTemplate
	for rows.Next() {
		var t models.EmailTemplate
		var variablesJSON []byte
		err := rows.Scan(
			&t.ID, &t.TemplateKey, &t.Name, &t.NameAr, &t.Subject, &t.SubjectAr,
			&t.BodyHTML, &t.BodyText, &t.Description, &variablesJSON, &t.IsActive,
			&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		if variablesJSON != nil {
			json.Unmarshal(variablesJSON, &t.Variables)
		}
		templates = append(templates, t)
	}

	return templates, nil
}

// GetTemplateByKey returns a template by its key
func (s *EmailTemplateService) GetTemplateByKey(ctx context.Context, key string) (*models.EmailTemplate, error) {
	query := `
		SELECT id, template_key, name, name_ar, subject, subject_ar, 
		       body_html, body_text, description, variables, is_active,
		       created_at, updated_at, created_by
		FROM email_templates
		WHERE template_key = $1 AND is_active = true
	`

	var t models.EmailTemplate
	var variablesJSON []byte
	err := s.db.QueryRowContext(ctx, query, key).Scan(
		&t.ID, &t.TemplateKey, &t.Name, &t.NameAr, &t.Subject, &t.SubjectAr,
		&t.BodyHTML, &t.BodyText, &t.Description, &variablesJSON, &t.IsActive,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if variablesJSON != nil {
		json.Unmarshal(variablesJSON, &t.Variables)
	}
	return &t, nil
}

// GetTemplateByID returns a template by its ID
func (s *EmailTemplateService) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.EmailTemplate, error) {
	query := `
		SELECT id, template_key, name, name_ar, subject, subject_ar, 
		       body_html, body_text, description, variables, is_active,
		       created_at, updated_at, created_by
		FROM email_templates
		WHERE id = $1
	`

	var t models.EmailTemplate
	var variablesJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.TemplateKey, &t.Name, &t.NameAr, &t.Subject, &t.SubjectAr,
		&t.BodyHTML, &t.BodyText, &t.Description, &variablesJSON, &t.IsActive,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	if variablesJSON != nil {
		json.Unmarshal(variablesJSON, &t.Variables)
	}
	return &t, nil
}

// UpdateTemplate updates an email template
func (s *EmailTemplateService) UpdateTemplate(ctx context.Context, id uuid.UUID, req *models.UpdateEmailTemplateRequest) (*models.EmailTemplate, error) {
	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}
	if req.NameAr != nil {
		updates = append(updates, fmt.Sprintf("name_ar = $%d", argIndex))
		args = append(args, *req.NameAr)
		argIndex++
	}
	if req.Subject != nil {
		updates = append(updates, fmt.Sprintf("subject = $%d", argIndex))
		args = append(args, *req.Subject)
		argIndex++
	}
	if req.SubjectAr != nil {
		updates = append(updates, fmt.Sprintf("subject_ar = $%d", argIndex))
		args = append(args, *req.SubjectAr)
		argIndex++
	}
	if req.BodyHTML != nil {
		updates = append(updates, fmt.Sprintf("body_html = $%d", argIndex))
		args = append(args, *req.BodyHTML)
		argIndex++
	}
	if req.BodyText != nil {
		updates = append(updates, fmt.Sprintf("body_text = $%d", argIndex))
		args = append(args, *req.BodyText)
		argIndex++
	}
	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	if req.Variables != nil {
		variablesJSON, _ := json.Marshal(req.Variables)
		updates = append(updates, fmt.Sprintf("variables = $%d", argIndex))
		args = append(args, variablesJSON)
		argIndex++
	}
	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return s.GetTemplateByID(ctx, id)
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE email_templates
		SET %s
		WHERE id = $%d
	`, strings.Join(updates, ", "), argIndex)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return s.GetTemplateByID(ctx, id)
}

// RenderTemplate renders a template with provided variables
func (s *EmailTemplateService) RenderTemplate(ctx context.Context, templateKey string, variables map[string]string, useArabic bool) (*models.TemplatePreviewResponse, error) {
	template, err := s.GetTemplateByKey(ctx, templateKey)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateKey)
	}

	// Choose subject and body based on language
	subject := template.Subject
	body := template.BodyHTML

	if useArabic && template.SubjectAr != nil && *template.SubjectAr != "" {
		subject = *template.SubjectAr
	}

	// Replace variables in subject and body
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		subject = strings.ReplaceAll(subject, placeholder, value)
		body = strings.ReplaceAll(body, placeholder, value)
	}

	return &models.TemplatePreviewResponse{
		Subject: subject,
		Body:    body,
	}, nil
}

// ResetTemplate resets a template to default (useful if user wants to restore)
func (s *EmailTemplateService) ResetTemplate(ctx context.Context, templateKey string) error {
	// Default templates - could be loaded from config
	defaults := map[string]struct {
		Subject, Body string
	}{
		"welcome_user": {
			Subject: "Welcome to {{org_name}}!",
			Body:    "<h1>مرحباً {{user_name}}</h1><p>تم إنشاء حسابك بنجاح.</p>",
		},
		"2fa_code": {
			Subject: "Your verification code: {{code}}",
			Body:    "<h1>رمز التحقق</h1><p>رمزك هو: <strong>{{code}}</strong></p><p>صالح لمدة 10 دقائق.</p>",
		},
	}

	def, ok := defaults[templateKey]
	if !ok {
		return fmt.Errorf("no default template for key: %s", templateKey)
	}

	query := `
		UPDATE email_templates
		SET subject = $1, body_html = $2, updated_at = NOW()
		WHERE template_key = $3
	`

	_, err := s.db.ExecContext(ctx, query, def.Subject, def.Body, templateKey)
	return err
}
