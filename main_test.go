package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"fmt"
	"strconv"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {

	for city := range cafeList {
		
		requests := []struct {
			count int
			want  int
		}{
			{0, 0},
			{1, 1},
			{2, 2},
			{100, min(len(cafeList[city]), 100)},
		}

		for _, request := range requests {
			url := fmt.Sprintf("/?city=%s&count=%s", city, strconv.Itoa(request.count))
			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(mainHandle)
			handler.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusOK, recorder.Code)

			responseBody := strings.TrimSpace(recorder.Body.String())

			var cafes []string
			
			if responseBody == "" {
				cafes = []string{}
			} else {
				cafes = strings.Split(responseBody, ",")
			}

			assert.Equal(t, request.want, len(cafes), 
			"Неверное количество кафе для count=%d", request.count)
		}
	}
}

func TestCafeSearch(t *testing.T) {

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, request := range requests {
		url := fmt.Sprintf("/?city=moscow&search=%s", request.search)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(mainHandle)
		handler.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)

		responseBody := strings.TrimSpace(recorder.Body.String())

		var cafes []string
		
		if responseBody == "" {
			cafes = []string{}
		} else {
			cafes = strings.Split(responseBody, ",")
		}

		searchString := strings.ToLower(request.search)
		for _, cafe := range cafes {
			cafeTrimmed := strings.ToLower(strings.TrimSpace(cafe))
			assert.True(t, strings.Contains(cafeTrimmed, searchString),
			"Название кафе %s не содержит строку %s", cafeTrimmed, searchString)
		}

		assert.Equal(t, request.wantCount, len(cafes), 
		"Неверное количество кафе для search=%s", request.search)
	}
}