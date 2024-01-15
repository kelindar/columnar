// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package main

import (
	"github.com/kelindar/column"
)

// Cache represents a key-value store
type Cache struct {
	store *column.Collection
}

// New creates a new key-value cache
func New() *Cache {
	db := column.NewCollection()
	db.CreateColumn("key", column.ForKey())
	db.CreateColumn("val", column.ForString())

	return &Cache{
		store: db,
	}
}

// Get attempts to retrieve a value for a key
func (c *Cache) Get(key string) (value string, found bool) {
	c.store.QueryKey(key, func(r column.Row) error {
		value, found = r.String("val")
		return nil
	})
	return
}

// Set updates or inserts a new value
func (c *Cache) Set(key, value string) {
	if err := c.store.UpsertKey(key, func(r column.Row) error {
		r.SetString("val", value)
		return nil
	}); err != nil {
		panic(err)
	}
}
