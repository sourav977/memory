package boltdb

import (
	"encoding/json"
	"fmt"
	"github.com/aldarisbm/memory/datasource"
	"github.com/aldarisbm/memory/internal"
	"github.com/aldarisbm/memory/types"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"os"
)

// bucketName is always the same
const (
	bucketName = "ltm"
	mode       = os.ModePerm
)

type localStorer struct {
	path string
	db   *bolt.DB
	DTO  *DTO
}

// NewLocalStorer returns a new local storer
// if path is empty, it will default to $HOME/memory/boltdb
func NewLocalStorer(opts ...CallOptions) *localStorer {
	o := applyCallOptions(opts)
	if o.path == "" {
		o.path = fmt.Sprintf("%s/%s", internal.CreateMemoryFolderInHomeDir(), internal.Generate(10))
	}
	dbm, err := bolt.Open(o.path, mode, nil)
	if err != nil {
		panic(err)
	}
	err = dbm.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	ls := &localStorer{
		db:   dbm,
		path: o.path,
		DTO: &DTO{
			Path: o.path,
		},
	}
	return ls
}

// Close closes the local storer
func (l *localStorer) Close() error {
	return l.db.Close()
}

// GetDocument returns the document with the given id
func (l *localStorer) GetDocument(id uuid.UUID) (*types.Document, error) {
	var doc types.Document
	err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(id.String()))
		err := json.Unmarshal(v, &doc)
		if err != nil {
			return fmt.Errorf("unmarshaling document: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetDocuments returns the documents with the given ids
func (l *localStorer) GetDocuments(ids []uuid.UUID) ([]*types.Document, error) {
	var docs []*types.Document
	for _, id := range ids {
		doc, err := l.GetDocument(id)
		if err != nil {
			return nil, fmt.Errorf("getting document: %s", err)
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// StoreDocument stores the given document
// We use a k/v store key being uuid and value being []byte of Document
func (l *localStorer) StoreDocument(document *types.Document) error {
	doc, err := json.Marshal(&document)
	if err != nil {
		return fmt.Errorf("marshaling document: %s", err)
	}
	err = l.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		err := b.Put([]byte(document.ID.String()), doc)
		return err
	})
	if err != nil {
		return fmt.Errorf("updating bolt db: %s", err)
	}
	return nil
}

func (l *localStorer) GetDTO() datasource.Converter {
	return l.DTO
}

// Ensure localStorer implements DataSourcer
var _ datasource.DataSourcer = (*localStorer)(nil)
