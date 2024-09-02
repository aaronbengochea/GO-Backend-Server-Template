package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got request: / \n")
	io.WriteString(w, "This is the root of the server")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got request: /hello \n")
	io.WriteString(w, "go server, hello!")
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", getRoot)
	r.HandleFunc("/hello", getHello)
	http.Handle("/", r)
	//http.HandleFunc("/", getRoot)
	//http.HandleFunc("/hello", getHello)

	go func() {
		err := http.ListenAndServe(":3333", nil)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed \n")
		} else if err != nil {
			fmt.Printf("error starting server: %s \n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("Server is listening on port 3333 \n")
	select {}
}
