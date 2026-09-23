package cache

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"uuid"
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
		fmt.Fprintf(&base, ":%s", formatParam(p))
	}
	return base.String()
}

func formatParam(p any) string {
	if p == nil {
		return ""
	}
	switch v := p.(type) {
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case *int:
		if v == nil {
			return ""
		}
		return strconv.Itoa(*v)
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.Format(time.RFC3339)
	case *uuid.UUID:
		if v == nil {
			return ""
		}
		return v.String()
	}
	return fmt.Sprintf("%v", p)
}
