// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package kvstore

import (
	"github.com/mattermost/mattermost-plugin-jira/server/utils/types"
)

type ValueIndexStore interface {
	Load() (*types.ValueSet, error)
	Store(*types.ValueSet) error
	Delete(id types.ID) error
	StoreValue(v types.Value) error
}

type valueIndexStore struct {
	proto types.ValueArray
	kv    KVStore
	key   string
}

func (s *store) ValueIndex(key string, proto types.ValueArray) ValueIndexStore {
	return &valueIndexStore{
		key:   key,
		kv:    s.KVStore,
		proto: proto,
	}
}

func (s *valueIndexStore) Load() (*types.ValueSet, error) {
	index := types.NewValueSet(s.proto)
	err := LoadJSON(s.kv, s.key, &index)
	if err != nil {
		return nil, err
	}
	return index, nil
}

func (s *valueIndexStore) Store(index *types.ValueSet) error {
	err := StoreJSON(s.kv, s.key, index)
	if err != nil {
		return err
	}
	return nil
}

func (s *valueIndexStore) Delete(id types.ID) error {
	index, err := s.Load()
	if err != nil {
		return err
	}

	index.Delete(id)

	return s.Store(index)
}

func (s *valueIndexStore) StoreValue(v types.Value) error {
	index, err := s.Load()
	if err != nil {
		return err
	}

	index.Set(v)

	return s.Store(index)
}
