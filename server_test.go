package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

var (
	testJson = `{"one":"val1"}`
)

func TestHandler(t *testing.T) {
	bb := bytes.NewBufferString(testJson)
	router := httprouter.New()
	router.PUT("/:topic", putTopicHandler)
	req, err := http.NewRequest("PUT", "/testTopic", bb)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Result().StatusCode != http.StatusCreated {
		t.Errorf("want (http status code): %d got: %d", http.StatusCreated, rr.Result().StatusCode)
	}
}

func TestTopicListHandler(t *testing.T) {
	router := httprouter.New()
	router.GET("/topics", topicListHandler)
	req, err := http.NewRequest("GET", "/topics", nil)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Body.Len() == 0 {
		t.Error("rr.Body can not be empty")
	}
	var got []string
	err = json.Unmarshal(rr.Body.Bytes(), &got)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTopicHandler(t *testing.T) {
	router := httprouter.New()
	router.GET("/topic/:topic/:offset", topicGetHandler)
	req, err := http.NewRequest("GET", "/topic/testTopic/0", nil)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Body.Len() == 0 {
		t.Error("rr.Body can not be empty")
	}
	if rr.Body.String() != testJson {
		t.Errorf("\nwant:\n\t'%+v'\ngot:\n\t'%+v'", testJson, rr.Body.String())
	}
}

func TestSubsribeHandler(t *testing.T) {
	router := httprouter.New()
	router.POST("/topic/:topic/subscribe", subscribeHandler)
	payload := bytes.NewBufferString(`{"endpoint":"http://test.com"}`)
	req, err := http.NewRequest("POST", "/topic/testTopic/subscribe", payload)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Result().StatusCode != http.StatusCreated {
		t.Errorf("\nwant (status code):\n\t'%+v'\ngot:\n\t'%+v'", http.StatusCreated, rr.Result().StatusCode)
	}
	want := make(map[string]string)
	got := make(map[string]string)
	want["subscriber_id"] = "test"
	if err = json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Error(err)
	}
	_, ok := got["subscriber_id"]
	if !ok {
		t.Errorf("\nwant:\n\t'%+v'\ngot:\n\t'%+v'", "subscriber_id", got)
	}
	if want["subscriber_id"] == got["subscriber_id"] {
		t.Errorf("\nwant:\n\t'%+v'\ngot:\n\t'%+v'", want, got)
	}
}
