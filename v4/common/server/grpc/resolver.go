package grpc

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/resolver"
	"regexp"
)

const (
	defaultPort ="8001"
)

var (
	errMissingAddr = errors.New("cells resolver: missing address")

	errAddrMisMatch = errors.New("cells resolver: invalid uri")

	regex, _ = regexp.Compile("^([A-z0-9.]+)(:[0-9]{1,5})?/([A-z_]+)$")
)

func init(){
	resolver.Register(NewBuilder())
}

type cellsBuilder struct {
}

type cellsResolver struct {
	address string
	cc resolver.ClientConn
	name string
	disableServiceConfig bool
}

func NewBuilder() resolver.Builder {
	return &cellsBuilder{}
}

func (b *cellsBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	fmt.Println("calling build")

	host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint))
	if err != nil {
		return nil, err
	}


	cr := &cellsResolver{
		address: fmt.Sprintf("%s%s", host, port),
		name: name,
		cc: cc,
		disableServiceConfig: opts.DisableServiceConfig,
	}

	if err := cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{{
			Addr: ":8001",
			ServerName: "main",
		}},
	}); err != nil {
		return nil, err
	}

	return cr, nil
}

func (b *cellsBuilder) Scheme() string {
	return "cells"
}

func (cr *cellsResolver) ResolveNow(opt resolver.ResolveNowOptions) {
	fmt.Println("Resolving now ?")
}

func (cr *cellsResolver) Close() {
}

func parseTarget(target string) (host, port, name string, err error) {

	fmt.Printf("target uri: %v\n", target)
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