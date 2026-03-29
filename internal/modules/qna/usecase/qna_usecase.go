package usecase

import (
	"context"
	"e-plan-ai/internal/modules/qna/domain"
	"fmt"
	"strings"
)

type QnaUsecaseImpl struct {
	repo domain.QnaRepository
}

func NewQnaUsecase(repo domain.QnaRepository) domain.QnaUsecase {
	return &QnaUsecaseImpl{repo: repo}
}

func (u *QnaUsecaseImpl) AskQuestion(ctx context.Context, q *domain.Question) error {
	q.Status = "open"
	q.ViewCount = 0
	return u.repo.CreateQuestion(ctx, q)
}

func (u *QnaUsecaseImpl) ListQuestions(ctx context.Context, search string, status string, page int) ([]domain.Question, int64, error) {
	limit := 10
	offset := (page - 1) * limit
	return u.repo.GetQuestions(ctx, search, status, limit, offset)
}

func (u *QnaUsecaseImpl) GetThread(ctx context.Context, id uint64) (*domain.Question, error) {
	// Increment view count
	_ = u.repo.IncrementViewCount(ctx, id)
	return u.repo.GetQuestionByID(ctx, id)
}

func (u *QnaUsecaseImpl) AnswerQuestion(ctx context.Context, a *domain.Answer) error {
	err := u.repo.CreateAnswer(ctx, a)
	if err != nil {
		return err
	}

	// Notify question owner
	q, err := u.repo.GetQuestionByID(ctx, a.QuestionID)
	if err == nil && q.UserID != a.UserID {
		notification := &domain.QnaNotification{
			UserID:     q.UserID,
			QuestionID: q.ID,
			Message:    "Pertanyaan Anda mendapat jawaban baru.",
			IsRead:     false,
		}
		_ = u.repo.CreateNotification(ctx, notification)
	}
	return nil
}

func (u *QnaUsecaseImpl) ResolveQuestion(ctx context.Context, id uint64, userID uint64) error {
	q, err := u.repo.GetQuestionByID(ctx, id)
	if err != nil {
		return err
	}
	// Only owner or admin (simulated) can resolve
	if q.UserID != userID {
		return fmt.Errorf("permission denied")
	}
	q.Status = "resolved"
	return u.repo.UpdateQuestion(ctx, q)
}

func (u *QnaUsecaseImpl) GetFaqs(ctx context.Context, limit int) ([]domain.Question, error) {
	// Get top viewed questions as FAQ
	questions, _, err := u.repo.GetQuestions(ctx, "", "", limit, 0)
	return questions, err
}

func (u *QnaUsecaseImpl) DeleteQuestion(ctx context.Context, id uint64, role string) error {
	role = strings.ToUpper(strings.TrimSpace(role))
	if role != "ADMIN" && role != "PIMPINAN" {
		return fmt.Errorf("hanya Admin yang dapat menghapus pertanyaan")
	}
	return u.repo.DeleteQuestion(ctx, id)
}

func (u *QnaUsecaseImpl) GetMyNotifications(ctx context.Context, userID uint64) ([]domain.QnaNotification, error) {
	return u.repo.GetNotificationsByUserID(ctx, userID)
}
