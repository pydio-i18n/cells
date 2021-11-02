package service

import (
	"github.com/pydio/cells/v4/common/registry"
	"strings"
)

// Service for the pydio app
type service struct {
	opts ServiceOptions
}

var (
	mandatoryOptions []ServiceOption
)

type Service interface{
}

func NewService(opts ...ServiceOption) Service {
	s := &service{
		opts: newOptions(append(mandatoryOptions, opts...)...),
	}

	name := s.opts.Name
	tags := s.opts.Tags

	registry.Register(name, strings.Join(tags, " "))

	return s
}
