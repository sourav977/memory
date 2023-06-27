package pc

import (
	"context"
	"fmt"
	"github.com/aldarisbm/memory/internal"
	"github.com/aldarisbm/memory/types"
	"github.com/aldarisbm/memory/vectorstore"
	"github.com/google/uuid"
	"github.com/nekomeowww/go-pinecone"
)

type storer struct {
	client    *pinecone.IndexClient
	namespace string
	DTO       *DTO
}

// NewStorer returns a new storer
func NewStorer(opts ...CallOptions) *storer {
	o := applyCallOptions(opts, options{
		namespace: internal.Generate(10),
	})
	c, err := pinecone.NewIndexClient(
		pinecone.WithAPIKey(o.apiKey),
		pinecone.WithIndexName(o.indexName),
		pinecone.WithEnvironment(o.environment),
		pinecone.WithProjectName(o.projectName),
	)
	if err != nil {
		panic(err)
	}
	return &storer{
		client:    c,
		namespace: o.namespace,
		DTO: &DTO{
			ApiKey:      o.apiKey,
			IndexName:   o.indexName,
			Namespace:   o.namespace,
			ProjectName: o.projectName,
			Environment: o.environment,
		},
	}
}

// StoreVector stores the given Document
// it attempts to save the metadata
func (p *storer) StoreVector(doc *types.Document) error {
	ctx := context.Background()
	req := pinecone.UpsertVectorsParams{
		Vectors: []*pinecone.Vector{
			{
				ID:       doc.ID.String(),
				Values:   doc.Vector,
				Metadata: doc.Metadata,
			},
		},
		Namespace: p.namespace,
	}

	resp, err := p.client.UpsertVectors(ctx, req)
	if err != nil {
		return fmt.Errorf("storing vector: %w", err)
	}
	if resp.UpsertedCount != 1 {
		return fmt.Errorf("storing vector: upserted count is not 1")
	}
	return nil
}

// QuerySimilarity returns the k most similar documents to the given vector
func (p *storer) QuerySimilarity(vector []float32, k int64) ([]uuid.UUID, error) {
	ctx := context.Background()
	req := pinecone.QueryParams{
		Vector:    vector,
		Namespace: p.namespace,
		TopK:      k,
	}
	resp, err := p.client.Query(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("querying vector: %w", err)
	}
	if len(resp.Matches) == 0 {
		// should we return an error here?
		return nil, nil
	}

	var uuids []uuid.UUID
	for _, match := range resp.Matches {
		id, err := uuid.Parse(match.ID)
		if err != nil {
			return nil, fmt.Errorf("querying vector: %w", err)
		}
		uuids = append(uuids, id)
	}
	return uuids, nil
}

// Close closes the storer not necessary for pinecone
func (p *storer) Close() error {
	return nil
}

func (p *storer) GetNamespace() string {
	return p.namespace
}

func (p *storer) GetDTO() vectorstore.Converter {
	return p.DTO
}

// Ensure that storer implements VectorStorer
var _ vectorstore.VectorStorer = (*storer)(nil)
