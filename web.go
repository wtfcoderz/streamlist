package main

import (
	//"crypto/sha512"
	//"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	//	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	humanize "github.com/dustin/go-humanize"
	"github.com/julienschmidt/httprouter"
	//"github.com/gorilla/securecookie"
)

/*
	"hasmedia": func(lid, mid string) bool {
		list, err := FindList(lid)
		if err != nil {
			return false
		}
		media, err := FindMedia(mid)
		if err != nil {
			return false
		}
		return list.HasMedia(media)
	},
*/

var (
	funcMap = template.FuncMap{
		"sub": func(a, b int64) int64 {
			return a - b
		},
		"add": func(a, b int64) int64 {
			return a + b
		},
		"nums": func(max int) (nums []int) {
			for i := 0; i < max; i++ {
				nums = append(nums, i)
			}
			return nums
		},
		"mediaexists": func(id string) bool {
			_, err := FindMedia(id)
			return err == nil
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"bytes": func(n int64) string {
			return humanize.Bytes(uint64(n))
		},
		"time": humanize.Time,
		"duration": func(seconds int64) string {
			hours := seconds / 3600
			seconds -= hours * 3600

			minutes := seconds / 60
			seconds -= minutes * 60

			if hours > 0 {
				return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
			}
			return fmt.Sprintf("%d:%02d", minutes, seconds)
		},
	}
	errorPageHTML = `
        <html>
            <head>
                <title>Error</title>
            </head>
            <body>
                <h2 style="color: orangered;">An error has occurred. <a href="/streamlist/logs">Check the logs</a></h2>
            </body>
        </html>
    `
)

func redirect(w http.ResponseWriter, r *http.Request, format string, a ...interface{}) {
	location := httpPrefix
	location += fmt.Sprintf(format, a...)
	http.Redirect(w, r, location, http.StatusFound)
}

func _error(w http.ResponseWriter, err error) {
	logger.Error(err)

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, errorPageHTML)
}

func prefix(path string) string {
	return httpPrefix + path
}

func log(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Request info
		addr := r.RemoteAddr
		xff := r.Header.Get("X-Forwarded-For")
		realip := r.Header.Get("X-Real-IP")
		method := r.Method
		rang := r.Header.Get("Range")
		path := r.RequestURI

		// Run the handler
		start := time.Now()
		h(w, r, ps)
		elapsed := int64(time.Since(start) / time.Millisecond)

		// Response info
		mime := w.Header().Get("Content-Type")
		logger.Infof("%q %q %q %q %q %q %q %d ms", addr, xff, realip, method, path, rang, mime, elapsed)
	}
}

func auth(h httprouter.Handle, role string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		juser := ""

		if role == "none" {
			h(w, r, ps)
			return
		}

		// If token, refresh it and send response
		reqToken, tokErr := r.Cookie("X-Streamlist-Token")
		if tokErr != http.ErrNoCookie {
			fmt.Println("a tokan?" + reqToken.Value)
			token, err := jwt.Parse(reqToken.Value, func(t *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err == nil && token.Valid {
				//		fmt.Printf("valid")
				juser = token.Claims.(jwt.MapClaims)["user"].(string)
				fmt.Println(juser)
				ps = append(ps, httprouter.Param{Key: "user", Value: juser})
				ps = append(ps, httprouter.Param{Key: "role", Value: "admin"})
				//		// Create JWT token
				//		token = jwt.New(jwt.GetSigningMethod("HS256"))
				//		claims := make(jwt.MapClaims)
				//		claims["user"] = juser
				//		claims["exp"] = time.Now().Add(time.Minute * 3600).Unix()
				//		token.Claims = claims
				//		tokenString, err := token.SignedString([]byte(secretKey))
				//		if err != nil {
				//			panic(err)
				//		}
				w.Header().Set("X-Streamlist-Token", "*")
				fmt.Println("auth - token ok")
			} else {
				fmt.Printf("token invalid")
				redirect(w, r, "/logout")
			}
		} else {
			redirect(w, r, "/logout")
		}
		h(w, r, ps)
	}
}

func toXML(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	if err := enc.Encode(data); err != nil {
		logger.Error(err)
	}
}

func toJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		logger.Error(err)
	}
}

func html(w http.ResponseWriter, target string, data interface{}) {
	t := template.New(target)
	t.Funcs(funcMap)
	for _, filename := range AssetNames() {
		if !strings.HasPrefix(filename, "templates/") {
			continue
		}
		name := strings.TrimPrefix(filename, "templates/")
		b, err := Asset(filename)
		if err != nil {
			_error(w, err)
			return
		}

		var tmpl *template.Template
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		if _, err := tmpl.Parse(string(b)); err != nil {
			_error(w, err)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		_error(w, err)
		return
	}
}
