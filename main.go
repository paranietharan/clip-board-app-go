package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type ValueStore struct {
	mu    sync.RWMutex
	value string
}

func (vs *ValueStore) Get() string {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return vs.value
}

func (vs *ValueStore) Set(v string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.value = v
}

func main() {
	store := &ValueStore{value: ""}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			setValue := r.URL.Query().Get("set")
			if setValue != "" {
				store.Set(setValue)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]string{"value": store.Get()}); err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				log.Printf("error encoding JSON: %v", err)
			}

		case http.MethodPost:
			var body struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			store.Set(body.Value)
			w.WriteHeader(http.StatusNoContent)

		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Listening on :8080 …")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
