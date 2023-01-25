package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("request at %s\n", r.RequestURI)
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{"192.0.2.1/32"})
	})
	fmt.Printf("starting...\n")
	http.ListenAndServe(":8000", nil)
}
