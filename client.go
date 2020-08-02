package segvdl

import (
	"context"
	"github.com/pkg/errors"
	"net/http"
)

type Response struct {
	*http.Response
	client *Client
}

func (rsp *Response) Read(p []byte) (n int, err error) {
	return rsp.Response.Body.Read(p)
}

func (rsp *Response) Close() error {
	err := rsp.Response.Body.Close()
	return err
}

type Client struct {}

func NewClient() *Client {
	return &Client{}
}

func (client *Client) Get(ctx context.Context, url string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "request error")
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "response error")
	}
	return &Response{rsp, client}, nil
}


