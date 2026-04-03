package domain

import (
	"context"
	"time"
)

type Question struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	UserID    uint64    `json:"user_id"`
	Username  string    `json:"username" gorm:"-"` // Virtual field for UI
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status" gorm:"size:50;not null;default:'open';index:idx_question_status"` // open, resolved
	ViewCount uint32    `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Answers   []Answer  `json:"answers,omitempty" gorm:"foreignKey:QuestionID"`
}

type Answer struct {
	ID           uint64    `json:"id" gorm:"primaryKey"`
	QuestionID   uint64    `json:"question_id"`
	UserID       uint64    `json:"user_id"`
	Username     string    `json:"username" gorm:"-"`
	Content      string    `json:"content"`
	BestPractice string    `json:"best_practice"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type QnaNotification struct {
	ID         uint64    `json:"id" gorm:"primaryKey"`
	UserID     uint64    `json:"user_id"`
	QuestionID uint64    `json:"question_id"`
	Message    string    `json:"message"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type QnaRepository interface {
	CreateQuestion(ctx context.Context, q *Question) error
	GetQuestions(ctx context.Context, search string, status string, limit int, offset int) ([]Question, int64, error)
	GetQuestionByID(ctx context.Context, id uint64) (*Question, error)
	UpdateQuestion(ctx context.Context, q *Question) error
	DeleteQuestion(ctx context.Context, id uint64) error
	IncrementViewCount(ctx context.Context, id uint64) error

	CreateAnswer(ctx context.Context, a *Answer) error
	GetAnswersByQuestionID(ctx context.Context, questionID uint64) ([]Answer, error)

	CreateNotification(ctx context.Context, n *QnaNotification) error
	GetNotificationsByUserID(ctx context.Context, userID uint64) ([]QnaNotification, error)
	MarkNotificationAsRead(ctx context.Context, id uint64) error
}

type QnaUsecase interface {
	AskQuestion(ctx context.Context, q *Question) error
	ListQuestions(ctx context.Context, search string, status string, page int) ([]Question, int64, error)
	GetThread(ctx context.Context, id uint64) (*Question, error)
	AnswerQuestion(ctx context.Context, a *Answer) error
	ResolveQuestion(ctx context.Context, id uint64, userID uint64) error
	DeleteQuestion(ctx context.Context, id uint64, role string) error
	GetFaqs(ctx context.Context, limit int) ([]Question, error)

	GetMyNotifications(ctx context.Context, userID uint64) ([]QnaNotification, error)
}
