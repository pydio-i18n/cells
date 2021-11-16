package main

import (
	"context"
	"fmt"
	"github.com/micro/micro/v3/proto/config"
	"github.com/micro/micro/v3/proto/registry"
	"github.com/pydio/cells/v4/common/proto/tree"
	"google.golang.org/grpc"
	"log"

	_ "github.com/pydio/cells/v4/common/server/grpc"
)

func main() {
	c, err := grpc.Dial("cells://127.0.0.1:8001/main", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Panic("error dialing", err)
	}

	/*
	confCli := config.NewConfigClient(c)
	if err := setConfig(confCli); err!= nil {
		log.Panic(err)
	}

	if err := getConfig(confCli); err!= nil {
		log.Panic(err)
	}

	regCli := registry.NewRegistryClient(c)
	go func() {
		err := watchRegistry(regCli)
		if err != nil {
			log.Panic("Error in watch reg ", err)
		}
	}()
	if err := setRegistry(regCli); err!= nil {
		log.Panic(err)
	}

	if err := getRegistry(regCli); err!= nil {
		log.Panic(err)
	}
	 */

	nodeProviderCli := tree.NewNodeProviderClient(c)
	if err := readNode(nodeProviderCli); err != nil {
		log.Panic(err)
	}
}

func setConfig(cli config.ConfigClient) error {
	req := &config.SetRequest{
		Namespace: "config",
		Path: "this/is/a/test",
		Value: &config.Value{Data: "my value"},
	}

	resp, err := cli.Set(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func getConfig(cli config.ConfigClient) error {
	req := &config.GetRequest{}

	resp, err := cli.Get(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func setRegistry(cli registry.RegistryClient) error {
	req := &registry.Service{
		Name: "testing",
	}

	resp, err := cli.Register(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func getRegistry(cli registry.RegistryClient) error {
	req := &registry.ListRequest{}

	resp, err := cli.ListServices(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func watchRegistry(cli registry.RegistryClient) error {
	req := &registry.WatchRequest{}

	resp, err := cli.Watch(context.Background(), req)
	if err != nil {
		return err
	}

	for {
		res, err := resp.Recv()
		if err != nil {
			return err
		}

		fmt.Println(res)
	}
}

func readNode(cli tree.NodeProviderClient) error {
	req := &tree.ReadNodeRequest{}

	resp, err := cli.ReadNode(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}