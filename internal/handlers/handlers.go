package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"tz/internal/repository"

	"github.com/google/uuid"
)

// @Summary      Создание подписки или получение списка
// @Description  POST - создание новой подписки, GET - список всех подписок
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Success      200 {array} repository.Subscription
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      405 {object} map[string]string
// @Router       /subscriptions [post]
// @Router       /subscriptions [get]
func Subscriptions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		CreateSubscription(w, r)
	case "GET":
		ListSubscriptions(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// @Summary      Работа с конкретной подпиской
// @Description  GET - получение, PUT - обновление, DELETE - удаление
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID подписки"
// @Success      200  {object}  repository.Subscription
// @Success      200  {object}  map[string]string
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      405  {object}  map[string]string
// @Router       /subscriptions/{id} [get]
// @Router       /subscriptions/{id} [put]
// @Router       /subscriptions/{id} [delete]
func SubscriptionsID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка URL при взятии id"})
		return
	}
	idStr := parts[len(parts)-1]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID не может быть пустым"})
		return
	}

	subID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Ошибка конвертации id в число: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный тип id, должно быть число"})
		return
	}

	switch r.Method {
	case "GET":
		GetSubscription(w, r, subID)
	case "PUT":
		UpdateSubscription(w, r, subID)
	case "DELETE":
		DeleteSubscription(w, r, subID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
	}
}

// @Summary      Создание подписки
// @Description  Создает новую подписку
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        request body repository.CreateSubscriptionRequest true "Данные подписки"
// @Success      201 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Router       /subscriptions [post]
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req repository.CreateSubscriptionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат запроса"})
		return
	}

	if req.ServiceName == "" || req.Price <= 0 || req.UserID == uuid.Nil || req.StartDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "пустые поля недопустимы"})
		return
	}

	err = repository.CreateSubscription(req)
	if err != nil {
		log.Printf("Ошибка создания подписки, %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка создания подписки"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Подписка создана"})
}

// @Summary      Получение подписки
// @Description  Возвращает подписку по ID
// @Tags         Subscriptions
// @Produce      json
// @Param        id path int true "ID подписки"
// @Success      200 {object} repository.Subscription
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [get]
func GetSubscription(w http.ResponseWriter, r *http.Request, subID int) {
	sub, err := repository.GetSub(subID)
	if err != nil {
		log.Printf("Ошибка получения подписки: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Подписка не найдена"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
}

// @Summary      Обновление подписки
// @Description  Обновляет существующую подписку
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        id path int true "ID подписки"
// @Param        request body repository.UpdateRequest true "Данные для обновления"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [put]
func UpdateSubscription(w http.ResponseWriter, r *http.Request, subID int) {
	var req repository.UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат JSON"})
		log.Printf("Ошибка Decode JSON")
		return
	}

	err = repository.UpdateSubscription(subID, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка обновления подписки"})
		log.Printf("Ошибка обновления подписки: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Подписка обновлена"})
}

// @Summary      Удаление подписки
// @Description  Удаляет подписку по ID
// @Tags         Subscriptions
// @Param        id path int true "ID подписки"
// @Success      204
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /subscriptions/{id} [delete]
func DeleteSubscription(w http.ResponseWriter, r *http.Request, subID int) {
	err := repository.DeleteSubscription(subID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка удаления подписки"})
		log.Printf("Ошибка удаления подписки: %v", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary      Список подписок
// @Description  Возвращает список всех подписок с пагинацией
// @Tags         Subscriptions
// @Produce      json
// @Param        page query int false "Номер страницы" default(1)
// @Param        limit query int false "Количество на странице" default(10)
// @Param        user_id query string false "Фильтр по пользователю"
// @Param        service_name query string false "Фильтр по сервису"
// @Success      200 {object} map[string]interface{}
// @Failure      500 {object} map[string]string
// @Router       /subscriptions [get]
func ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.URL.Query().Get("user_id"))
	if err != nil {
		log.Printf("Ошибка конвертации string->UUID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id должен быть в формате UUID"})
		return
	}
	subs, err := repository.ListSubscriptions(userID)
	if err != nil {
		log.Printf("Ошибка получения списка подписок: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка получения списка подписок"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subs)
}

// @Summary      Сумма подписок
// @Description  Подсчитывает суммарную стоимость подписок за период с фильтрацией
// @Tags         Subscriptions
// @Produce      json
// @Param        start_date query string true "Дата начала (MM-YYYY)"
// @Param        end_date query string true "Дата окончания (MM-YYYY)"
// @Param        user_id query string false "Фильтр по пользователю"
// @Param        service_name query string false "Фильтр по сервису"
// @Success      200 {object} SumResponse
// @Failure      400 {object} map[string]string
// @Failure      405 {object} map[string]string
// @Router       /sumAllSub [get]
func Sum(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не поддерживается"})
		return
	}

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	//Для подсчета суммы за период end_date обязателен, иначе непонятно, за какой период считать сумму
	if startDate == "" || endDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "start_date и end_date обязательны"})
		return
	}

	sum, err := repository.GetSum(userID, serviceName, startDate, endDate)
	if err != nil {
		log.Printf("Ошибка подсчета суммарной стоимости всех подписок: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ошибка подсчета суммарной стоимости всех подписок"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sum)
}
