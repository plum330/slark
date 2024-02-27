package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/estransport"
	"io"
	"os"
	"strings"
)

type Config struct {
	Addr     []string `json:"addr"`
	UserName string   `json:"user_name"`
	Password string   `json:"password"`
}

type Client struct {
	*elasticsearch.Client
}

func New(c Config) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: c.Addr,
		Username:  c.UserName,
		Password:  c.Password,
		Logger:    &estransport.JSONLogger{Output: os.Stdout, EnableRequestBody: true, EnableResponseBody: true},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}

func (c *Client) CreateIndex(index, str string) ([]byte, error) {
	rsp, err := c.Indices.Create(index, c.Indices.Create.WithBody(strings.NewReader(str)))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

func (c *Client) GetIndex(index []string) ([]byte, error) {
	rsp, err := c.Indices.Get(index)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

func (c *Client) Create(index, ID, docType string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Create(index, ID, buf, c.Client.Create.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// create / update index (if index not exist, create index and doc, create: id default nil)

func (c *Client) Index(index, docType string, doc interface{}, ID ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	var id string
	if len(ID) != 0 {
		id = ID[0]
	}

	opt := []func(*esapi.IndexRequest){
		c.Client.Index.WithDocumentID(id),
		c.Client.Index.WithDocumentType(docType),
	}
	rsp, err := c.Client.Index(index, buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// bulk create / update / index doc

func (c *Client) Bulk(index, docType string, docs []interface{}) ([]byte, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	buf := &bytes.Buffer{}
	for _, doc := range docs {
		//meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%s" } }%s`, "id", "\n")) // 批量操作指定_id(_id存在则更新)
		meta := []byte(fmt.Sprintf(`{ "index" : {} }%s`, "\n"))
		data, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	opt := []func(*esapi.BulkRequest){
		c.Client.Bulk.WithIndex(index),
		c.Client.Bulk.WithDocumentType(docType),
	}
	rsp, err := c.Client.Bulk(buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// update doc

func (c *Client) Update(index, docType, ID string, doc interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(doc)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Client.Update(index, ID, buf, c.Client.Update.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// update by condition

func (c *Client) UpdateByQuery(index []string, docType string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	opt := []func(*esapi.UpdateByQueryRequest){
		c.Client.UpdateByQuery.WithDocumentType(docType),
		c.Client.UpdateByQuery.WithBody(buf),
		c.Client.UpdateByQuery.WithContext(context.Background()),
		c.Client.UpdateByQuery.WithPretty(),
	}
	rsp, err := c.Client.UpdateByQuery(index, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// delete doc

func (c *Client) Delete(index, docType, ID string) ([]byte, error) {
	rsp, err := c.Client.Delete(index, ID, c.Client.Delete.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

func (c *Client) DeleteBulk(index, docType string, ids []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, id := range ids {
		meta := []byte(fmt.Sprintf(`{ "delete" : { "_id" : "%s" } }%s`, id, "\n"))
		buf.Grow(len(meta))
		buf.Write(meta)
	}
	opt := []func(*esapi.BulkRequest){
		c.Client.Bulk.WithIndex(index),
		c.Client.Bulk.WithDocumentType(docType),
	}
	rsp, err := c.Client.Bulk(buf, opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

func (c *Client) DeleteByQuery(index []string, query interface{}, docType ...string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	rsp, err := c.Client.DeleteByQuery(index, buf, c.Client.DeleteByQuery.WithDocumentType(docType...))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// query doc

func (c *Client) Get(index, docType, ID string) ([]byte, error) {
	rsp, err := c.Client.Get(index, ID, c.Client.Get.WithDocumentType(docType))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// search(hit total diff,inaccuracy)

func (c *Client) Search(index, docType string, query interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(query)
	if err != nil {
		return nil, err
	}
	opt := []func(*esapi.SearchRequest){
		c.Client.Search.WithContext(context.Background()),
		c.Client.Search.WithIndex(index),
		c.Client.Search.WithDocumentType(docType),
		c.Client.Search.WithBody(buf),
		c.Client.Search.WithTrackTotalHits(true),
		c.Client.Search.WithPretty(),
		//client.Search.WithFrom(0), // offset
		//client.Search.WithSize(3), // limit
		//client.Search.WithSort([]string{"_source:{name:desc}", "_score:asc", "_id:desc"}...), // 多字段排序
		//client.Search.WithScroll(),
	}
	rsp, err := c.Client.Search(opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}

// count(accurate)

func (c *Client) Count(index, docType []string) ([]byte, error) {
	opt := []func(*esapi.CountRequest){
		c.Client.Count.WithIndex(index...),
		c.Client.Count.WithDocumentType(docType...),
	}
	rsp, err := c.Client.Count(opt...)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	return io.ReadAll(rsp.Body)
}
