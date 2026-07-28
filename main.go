package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("AGENT_LOCAL_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "5000"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello world from GGolang agents!")
	})

	fmt.Printf("Starting hello-world agent on 0.0.0.0:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
