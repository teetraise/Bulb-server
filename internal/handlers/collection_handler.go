package handlers

import (
	"net/http"
	"strconv"

	"github.com/KoLili12/bulb-server/internal/middleware"
	"github.com/KoLili12/bulb-server/internal/models"
	"github.com/KoLili12/bulb-server/internal/services"
	"github.com/gin-gonic/gin"
)

// CollectionRequest представляет структуру запроса для создания/обновления коллекции
type CollectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
}

// CreateCollectionWithActionsRequest представляет запрос на создание коллекции с карточками
type CreateCollectionWithActionsRequest struct {
	Name        string                `json:"name" binding:"required"`
	Description string                `json:"description"`
	ImageURL    string                `json:"imageUrl"`
	Actions     []CreateActionRequest `json:"actions"`
}

// CreateActionRequest представляет структуру запроса для создания действия с типом
type CreateActionRequest struct {
	Text  string `json:"text" binding:"required"`
	Type  string `json:"type" binding:"required"` // "truth" или "dare"
	Order int    `json:"order"`
}

// ActionRequest представляет структуру запроса для добавления действия
type ActionRequest struct {
	Text  string `json:"text" binding:"required"`
	Order int    `json:"order"`
}

// CollectionResponse представляет структуру ответа с данными коллекции
type CollectionResponse struct {
	ID          uint                      `json:"id"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	ImageURL    string                    `json:"imageUrl"`
	UserID      uint                      `json:"userId"`
	PlayCount   int                       `json:"playCount"`
	Actions     []ActionResponseWithType  `json:"actions,omitempty"`
	CreatedAt   string                    `json:"createdAt"`
}

// ActionResponse представляет структуру ответа с данными действия
type ActionResponse struct {
	ID    uint   `json:"id"`
	Text  string `json:"text"`
	Order int    `json:"order"`
}

// ActionResponseWithType представляет структуру ответа с данными действия включая тип
type ActionResponseWithType struct {
	ID    uint   `json:"id"`
	Text  string `json:"text"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

// CollectionStatsResponse представляет статистику коллекции
type CollectionStatsResponse struct {
	TotalActions int `json:"totalActions"`
	TruthCount   int `json:"truthCount"`
	DareCount    int `json:"dareCount"`
}

// PaginationResponse представляет структуру ответа с пагинацией
type PaginationResponse struct {
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
	Items []CollectionResponse `json:"items"`
}

// CollectionHandler обрабатывает запросы, связанные с коллекциями
type CollectionHandler struct {
	collectionService services.CollectionService
}

// NewCollectionHandler создает новый обработчик коллекций
func NewCollectionHandler(collectionService services.CollectionService) *CollectionHandler {
	return &CollectionHandler{
		collectionService: collectionService,
	}
}

// Create обрабатывает запрос на создание новой коллекции (без карточек)
func (h *CollectionHandler) Create(c *gin.Context) {
	var req CollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	collection := &models.Collection{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		UserID:      userID,
	}

	if err := h.collectionService.Create(collection); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Message: "Collection created successfully"})
}

// CreateWithActions обрабатывает запрос на создание коллекции с карточками
func (h *CollectionHandler) CreateWithActions(c *gin.Context) {
	var req CreateCollectionWithActionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Создаем коллекцию
	collection := &models.Collection{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		UserID:      userID,
	}

	// Преобразуем действия из запроса в модель
	var actions []*models.Action
	for _, actionReq := range req.Actions {
		// Валидируем тип действия
		var actionType models.ActionType
		switch actionReq.Type {
		case "truth":
			actionType = models.ActionTypeTruth
		case "dare":
			actionType = models.ActionTypeDare
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid action type: " + actionReq.Type})
			return
		}

		action := &models.Action{
			Text:  actionReq.Text,
			Type:  actionType,
			Order: actionReq.Order,
		}
		actions = append(actions, action)
	}

	// Создаем коллекцию с действиями
	if err := h.collectionService.CreateWithActions(collection, actions); err != nil {
		if err == services.ErrInvalidActionType {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid action type"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create collection with actions"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Message: "Collection with actions created successfully"})
}

// GetByID обрабатывает запрос на получение коллекции по ID
func (h *CollectionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	collection, err := h.collectionService.GetByID(uint(id))
	if err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get collection"})
		return
	}

	// Преобразуем коллекцию в ответ
	actions := make([]ActionResponseWithType, 0, len(collection.Actions))
	for _, action := range collection.Actions {
		actions = append(actions, ActionResponseWithType{
			ID:    action.ID,
			Text:  action.Text,
			Type:  string(action.Type),
			Order: action.Order,
		})
	}

	response := CollectionResponse{
		ID:          collection.ID,
		Name:        collection.Name,
		Description: collection.Description,
		ImageURL:    collection.ImageURL,
		UserID:      collection.UserID,
		PlayCount:   collection.PlayCount,
		Actions:     actions,
		CreatedAt:   collection.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// Update обрабатывает запрос на обновление коллекции
func (h *CollectionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	var req CollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	collection := &models.Collection{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	if err := h.collectionService.Update(collection, userID); err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		if err == services.ErrNotCollectionOwner {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not the owner of this collection"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update collection"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Collection updated successfully"})
}

// Delete обрабатывает запрос на удаление коллекции
func (h *CollectionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.collectionService.Delete(uint(id), userID); err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		if err == services.ErrNotCollectionOwner {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not the owner of this collection"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete collection"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Collection deleted successfully"})
}

// List обрабатывает запрос на получение списка коллекций
func (h *CollectionHandler) List(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	collections, total, err := h.collectionService.List(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get collections"})
		return
	}

	// Преобразуем коллекции в ответ
	items := make([]CollectionResponse, 0, len(collections))
	for _, collection := range collections {
		items = append(items, CollectionResponse{
			ID:          collection.ID,
			Name:        collection.Name,
			Description: collection.Description,
			ImageURL:    collection.ImageURL,
			UserID:      collection.UserID,
			PlayCount:   collection.PlayCount,
			CreatedAt:   collection.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, PaginationResponse{
		Total: total,
		Page:  page,
		Size:  size,
		Items: items,
	})
}

// GetTrending обрабатывает запрос на получение популярных коллекций
func (h *CollectionHandler) GetTrending(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	collections, err := h.collectionService.GetTrending(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get trending collections"})
		return
	}

	// Преобразуем коллекции в ответ
	items := make([]CollectionResponse, 0, len(collections))
	for _, collection := range collections {
		// Получаем статистику для каждой коллекции
		truthCount, dareCount, total, _ := h.collectionService.GetActionCounts(collection.ID)
		
		items = append(items, CollectionResponse{
			ID:          collection.ID,
			Name:        collection.Name,
			Description: collection.Description,
			ImageURL:    collection.ImageURL,
			UserID:      collection.UserID,
			PlayCount:   collection.PlayCount,
			CreatedAt:   collection.CreatedAt.Format("2006-01-02T15:04:05Z"),
			// Можно добавить статистику в ответ если нужно
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// GetUserCollections обрабатывает запрос на получение коллекций пользователя
func (h *CollectionHandler) GetUserCollections(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	collections, err := h.collectionService.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get user collections"})
		return
	}

	// Преобразуем коллекции в ответ
	items := make([]CollectionResponse, 0, len(collections))
	for _, collection := range collections {
		// Получаем количество действий
		_, _, total, _ := h.collectionService.GetActionCounts(collection.ID)
		
		items = append(items, CollectionResponse{
			ID:          collection.ID,
			Name:        collection.Name,
			Description: collection.Description,
			ImageURL:    collection.ImageURL,
			UserID:      collection.UserID,
			PlayCount:   collection.PlayCount,
			CreatedAt:   collection.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// GetCollectionStats возвращает статистику коллекции
func (h *CollectionHandler) GetCollectionStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	truthCount, dareCount, total, err := h.collectionService.GetActionCounts(uint(id))
	if err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get collection stats"})
		return
	}

	stats := CollectionStatsResponse{
		TotalActions: total,
		TruthCount:   truthCount,
		DareCount:    dareCount,
	}

	c.JSON(http.StatusOK, stats)
}

// AddAction обрабатывает запрос на добавление действия в коллекцию
func (h *CollectionHandler) AddAction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	var req CreateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	// Валидируем тип действия
	var actionType models.ActionType
	switch req.Type {
	case "truth":
		actionType = models.ActionTypeTruth
	case "dare":
		actionType = models.ActionTypeDare
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid action type: " + req.Type})
		return
	}

	action := &models.Action{
		Text:  req.Text,
		Type:  actionType,
		Order: req.Order,
	}

	if err := h.collectionService.AddAction(uint(id), action); err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		if err == services.ErrInvalidActionType {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid action type"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add action"})
		return
	}

	c.JSON(http.StatusCreated, SuccessResponse{Message: "Action added successfully"})
}

// GetActions обрабатывает запрос на получение действий коллекции
func (h *CollectionHandler) GetActions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid collection ID"})
		return
	}

	actions, err := h.collectionService.GetActions(uint(id))
	if err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get actions"})
		return
	}

	// Преобразуем действия в ответ с типами
	items := make([]ActionResponseWithType, 0, len(actions))
	for _, action := range actions {
		items = append(items, ActionResponseWithType{
			ID:    action.ID,
			Text:  action.Text,
			Type:  string(action.Type),
			Order: action.Order,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// RemoveAction обрабатывает запрос на удаление действия
func (h *CollectionHandler) RemoveAction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid action ID"})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
		return
	}

	if err := h.collectionService.RemoveAction(uint(id), userID); err != nil {
		if err == services.ErrCollectionNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Collection not found"})
			return
		}
		if err == services.ErrNotCollectionOwner {
			c.JSON(http.StatusForbidden, ErrorResponse{Error: "You are not the owner of this collection"})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove action"})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{Message: "Action removed successfully"})
}
