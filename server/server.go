package server

import (
	"fmt"
	"golaco/model"
	"golaco/service"
	"io"
	"net/http"
	"strings"
	"time"
)

type handler struct{}

func request(r *http.Request) model.Data {
	return model.Data{
		Address: "[" + strings.Split(r.RemoteAddr, ":")[0] + "]",
		URL:     r.URL,
		Path:    strings.ReplaceAll(r.URL.Path, "/", ""),
		Query:   r.URL.Query(),
		Method:  r.Method,
		Header:  r.Header,
		Body:    r.Body,
	}
}

func response(w http.ResponseWriter, data model.Data, callback model.Callback) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(callback.Code)
	if callback.Code > 399 && callback.Code < 500 {
		if callback.Err != nil {
			io.WriteString(w, fmt.Sprintf(`{"error":%q}`, callback.Err))
		}
	} else if callback.Code < 500 {
		io.WriteString(w, string(callback.Result))
	}
	color := map[bool]string{
		callback.Code > 199 && callback.Code < 300: "\033[32m",
		callback.Code > 299 && callback.Code < 400: "\033[35m",
		callback.Code > 399 && callback.Code < 500: "\033[33m",
		callback.Code > 499 && callback.Code < 600: "\033[31m",
	}
	if callback.Code < 500 {
		fmt.Println(color[true], fmt.Sprintf("%s %s %d %s %s", time.Now().Format("2006-01-02 15:04:05"), data.Address, callback.Code, data.Method, data.URL))
	} else {
		fmt.Println(color[true], fmt.Sprintf("%s %s %d %s %s\terror: %s", time.Now().Format("2006-01-02 15:04:05"), data.Address, callback.Code, data.Method, data.URL, callback.Err))
	}
	fmt.Print("\033[0m")
}

func router(data model.Data, w http.ResponseWriter) model.Callback {
	routes := map[string]func(model.Data) model.Callback{
		"car":   service.Cars,
		"user":  service.Users,
		"login": service.Login,
	}
	for k := range routes {
		if k == data.Path {
			return routes[k](data)
		}
	}
	return model.Callback{
		Code:   404,
		Result: nil,
		Err:    nil,
	}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := request(r)
	callback := router(data, w)
	response(w, data, callback)
}

func Start() {
	fmt.Println("Server started!")
	http.ListenAndServe(":9000", handler{})
}
