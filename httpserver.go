package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &StatusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		h.ServeHTTP(recorder, r)
		log.Printf("Handling request for %s from %s, status: %d", r.URL.Path, r.RemoteAddr, recorder.Status)
	})
}

func main() {
	// define log
	fn := "mylog.log"
	logFile, err := os.Create(fn)
	defer logFile.Close()
	if err != nil {
		log.Fatalln("Open file error!")
	}
	log.SetOutput(logFile)
	myHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s", r.RemoteAddr)
	})
	handlerWithLogging := WithLogging(myHandler)
	http.Handle("/test/", handlerWithLogging)
	//define handles
	http.HandleFunc("/", HelloHandler)
	http.HandleFunc("/rheader/", handler)
	http.HandleFunc("/sysvar/", systemvar)
	http.HandleFunc("/healthz/", healthz)
	http.HandleFunc("/userip/", userip)
	fmt.Println("Server started at port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "%s %s %s \n", r.Method, r.URL, r.Proto)
	//Iterate over all header fields
	for k, v := range r.Header {
		w.Header().Set(k, v[0])

	}
	//w.Header().Set("mykey", "myvalue")
	for m, n := range w.Header() {
		fmt.Fprintf(w, "Header field %q, Value %q\n", m, n)
	}
}

func systemvar(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "10"
	}
	w.Header().Set("VERSION", version)
	//fmt.Fprintf(w, "Header field 'VERSION': %q ", version)
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, there\n")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "200\n")
}

func userip(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, r.Header.Get("X-Forwarded-For"))
}
