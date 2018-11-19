package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type MadlibServer struct {
	madlibs []MadlibTemplate
	ongoing map[string]*Madlib
	addr    string
	sync.RWMutex
}

type KeyResponse struct {
	Key string `json:"key"`
}

type NextRequest struct {
	Key string `json:"key"`
}

type NextResponse struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Prompt string `json:"prompt,omitempty"`
	Madlib string `json:"madlib,omitempty"`
	Done   bool   `json:"done"`
}

type AnswerRequest struct {
	Key    string `json:"key"`
	Answer string `json:"answer"`
}

// Document the API:
//  GET /new -> gets new Madlib instance key; this tracks state for a madlib in progress
//  GET /next -> retrieves the prompt. Returns same prompt until they POST the answer.
//     When there are no more prompts, this returns the filled out madlib.
//     Also has bool flag if it is done or not.
//  POST /answer -> answers the prompt with a string
func NewMadlibServer(addr string, madlibs []MadlibTemplate) *MadlibServer {
	return &MadlibServer{
		addr:    addr,
		madlibs: madlibs,
		ongoing: make(map[string]*Madlib),
	}
}

func (srv *MadlibServer) New(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var b = make([]byte, 16)
	n, err := rand.Read(b)
	if n != 16 || err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}
	key := base64.StdEncoding.EncodeToString(b)

	srv.Lock()
	srv.ongoing[key], err = srv.madlibs[rand.Intn(len(srv.madlibs))].NewMadlib()
	srv.Unlock()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	resp := KeyResponse{
		Key: key,
	}
	log.Printf("Creating new key %s for client %s", key, r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(resp)
	return
}

func (srv *MadlibServer) Next(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	var nextreq = new(NextRequest)
	err := dec.Decode(nextreq)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	log.Printf("Fetching next for key %s from client %s", nextreq.Key, r.RemoteAddr)

	srv.RLock()
	madlib, found := srv.ongoing[nextreq.Key]
	srv.RUnlock()
	if !found {
		log.Println("no key found")
		http.Error(w, fmt.Sprintf("No key '%s' found. Get another one by GETting /new", nextreq.Key), 500)
		return
	}

	resp := &NextResponse{
		Key:   nextreq.Key,
		Done:  !madlib.HasNextPrompt(),
		Title: madlib.Title,
	}
	log.Printf("Key %s from client %s: %v", nextreq.Key, r.RemoteAddr, resp)

	if resp.Done {
		resp.Madlib = madlib.Finish()

		srv.Lock()
		_, found := srv.ongoing[nextreq.Key]
		if found {
			srv.ongoing[nextreq.Key], err = srv.madlibs[rand.Intn(len(srv.madlibs))].NewMadlib()
		}
		srv.Unlock()
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 400)
			return
		}

	} else {
		resp.Prompt = madlib.NextPrompt()
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(resp)
	return
}

func (srv *MadlibServer) Answer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	var ansreq = new(AnswerRequest)
	err := dec.Decode(ansreq)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("Submitting answer %s for key %s from client %s", ansreq.Answer, ansreq.Key, r.RemoteAddr)
	srv.RLock()
	madlib, found := srv.ongoing[ansreq.Key]
	srv.RUnlock()
	if !found {
		log.Println("no key found")
		http.Error(w, fmt.Sprintf("No key '%s' found. Get another one by GETting /new", ansreq.Key), 500)
		return
	}

	madlib.AddAnswer(ansreq.Answer)
	return
}

func (srv *MadlibServer) Skip(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	var ansreq = new(NextRequest)
	err := dec.Decode(ansreq)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("Skipping madlib for key %s from client %s", ansreq.Key, r.RemoteAddr)
	srv.Lock()
	_, found := srv.ongoing[ansreq.Key]
	if found {
		srv.ongoing[ansreq.Key], err = srv.madlibs[rand.Intn(len(srv.madlibs))].NewMadlib()
	}
	srv.Unlock()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}
	if !found {
		log.Println(err)
		http.Error(w, fmt.Sprintf("No key '%s' found. Get another one by GETting /new", ansreq.Key), 500)
		return
	}
	return
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func (srv *MadlibServer) ServeMadlibs() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/api/new", srv.New)
	http.HandleFunc("/api/next", srv.Next)
	http.HandleFunc("/api/answer", srv.Answer)
	http.HandleFunc("/api/skip", srv.Skip)
	http.HandleFunc("/", serveHome)

	log.Println("Serving on", srv.addr)
	log.Fatal(http.ListenAndServe(srv.addr, nil))
}
