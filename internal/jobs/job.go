package jobs

import (
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Job struct {
	ID           string         `json:"id"`
	Status       Status         `json:"status"`
	InputFormat  formats.Format `json:"input_format"`
	OutputFormat formats.Format `json:"output_format"`
	CreatedAt    time.Time      `json:"created_at"`
	StartedAt    *time.Time     `json:"started_at,omitempty"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	Error        string         `json:"error,omitempty"`
}
