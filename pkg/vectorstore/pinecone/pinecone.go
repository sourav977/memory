package pinecone

import (
	pc "github.com/nekomeowww/go-pinecone"
)

type Client struct {
	c         *pc.Client
	indexName string
	namespace string
}

func NewClient(opts ...CallOptions) *Client {
	return nil
}

func (p *Client) StoreVector(id string, vector []float32) {}

func (p *Client) QueryVector(vector []float32, k int) ([]float32, error) {
	return nil, nil
}
