package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/apex/gateway/v2"
)

func dumpobj(prefix string, x interface{}) string {

	s := ""
	switch t := x.(type) {
	case map[string]interface{}:
		for k, v := range t {
			s += dumpobj(prefix+"."+k, v)
		}
	case []interface{}:
		for i, v := range t {
			s += dumpobj(prefix+"["+strconv.Itoa(i)+"]", v)
		}
	case string:
		s += fmt.Sprintf("%s = %q\n", prefix, t)
	case float64:
		s += fmt.Sprintf("%s = %f\n", prefix, t)
	default:
		fmt.Printf("Unhandled: %T\n", t)
	}
	return s
}

func (s *server) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		t, err := template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
<title>JSON unroller</title>
<meta name="viewport" content="width=device-width, initial-scale=1.0" />
<link rel="icon" href="data:,">
<style>
textarea {
  width: 100%;
  height: 50vh;
}
</style>
</head>
<body>
<form action="/unroll" method="post">
<p>
<textarea required name="content"></textarea>
<p>
<input type="submit">
</form>
<p>
<pre>curl --data-urlencode content@foobar.json https://jsonunroller.dabase.com/unroll</pre>
<p>
<a href=https://github.com/kaihendry/jsonunroller>MIT source code</a>
</body>
</html>`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func (s *server) unroll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// if not post then return
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		f := r.FormValue("content")
		if f == "" {
			err := errors.New("No content!")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var pj interface{}
		err := json.Unmarshal([]byte(f), &pj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		d := dumpobj("this", pj)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, d)
	}
}

type server struct {
	router *http.ServeMux
}

func newServer(local bool) *server {
	s := &server{router: &http.ServeMux{}}

	s.router.Handle("/", s.index())
	s.router.Handle("/unroll", s.unroll())

	return s
}

func main() {
	_, awsDetected := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME")
	s := newServer(!awsDetected)
	if awsDetected {
		gateway.ListenAndServe("", s.router)
	} else {
		http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), s.router)
	}
}
