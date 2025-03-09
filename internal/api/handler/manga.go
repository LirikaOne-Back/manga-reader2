package handler

import (
	"encoding/json"
	"manga-reader2/internal/api/response"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/usecase"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// MangaHandler обработчик запросов для API манги
type MangaHandler struct {
	mangaUseCase usecase.MangaUseCase
	log          logger.Logger
}

// NewMangaHandler создает новый экземпляр MangaHandler
func NewMangaHandler(mangaUseCase usecase.MangaUseCase, log logger.Logger) *MangaHandler {
	return &MangaHandler{
		mangaUseCase: mangaUseCase,
		log:          log,
	}
}

// List обрабатывает запрос на получение списка манги
// @Summary      Список манги
// @Description  Получить список всей манги с фильтрацией и пагинацией
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        title    query     string  false  "Фильтр по названию"
// @Param        status   query     string  false  "Фильтр по статусу (ongoing, completed, hiatus)"
// @Param        genres   query     string  false  "Фильтр по жанрам (через запятую)"
// @Param        limit    query     int     false  "Лимит результатов"
// @Param        offset   query     int     false  "Смещение результатов"
// @Success      200      {object}  response.Response{data=[]entity.Manga}
// @Failure      400      {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500      {object}  response.Response{error=errors.ErrorResponse}
// @Router       /manga [get]
func (h *MangaHandler) List(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	status := r.URL.Query().Get("status")
	genresStr := r.URL.Query().Get("genres")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var genres []string
	if genresStr != "" {
		genres = parseGenres(genresStr)
	}

	filter := entity.MangaFilter{
		Title:  title,
		Status: status,
		Genres: genres,
		Limit:  limit,
		Offset: offset,
	}

	manga, err := h.mangaUseCase.List(r.Context(), filter)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	meta := response.MetaPagination{
		Total:       len(manga),
		PerPage:     limit,
		CurrentPage: offset/limit + 1,
		LastPage:    (len(manga) + limit - 1) / limit,
	}

	response.SuccessWithMeta(w, http.StatusOK, manga, meta)
}

// GetByID обрабатывает запрос на получение манги по ID
// @Summary      Получить мангу
// @Description  Получить детальную информацию о манге по ID
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID манги"
// @Success      200  {object}  response.Response{data=entity.Manga}
// @Failure      404  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500  {object}  response.Response{error=errors.ErrorResponse}
// @Router       /manga/{id} [get]
func (h *MangaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Некорректный ID", err))
		return
	}

	manga, err := h.mangaUseCase.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.Success(w, http.StatusOK, manga)
}

// Create обрабатывает запрос на создание новой манги
// @Summary      Создать мангу
// @Description  Создать новую мангу
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        manga  body      entity.Manga  true  "Данные манги"
// @Success      201    {object}  response.Response{data=entity.Manga}
// @Failure      400    {object}  response.Response{error=errors.ErrorResponse}
// @Failure      401    {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500    {object}  response.Response{error=errors.ErrorResponse}
// @Security     Bearer
// @Router       /manga [post]
func (h *MangaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var manga entity.Manga
	if err := json.NewDecoder(r.Body).Decode(&manga); err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Ошибка парсинга JSON", err))
		return
	}

	createdManga, err := h.mangaUseCase.Create(r.Context(), &manga)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.Success(w, http.StatusCreated, createdManga)
}

// Update обрабатывает запрос на обновление манги
// @Summary      Обновить мангу
// @Description  Обновить существующую мангу
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        id     path      int           true  "ID манги"
// @Param        manga  body      entity.Manga  true  "Новые данные манги"
// @Success      200    {object}  response.Response{data=entity.Manga}
// @Failure      400    {object}  response.Response{error=errors.ErrorResponse}
// @Failure      401    {object}  response.Response{error=errors.ErrorResponse}
// @Failure      404    {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500    {object}  response.Response{error=errors.ErrorResponse}
// @Security     Bearer
// @Router       /manga/{id} [put]
func (h *MangaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Некорректный ID", err))
		return
	}

	var manga entity.Manga
	if err := json.NewDecoder(r.Body).Decode(&manga); err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Ошибка парсинга JSON", err))
		return
	}

	manga.ID = id

	updatedManga, err := h.mangaUseCase.Update(r.Context(), &manga)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.Success(w, http.StatusOK, updatedManga)
}

// Delete обрабатывает запрос на удаление манги
// @Summary      Удалить мангу
// @Description  Удалить мангу по ID
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID манги"
// @Success      204  {object}  nil
// @Failure      400  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      401  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      404  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500  {object}  response.Response{error=errors.ErrorResponse}
// @Security     Bearer
// @Router       /manga/{id} [delete]
func (h *MangaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Некорректный ID", err))
		return
	}

	if err = h.mangaUseCase.Delete(r.Context(), id); err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.NoContent(w)
}

// GetChapters обрабатывает запрос на получение глав манги
// @Summary      Получить главы манги
// @Description  Получить список всех глав манги
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID манги"
// @Success      200  {object}  response.Response{data=[]entity.Chapter}
// @Failure      400  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      404  {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500  {object}  response.Response{error=errors.ErrorResponse}
// @Router       /manga/{id}/chapters [get]
func (h *MangaHandler) GetChapters(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, h.log, errors.NewBadRequestError("Некорректный ID", err))
		return
	}

	chapters, err := h.mangaUseCase.GetChapters(r.Context(), id)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.Success(w, http.StatusOK, chapters)
}

// GetPopular обрабатывает запрос на получение популярной манги
// @Summary      Получить популярную мангу
// @Description  Получить список популярной манги
// @Tags         manga
// @Accept       json
// @Produce      json
// @Param        period  query     string  false  "Период статистики (daily, weekly, monthly, all_time)"
// @Param        limit   query     int     false  "Лимит результатов"
// @Success      200     {object}  response.Response{data=[]entity.MangaStat}
// @Failure      400     {object}  response.Response{error=errors.ErrorResponse}
// @Failure      500     {object}  response.Response{error=errors.ErrorResponse}
// @Router       /manga/popular [get]
func (h *MangaHandler) GetPopular(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	limitStr := r.URL.Query().Get("limit")

	limit := 10

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	statsPeriod := entity.StatsPeriodAllTime
	switch period {
	case "daily":
		statsPeriod = entity.StatsPeriodDaily
	case "weekly":
		statsPeriod = entity.StatsPeriodWeekly
	case "monthly":
		statsPeriod = entity.StatsPeriodMonthly
	}

	popular, err := h.mangaUseCase.GetPopular(r.Context(), statsPeriod, limit)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.Success(w, http.StatusOK, popular)
}

// parseGenres разбивает строку с жанрами на список
func parseGenres(genresStr string) []string {
	return nil // Заглушка, будет реализована позже
}
