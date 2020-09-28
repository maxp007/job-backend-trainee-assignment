package cache

import (
	"context"
)

// DummyCacheCommon is a struct implementing ICacher interface, made for tests, returns false, nil error
type DummyCacheCommon struct {
}

func (c *DummyCacheCommon) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	return false, nil
}

//AddKey does nothing,always returns nil error
func (c *DummyCacheCommon) AddKey(ctx context.Context, key string) (err error) {
	return nil
}

//StubCacheFaulty is a struct implementing ICacher interface,made for tests, its method always returns an error
type StubCacheFaulty struct {
}

//CheckKeyExistence checks nothing,always returns error
func (c *StubCacheFaulty) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	return false, ErrFailedToGetConnFromPool
}

//AddKey does nothing,always returns error
func (c *StubCacheFaulty) AddKey(ctx context.Context, key string) (err error) {
	return ErrFailedToGetConnFromPool
}

// DummyCacheWithAnyKeyExists is a struct implementing ICacher interface, made for tests
type DummyCacheWithAnyKeyExists struct {
}

func (c *DummyCacheWithAnyKeyExists) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	return true, nil
}

//AddKey does nothing,always returns nil error
func (c *DummyCacheWithAnyKeyExists) AddKey(ctx context.Context, key string) (err error) {
	return nil
}

// DummyCacheWithNoKeyExists is a struct implementing ICacher interface, made for tests
type DummyCacheWithNoKeyExists struct {
}

func (c *DummyCacheWithNoKeyExists) CheckKeyExistence(ctx context.Context, key string) (keyExists bool, err error) {
	return false, nil
}

//AddKey does nothing,always returns nil error
func (c *DummyCacheWithNoKeyExists) AddKey(ctx context.Context, key string) (err error) {
	return nil
}
