package cache

import (
	"fmt"
	"strings"
)

type keyBuilder struct {
	prefix string
}

type KeyBuilder interface {
	BuildItemKey(id string) string
	BuildListKey(params ...any) string
	GetPrefix() string
}

func (kb *keyBuilder) GetPrefix() string {
	return kb.prefix
}

func NewKeyBuilder(prefix string) *keyBuilder {
	return &keyBuilder{prefix: prefix}
}

func (kb *keyBuilder) BuildItemKey(id string) string {
	return fmt.Sprintf("%s:%s", kb.prefix, id)
}

func (kb *keyBuilder) BuildListKey(params ...any) string {
	var base strings.Builder
	fmt.Fprintf(&base, "%s:list", kb.prefix)
	for _, p := range params {
		fmt.Fprintf(&base, ":%v", p)
	}
	return base.String()
}
