package http

import (
	"fmt"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/encoding/form"
	"github.com/go-slark/slark/encoding/json"
	"github.com/go-slark/slark/errors"
	utils "github.com/go-slark/slark/pkg"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func SubContentType(name string) string {
	left := strings.Index(name, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(name, ";")
	if right == -1 {
		right = len(name)
	}
	if right < left {
		return ""
	}
	return name[left+1 : right]
}

func Codec(req *http.Request, name string) (encoding.Codec, bool) {
	for _, n := range req.Header[name] {
		codec := encoding.GetCodec(SubContentType(n))
		if codec != nil {
			return codec, true
		}
	}

	return encoding.GetCodec(json.Name), false
}

type Codecs struct {
	bodyDecoder  func(*http.Request, interface{}) error
	varsDecoder  func(*http.Request, interface{}) error
	queryDecoder func(*http.Request, interface{}) error
	rspEncoder   func(*http.Request, http.ResponseWriter, interface{}) error
	errorEncoder func(*http.Request, http.ResponseWriter, error)
}

func RequestBodyDecoder(req *http.Request, v interface{}) error {
	codec, valid := Codec(req, utils.ContentType)
	if !valid {
		return errors.BadRequest("request body decoder", fmt.Sprintf("content-type:%s codec miss", req.Header.Get(utils.ContentType)))
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return errors.BadRequest("request body decoder", err.Error())
	}
	if len(body) == 0 {
		return nil
	}

	err = codec.Unmarshal(body, v)
	if err != nil {
		return errors.BadRequest("request body decoder", fmt.Sprintf("coec unmarshal body:%s", err.Error()))
	}

	//req.Body = io.NopCloser(bytes.NewBuffer(body))
	return nil
}

func bind(vars url.Values, v interface{}) error {
	if err := encoding.GetCodec(form.Name).Unmarshal([]byte(vars.Encode()), v); err != nil {
		return errors.BadRequest("bind query", err.Error())
	}
	return nil
}

func RequestVarsDecoder(req *http.Request, v interface{}) error {
	params, ok := req.Context().Value(utils.RequestVars).(map[string]string)
	if !ok {
		return nil
	}
	vars := make(url.Values, len(params))
	for key, value := range params {
		vars[key] = []string{value}
	}
	return bind(vars, v)
}

func RequestQueryDecoder(req *http.Request, v interface{}) error {
	return bind(req.URL.Query(), v)
}

func SetContentType(subtype string) string {
	return strings.Join([]string{utils.Application, subtype}, "/")
}

func ResponseEncoder(req *http.Request, rsp http.ResponseWriter, v interface{}) error {
	codec, _ := Codec(req, utils.Accept)
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	rsp.Header().Set(utils.ContentType, SetContentType(codec.Name()))
	_, err = rsp.Write(data)
	return err
}

func ErrorEncoder(req *http.Request, rsp http.ResponseWriter, err error) {
	e := errors.FromError(err)
	codec, _ := Codec(req, utils.Accept)
	body, err := codec.Marshal(e)
	if err != nil {
		rsp.WriteHeader(http.StatusInternalServerError)
		return
	}
	rsp.Header().Set(utils.ContentType, SetContentType(codec.Name()))
	rsp.WriteHeader(int(e.Code))
	_, _ = rsp.Write(body)
}
