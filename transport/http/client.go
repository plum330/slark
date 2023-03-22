package http

import (
	"bytes"
	"context"
	"github.com/go-slark/slark/encoding"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/transport/http/bind"
	"io"
	"net/http"
	"time"
)

type Client struct {
	*http.Client
	transport http.RoundTripper

	// retry
}

type ClientOption func(client *Client)

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		Client:    &http.Client{},
		transport: http.DefaultTransport,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func Timeout(tm time.Duration) ClientOption {
	return func(client *Client) {
		client.Timeout = tm
	}
}

func Transport(tr *http.Transport) ClientOption {
	return func(client *Client) {
		client.transport = tr
	}
}

type Encoder func(ctx context.Context, typ string, v interface{}) ([]byte, error)

type Decoder func(ctx context.Context, rsp *http.Response, v interface{}) error

type ErrDecoder func(ctx context.Context, response *http.Response) error

type Request struct {
	url        string
	method     string
	header     map[string]string
	param      interface{}
	decoder    Decoder
	encoder    Encoder
	errDecoder ErrDecoder
}

func NewRequest() *Request {
	req := &Request{
		encoder: func(ctx context.Context, typ string, v interface{}) ([]byte, error) {
			return encoding.GetCodec(bind.ContentSubtype(typ)).Marshal(v)
		},

		decoder: func(ctx context.Context, rsp *http.Response, v interface{}) error {
			body, err := io.ReadAll(rsp.Body)
			if err != nil {
				return err
			}
			codec := encoding.GetCodec(bind.ContentSubtype(rsp.Header.Get("Content-Type")))
			if codec == nil {
				codec = encoding.GetCodec("json")
			}
			return codec.Unmarshal(body, v)
		},

		errDecoder: func(ctx context.Context, rsp *http.Response) error {
			if rsp.StatusCode >= http.StatusOK && rsp.StatusCode < http.StatusMultipleChoices {
				return nil
			}
			body, err := io.ReadAll(rsp.Body)
			return errors.New(rsp.StatusCode, errors.UnknownReason, errors.UnknownReason).WithError(err).WithReason(string(body))
		},
	}
	return req
}

func (r *Request) URL(url string) *Request {
	r.url = url
	return r
}

func (r *Request) Method(method string) *Request {
	r.method = method
	return r
}

func (r *Request) Param(param interface{}) *Request {
	r.param = param
	return r
}

func (r *Request) Header(header map[string]string) *Request {
	r.header = header
	return r
}

func (r *Request) Decoder(dec Decoder) *Request {
	r.decoder = dec
	return r
}

func (r *Request) Encoder(enc Encoder) *Request {
	r.encoder = enc
	return r
}

func (r *Request) ErrDecoder(errDec ErrDecoder) *Request {
	r.errDecoder = errDec
	return r
}

func (c *Client) DoHTTPReq(ctx context.Context, req *Request, v interface{}) error {
	var body io.Reader
	if req.param != nil {
		enc, err := req.encoder(ctx, req.header["Content-Type"], req.param)
		if err != nil {
			return err
		}
		body = bytes.NewReader(enc)
	}

	request, err := http.NewRequest(req.method, req.url, body)
	if err != nil {
		return err
	}

	for hk, hv := range req.header {
		request.Header.Set(hk, hv)
	}

	rsp, err := c.Do(request)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	err = req.errDecoder(ctx, rsp)
	if err != nil {
		return err
	}

	return req.decoder(ctx, rsp, v)
}
