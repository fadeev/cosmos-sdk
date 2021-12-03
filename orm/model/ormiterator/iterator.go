package ormiterator

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type Iterator interface {
	Next() (bool, error)
	IndexKey() ([]protoreflect.Value, error)
	PrimaryKey() ([]protoreflect.Value, error)
	GetMessage(proto.Message) error

	Cursor() Cursor
	Close()

	mustEmbedUnimplementedIterator()
}

type Cursor []byte

type UnimplementedIterator struct{}

func (u UnimplementedIterator) mustEmbedUnimplementedIterator() {}

func (u UnimplementedIterator) Next() (bool, error) {
	return false, ormerrors.UnsupportedOperation
}

func (u UnimplementedIterator) IndexKey() ([]protoreflect.Value, error) {
	return nil, ormerrors.UnsupportedOperation
}

func (u UnimplementedIterator) PrimaryKey() ([]protoreflect.Value, error) {
	return nil, ormerrors.UnsupportedOperation
}

func (u UnimplementedIterator) GetMessage(proto.Message) error {
	return ormerrors.UnsupportedOperation
}

func (u UnimplementedIterator) Cursor() Cursor { return nil }

func (u UnimplementedIterator) Close() {}

var _ Iterator = UnimplementedIterator{}

type ErrIterator struct {
	UnimplementedIterator
	Err error
}

func (e ErrIterator) Next() (bool, error) { return false, e.Err }

func (e ErrIterator) IndexKey() ([]protoreflect.Value, error) { return nil, e.Err }

func (e ErrIterator) PrimaryKey() ([]protoreflect.Value, error) { return nil, e.Err }

func (e ErrIterator) GetMessage(proto.Message) error { return e.Err }

func (e ErrIterator) Cursor() Cursor { return nil }

func (e ErrIterator) Close() {}

var _ Iterator = ErrIterator{}