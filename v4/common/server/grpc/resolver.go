package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/pydio/cells/v4/common/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"regexp"
)

const (
	defaultPort ="8001"
)

var (
	errMissingAddr = errors.New("cells resolver: missing address")

	errAddrMisMatch = errors.New("cells resolver: invalid uri")

	regex, _ = regexp.Compile("^([A-z0-9.]*?)(:[0-9]{1,5})?\\/([A-z_]*)$")
)

func init(){
	resolver.Register(NewBuilder())
}

type cellsBuilder struct {
}

type cellsResolver struct {
	reg registry.Registry
	address string
	cc resolver.ClientConn
	name string
	m map[string][]string
	disableServiceConfig bool
}

func NewBuilder() resolver.Builder {
	return &cellsBuilder{}
}

func (b *cellsBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint))
	if err != nil {
		return nil, err
	}

	reg, err  := registry.OpenRegistry(context.Background(), fmt.Sprintf("grpc://%s%s", host, port))
	if err != nil {
		return nil, err
	}

	services, err := reg.ListServices()
	if err != nil {
		return nil, err
	}

	var m = map[string][]string{}
	for _, s := range services {
		for _, n := range s.Nodes() {
			m[n.Address()[0]] = append(m[n.Address()[0]], s.Name())
		}
	}

	cr := &cellsResolver{
		reg: reg,
		name: name,
		cc: cc,
		m: m,
		disableServiceConfig: opts.DisableServiceConfig,
	}

	cr.updateState()
	go cr.watch()

	return cr, nil
}

func (cr *cellsResolver) watch() {
	w, err := cr.reg.WatchServices()
	if err != nil {
		return
	}

	for {
		r, err := w.Next()
		if err != nil {
			return
		}

		// s := r.Service()
		if r.Action() == "create" {
			/*for _, n := range r.Service().Nodes() {
				cr.m[n.Address()[0]] = append(cr.m[n.Address()[0]], s.Name())
			}*/

			cr.updateState()
		}
	}
}

func (cr *cellsResolver) updateState() error {
	var addresses []resolver.Address
	for k, v := range cr.m {
		addresses = append(addresses, resolver.Address{
			Addr: k,
			ServerName: "main",
			Attributes: attributes.New("services", v),
		})
	}

	if err := cr.cc.UpdateState(resolver.State{
		Addresses: addresses,
		ServiceConfig: cr.cc.ParseServiceConfig(`{"loadBalancingPolicy": "lb"}`),
	}); err != nil {
		return err
	}

	return nil
}

func (b *cellsBuilder) Scheme() string {
	return "cells"
}

func (cr *cellsResolver) ResolveNow(opt resolver.ResolveNowOptions) {
}

func (cr *cellsResolver) Close() {
}

func parseTarget(target string) (host, port, name string, err error) {
	if target == "" {
		return "", "", "", errMissingAddr
	}

	if !regex.MatchString(target) {
		return "", "", "", errAddrMisMatch
	}

	groups := regex.FindStringSubmatch(target)
	host = groups[1]
	port = groups[2]
	name = groups[3]
	if port == "" {
		port = defaultPort
	}
	return host, port, name, nil
}