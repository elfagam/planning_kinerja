package http

import (
	"e-plan-ai/internal/modules/qna/domain"
	"e-plan-ai/internal/shared/response"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type QnaHandler struct {
	usecase domain.QnaUsecase
}

func NewQnaHandler(u domain.QnaUsecase) *QnaHandler {
	return &QnaHandler{usecase: u}
}

func (h *QnaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	qna := rg.Group("/qna")
	{
		qna.GET("/questions", h.listQuestions)
		qna.GET("/questions/:id", h.getThread)
		qna.POST("/questions", h.askQuestion)
		qna.POST("/questions/:id/answers", h.answerQuestion)
		qna.POST("/questions/:id/resolve", h.resolveQuestion)
		qna.DELETE("/questions/:id", h.deleteQuestion)
		qna.GET("/faq", h.getFaqs)
		qna.GET("/notifications", h.getNotifications)
	}
}

func (h *QnaHandler) listQuestions(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	questions, total, err := h.usecase.ListQuestions(c.Request.Context(), search, status, page)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"items": questions,
		"total": total,
		"page":  page,
	})
}

func (h *QnaHandler) askQuestion(c *gin.Context) {
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from middleware context
	userIDRaw, _ := c.Get("auth.user_id")
	userID, _ := userIDRaw.(uint64)

	q := &domain.Question{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.usecase.AskQuestion(c.Request.Context(), q); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, q)
}

func (h *QnaHandler) getThread(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	q, err := h.usecase.GetThread(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Question not found")
		return
	}
	response.Success(c, q)
}

func (h *QnaHandler) answerQuestion(c *gin.Context) {
	questionID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userIDRaw, _ := c.Get("auth.user_id")
	userID, _ := userIDRaw.(uint64)

	a := &domain.Answer{
		QuestionID: questionID,
		UserID:     userID,
		Content:    req.Content,
	}

	if err := h.usecase.AnswerQuestion(c.Request.Context(), a); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, a)
}

func (h *QnaHandler) resolveQuestion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	userIDRaw, _ := c.Get("auth.user_id")
	userID, _ := userIDRaw.(uint64)

	if err := h.usecase.ResolveQuestion(c.Request.Context(), id, userID); err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, gin.H{"status": "resolved"})
}

func (h *QnaHandler) getFaqs(c *gin.Context) {
	faqs, err := h.usecase.GetFaqs(c.Request.Context(), 5)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, faqs)
}

func (h *QnaHandler) deleteQuestion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roleRaw, _ := c.Get("auth.role")
	role, _ := roleRaw.(string)

	if err := h.usecase.DeleteQuestion(c.Request.Context(), id, role); err != nil {
		response.Error(c, http.StatusForbidden, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Pertanyaan berhasil dihapus"})
}

func (h *QnaHandler) getNotifications(c *gin.Context) {
	userIDRaw, _ := c.Get("auth.user_id")
	userID, _ := userIDRaw.(uint64)

	notifs, err := h.usecase.GetMyNotifications(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, notifs)
}
