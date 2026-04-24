package api

import (
	"net/http"
	"errors"
	"fmt"
	"log"
	"time"
	"strconv"	
	"strings"
)

const TimeLayout = "20060102"
type dateRule func(current time.Time, parts []string, now time.Time) (time.Time, error)

var rules = map[string]dateRule{
	"y": func(current time.Time, _ []string, now time.Time) (time.Time, error) {
		for {
			current = current.AddDate(1, 0, 0)
			if current.After(now) {
				return current, nil
			}
		}
	},
	"d": func(current time.Time, parts []string, now time.Time) (time.Time, error) {
		if len(parts) < 2 {
			return time.Time{}, errors.New("пропущен параметр количества дней")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days < 1 || days > 400 {
			return time.Time{}, errors.New("неверное количество дней (1-400)")
		}
		for {
			current = current.AddDate(0, 0, days)
			if current.After(now) {
				return current, nil
			}
		}
	},
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("пустое правило повторения")
	}

	start, err := time.Parse(TimeLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("ошибка формата даты: %w", err)
	}

	parts := strings.Fields(repeat)
	ruleName := parts[0]
	calculator, ok := rules[ruleName]
	if !ok {
		log.Printf("NextDate: Неизвестное правило '%s'", ruleName)
		return "", errors.New("выбранный тип повторения пока не поддерживается")
	}

	log.Printf("NextDate: Расчет по правилу '%s' (начало: %s, сейчас: %s)", repeat, dstart, now.Format(TimeLayout))

	next, err := calculator(start, parts, now)
	if err != nil {
		return "", err
	}

	result := next.Format(TimeLayout)
	log.Printf("NextDate: Результат расчета — %s", result)
	return result, nil
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("API: Метод %s не разрешен для /api/nextdate", r.Method)
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	log.Printf("API: Запрос NextDate (now=%s, date=%s, repeat=%s)", nowParam, dateParam, repeatParam)

	now := time.Now()
	if nowParam != "" {
		var err error
		now, err = time.Parse(TimeLayout, nowParam)
		if err != nil {
			log.Printf("API: Ошибка формата параметра now: %v", err)
			http.Error(w, "неверный формат 'now'", http.StatusBadRequest)
			return
		}
	}

	res, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		log.Printf("API: Ошибка в NextDate: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(res)); err != nil {
		log.Printf("API: Ошибка записи ответа: %v", err)
	}
}