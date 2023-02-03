package bind

import (
	"bytes"
	"fmt"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/errors"
	"io"
	"net/http"
	"strings"
)

func ContentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}

func Bind(r *http.Request, result interface{}) error {
	contentType := r.Header.Get("Content-Type")
	codec := encoding.GetCodec(ContentSubtype(contentType))
	if codec == nil {
		return errors.BadRequest("bind", fmt.Sprintf("content-type:%s codec miss", contentType))
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.BadRequest("bind", err.Error())
	}
	if len(body) == 0 {
		return nil
	}

	err = codec.Unmarshal(body, result)
	if err != nil {
		return errors.BadRequest("bind", fmt.Sprintf("coec unmarshal body:%s", err.Error()))
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return nil
}
