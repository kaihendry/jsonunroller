package jsonunroller

import (
	//	"appengine"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	default:
		fmt.Printf("Unhandled: %T\n", t)
	}
	return s
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/unroll", unroll)
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, postForm)
}

const postForm = `
<html>
<body>
<form action="/unroll" method="post">
<div><textarea style="box-sizing: border-box; height: 60%; width: 100%;" name="content"></textarea></div>
<div><input type="submit"></div>
<pre>curl -d 'content={ "foo": "bar" }' http://jsonunroller.appspot.com/unroll<pre>
<p><a href=https://github.com/kaihendry/GAE-jsonunroller>MIT source code</a></p>
</form>
</body>
</html>
`

func unroll(w http.ResponseWriter, r *http.Request) {

	f := r.FormValue("content")
	if f == "" {
		err := errors.New("No content!")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var pj interface{}
	err := json.Unmarshal([]byte(f), &pj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := dumpobj("this", pj)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, s)

}
