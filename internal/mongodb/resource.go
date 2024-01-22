// SPDX-FileCopyrightText: 2023 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: MPL-2.0

package mongodb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type ResourceWrapper struct {
	Union Resource
}

// Ensure it implements the interfaces.
var _ bson.ValueMarshaler = &ResourceWrapper{}
var _ bson.ValueUnmarshaler = &ResourceWrapper{}

// MarshalBSONValue implements [bson.ValueMarshaler].
func (r *ResourceWrapper) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(r.Union)
}

// UnmarshalBSONValue implements [bson.ValueUnmarshaler].
func (r *ResourceWrapper) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	rv := bson.RawValue{Type: t, Value: data}
	var m map[string]any
	if err := rv.Unmarshal(&m); err != nil {
		return err
	}
	if value, ok := m["system_buckets"]; ok {
		if buckets, ok := value.(string); ok {
			r.Union = ResourceSystemBuckets{SystemBuckets: buckets}
			return nil
		} else {
			return fmt.Errorf("resource.system_buckets must be string, got %T", value)
		}
	}
	if value, ok := m["anyResource"]; ok {
		if anyResource, ok := value.(bool); ok {
			r.Union = ResourceAny{AnyResource: anyResource}
			return nil
		} else {
			return fmt.Errorf("resource.anyResource must be bool, got %T", value)
		}
	}
	if value, ok := m["cluster"]; ok {
		if cluster, ok := value.(bool); ok {
			r.Union = ResourceCluster{Cluster: cluster}
			return nil
		} else {
			return fmt.Errorf("resource.cluster must be bool, got %T", value)
		}
	}
	var col ResourceCollection
	if err := rv.Unmarshal(&col); err != nil {
		return err
	}
	r.Union = col
	return nil
}

type Resource interface {
	isResource()
}

type ResourceAny struct {
	AnyResource bool `bson:"anyResource"`
}

func (ResourceAny) isResource() {}

type ResourceCluster struct {
	Cluster bool `bson:"cluster"`
}

func (ResourceCluster) isResource() {}

type ResourceCollection struct {
	DB         string `bson:"db"`
	Collection string `bson:"collection"`
}

func (ResourceCollection) isResource() {}

type ResourceSystemBuckets struct {
	SystemBuckets string `bson:"system_buckets"`
}

func (ResourceSystemBuckets) isResource() {}
