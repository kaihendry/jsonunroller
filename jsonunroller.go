package jsonunroller

import (
	"appengine"
	"appengine/urlfetch"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	u := r.FormValue("u")
	log.Println("URL:", u)
	if u == "" {
		fmt.Fprint(w, postForm)
		return
	}

	pu, err := url.ParseRequestURI(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if pu.IsAbs() != true {
		err = errors.New("Not absolute URL")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := appengine.NewContext(r)
	client := urlfetch.Client(c)
	res, err := client.Get(pu.String())
	defer res.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contenttype := res.Header.Get("Content-Type")
	if strings.Contains(contenttype, "application/json") != true {
		err = errors.New("Not application/json: " + contenttype)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	j, err := ioutil.ReadAll(res.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var pj interface{}
	err = json.Unmarshal([]byte(j), &pj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s := dumpobj("this", pj)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, s)

}

const postForm = `
<html>
<body>
<form action="/" method="get">
<div><input type=url name=u size=100></input></div>
<div><input type="submit" value="Unroll JSON URL"></div>
</form>
<form action="/unroll" method="post">
<div><textarea style="box-sizing: border-box; height: 60%; width: 100%;" name="content"></textarea></div>
<div><input type="submit"></div>
<pre>curl --data-urlencode content@foobar.json http://jsonunroller.appspot.com/unroll<pre>
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
