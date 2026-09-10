package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("idempotency key reused with different input")
	ErrInstruction = errors.New("unknown instruction")
)

type Band struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Identity string `json:"identity"`
}

type Character struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
	Profile string `json:"profile"`
}

type Instruction struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type MeetingInput struct {
	Topic          string   `json:"topic"`
	InstructionIDs []string `json:"instruction_ids"`
}

type Meeting struct {
	ID            string        `json:"id"`
	JobID         string        `json:"job_id"`
	Topic         string        `json:"topic"`
	Status        string        `json:"status"`
	BlockedReason string        `json:"blocked_reason,omitempty"`
	Band          Band          `json:"band_snapshot"`
	Participants  []Character   `json:"participants"`
	Instructions  []Instruction `json:"instructions"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type MeetingAccepted struct {
	MeetingID string `json:"meeting_id"`
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
}

type Job struct {
	ID            string    `json:"id"`
	MeetingID     string    `json:"meeting_id"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	BlockedReason string    `json:"blocked_reason,omitempty"`
	Attempts      int       `json:"attempts"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Message struct {
	ID          string    `json:"id"`
	CharacterID string    `json:"character_id"`
	Sequence    int       `json:"sequence"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}
