package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	port          = ":8080"
	topicExt      = ".topic"
	topicDir      = "topcis/"
	subscriberDir = "subscribers/"
)

func main() {
	router := httprouter.New()
	router.PUT("/topic/:topic", putTopicHandler)
	router.GET("/topic", topicListHandler)
	router.GET("/topic/:topic/:offset", topicGetHandler)
	router.POST("/topic/:topic/subscribe", subscribeHandler)
	log.Fatal(http.ListenAndServe(port, &Server{router}))
}

type Server struct {
	r *httprouter.Router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	s.r.ServeHTTP(w, r)
}

var (
	writeFiles = make(map[string]*os.File)
	readFiles  = make(map[string]*os.File)
)

func putTopicHandler(w http.ResponseWriter, r *http.Request, s httprouter.Params) {
	var err error
	var f *os.File
	topic := s.ByName("topic")
	err = os.MkdirAll(topicDir, 0755)
	f, ok := writeFiles[topic]
	if !ok {
		f, err = os.OpenFile(topicDir+topic+topicExt, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err500(w, err) {
			return
		}
		writeFiles[topic] = f
	}
	m := make(map[string]int64)
	m["offset"], err = f.Seek(0, os.SEEK_CUR)
	_, err = io.Copy(f, r.Body)
	if err500(w, err) {
		return
	}
	defer func() {
		f.Write([]byte{'\n'})
		r.Body.Close()
		return
	}()
	m["next"], err = f.Seek(0, os.SEEK_CUR)
	if err500(w, err) {
		return
	}
	loc := fmt.Sprintf("%s/%d", r.RequestURI, m["offset"])
	w.Header().Set("Location", loc)
	w.Header().Set("Content-Type", "json/application")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(m)
	if err500(w, err) {
		return
	}
}

func topicListHandler(w http.ResponseWriter, r *http.Request, s httprouter.Params) {
	var key []string
	for k := range writeFiles {
		key = append(key, k)
	}
	b, err := json.Marshal(key)
	_, err = w.Write(b)
	if err500(w, err) {
		return
	}
}

func topicGetHandler(w http.ResponseWriter, r *http.Request, s httprouter.Params) {
	var err error
	topic := s.ByName("topic")
	f, ok := readFiles[topic]
	if !ok {
		f, err = os.OpenFile(topicDir+topic+topicExt, os.O_RDONLY, 0666)
		if err500(w, err) {
			return
		}
		readFiles[topic] = f
	}
	var offset int64
	if s.ByName("offset") == "" {
		offset = 0
	} else {
		offset, err = strconv.ParseInt(s.ByName("offset"), 0, 64)
	}
	if err500(w, err) {
		return
	}
	_, err = f.Seek(offset, os.SEEK_SET)
	if err500(w, err) {
		return
	}
	switch r.URL.Query().Get("cmd") {
	default:
		if err500(w, readOne(w, f)) {
			return
		}
	case "follow":
		follow(w, f)
	}
}

func readOne(w http.ResponseWriter, f io.Reader) (err error) {
	b, _, err := bufio.NewReader(f).ReadLine()
	if err != nil {
		return
	}
	_, err = w.Write(b)
	return
}

func follow(w http.ResponseWriter, f io.ReadCloser) (err error) {
	_, err = io.Copy(w, f)
	defer f.Close()
	return
}

type sub struct {
	url string
}

var subs = make(map[string][]sub)

func subscribeHandler(w http.ResponseWriter, r *http.Request, s httprouter.Params) {
	var f *os.File
	var err error
	topic := s.ByName("topic")
	err = os.MkdirAll(subscriberDir+topic, 0755)
	if err500(w, err) {
		return
	}
	msg := make(map[string]string)
	msg["subscriber_id"] = id(20)
	f, err = os.OpenFile(subscriberDir+topic+"/"+msg["id"]+".sub", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err500(w, err) {
		return
	}
	_, err = io.Copy(f, r.Body)
	if err500(w, err) {
		return
	}
	defer r.Body.Close()
	defer f.Close()
	f.Write([]byte{'\n'})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func err500(w http.ResponseWriter, err error) (b bool) {
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return true
	}
	return false
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func id(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
