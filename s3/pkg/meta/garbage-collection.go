package meta

import (
	. "github.com/opensds/multi-cloud/s3/pkg/meta/types"
)

// Insert object to `garbageCollection` table
func (m *Meta) PutObjectToGarbageCollection(object *Object) error {
	return m.Client.PutObjectToGarbageCollection(object, nil)
}

func (m *Meta) ScanGarbageCollection(limit int, startRowKey string) ([]GarbageCollection, error) {
	return m.Client.ScanGarbageCollection(limit, startRowKey)
}

func (m *Meta) RemoveGarbageCollection(garbage GarbageCollection) error {
	return m.Client.RemoveGarbageCollection(garbage)
}
