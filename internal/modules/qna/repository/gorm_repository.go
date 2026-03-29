package repository

import (
	"context"
	"e-plan-ai/internal/modules/qna/domain"
	"gorm.io/gorm"
)

type QnaGormRepository struct {
	db *gorm.DB
}

func NewQnaGormRepository(db *gorm.DB) *QnaGormRepository {
	return &QnaGormRepository{db: db}
}

func (r *QnaGormRepository) CreateQuestion(ctx context.Context, q *domain.Question) error {
	return r.db.WithContext(ctx).Create(q).Error
}

func (r *QnaGormRepository) GetQuestions(ctx context.Context, search string, status string, limit int, offset int) ([]domain.Question, int64, error) {
	var questions []domain.Question
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Question{})

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch usernames (Mock join simulation for simplicity, or we could join with users table if schema is stable)
	for i := range questions {
		questions[i].Username = "User #" + string(rune(questions[i].UserID)) // In real app, join with users table
	}

	return questions, total, nil
}

func (r *QnaGormRepository) GetQuestionByID(ctx context.Context, id uint64) (*domain.Question, error) {
	var q domain.Question
	err := r.db.WithContext(ctx).Preload("Answers").First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QnaGormRepository) UpdateQuestion(ctx context.Context, q *domain.Question) error {
	return r.db.WithContext(ctx).Save(q).Error
}

func (r *QnaGormRepository) DeleteQuestion(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Question{}, id).Error
}

func (r *QnaGormRepository) IncrementViewCount(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&domain.Question{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *QnaGormRepository) CreateAnswer(ctx context.Context, a *domain.Answer) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *QnaGormRepository) GetAnswersByQuestionID(ctx context.Context, questionID uint64) ([]domain.Answer, error) {
	var answers []domain.Answer
	err := r.db.WithContext(ctx).Where("question_id = ?", questionID).Order("created_at ASC").Find(&answers).Error
	return answers, err
}

func (r *QnaGormRepository) CreateNotification(ctx context.Context, n *domain.QnaNotification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *QnaGormRepository) GetNotificationsByUserID(ctx context.Context, userID uint64) ([]domain.QnaNotification, error) {
	var notifications []domain.QnaNotification
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_read = ?", userID, false).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *QnaGormRepository) MarkNotificationAsRead(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&domain.QnaNotification{}).Where("id = ?", id).Update("is_read", true).Error
}
