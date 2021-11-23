package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/micro/micro/v3/proto/config"
	"github.com/micro/micro/v3/proto/registry"
	"github.com/pydio/cells/v4/common/proto/test"
	"github.com/pydio/cells/v4/common/proto/tree"
	"google.golang.org/grpc"

	_ "github.com/pydio/cells/v4/common/server/grpc"
)

func testNodes(c grpc.ClientConnInterface) {
	nodeReceiverCli := tree.NewNodeReceiverClient(c)
	if err := createNode(nodeReceiverCli); err != nil {
		log.Panic(err)
	}
	if err := updateNode(nodeReceiverCli); err != nil {
		log.Panic(err)
	}
	if err := deleteNode(nodeReceiverCli); err != nil {
		log.Panic(err)
	}

	nodeProviderCli := tree.NewNodeProviderClient(c)
	if err := readNode(nodeProviderCli); err != nil {
		log.Panic(err)
	}

	if err := listNodes(nodeProviderCli); err != nil {
		log.Panic(err)
	}

	nodeReceiverStreamCli := tree.NewNodeReceiverStreamClient(c)
	if err := createNodeStream(nodeReceiverStreamCli); err != nil {
		log.Panic(err)
	}
	if err := updateNodeStream(nodeReceiverStreamCli); err != nil {
		log.Panic(err)
	}
	if err := deleteNodeStream(nodeReceiverStreamCli); err != nil {
		log.Panic(err)
	}

	nodeProviderStreamCli := tree.NewNodeProviderStreamerClient(c)
	if err := readNodeStream(nodeProviderStreamCli); err != nil {
		log.Panic(err)
	}
}

func testObjectsTest(c grpc.ClientConnInterface) {
	testClient := test.NewTesterClient(c)
	resp, e := testClient.Run(context.Background(), &test.RunTestsRequest{})
	if e != nil {
		log.Panic(e)
	} else {
		fmt.Println(resp.Results)
	}
}

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

	testObjectsTest(c)
}

func setConfig(cli config.ConfigClient) error {
	req := &config.SetRequest{
		Namespace: "config",
		Path:      "this/is/a/test",
		Value:     &config.Value{Data: "my value"},
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

func createNode(cli tree.NodeReceiverClient) error {
	req := &tree.CreateNodeRequest{
		Node: &tree.Node{
			Path: "/test.txt",
		},
	}

	resp, err := cli.CreateNode(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func updateNode(cli tree.NodeReceiverClient) error {
	req := &tree.UpdateNodeRequest{
		From: &tree.Node{
			Path: "/test.txt",
		},
		To: &tree.Node{
			Path: "/test2.txt",
		},
	}

	resp, err := cli.UpdateNode(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func deleteNode(cli tree.NodeReceiverClient) error {
	req := &tree.DeleteNodeRequest{
		Node: &tree.Node{
			Path: "/test2.txt",
		},
	}

	resp, err := cli.DeleteNode(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func readNode(cli tree.NodeProviderClient) error {
	req := &tree.ReadNodeRequest{
		Node: &tree.Node{
			Path: "/",
		},
	}

	resp, err := cli.ReadNode(context.Background(), req)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func listNodes(cli tree.NodeProviderClient) error {
	req := &tree.ListNodesRequest{
		Node: &tree.Node{
			Path: "",
		},
		Recursive: true,
	}

	stream, err := cli.ListNodes(context.Background(), req)
	if err != nil {
		return err
	}

	defer stream.CloseSend()
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		fmt.Println("List nodes ", resp)
	}

	return nil
}

func createNodeStream(cli tree.NodeReceiverStreamClient) error {
	stream, err := cli.CreateNodeStream(context.Background())
	if err != nil {
		return err
	}

	req := &tree.CreateNodeRequest{
		Node: &tree.Node{
			Path: "/test.txt",
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}

			fmt.Println(resp)
		}
	}()

	if err := stream.Send(req); err != nil {
		return err
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	wg.Wait()

	return nil
}

func updateNodeStream(cli tree.NodeReceiverStreamClient) error {
	stream, err := cli.UpdateNodeStream(context.Background())
	if err != nil {
		return err
	}

	req := &tree.UpdateNodeRequest{
		From: &tree.Node{
			Path: "/test.txt",
		},
		To: &tree.Node{
			Path: "/test2.txt",
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}

			fmt.Println(resp)
		}
	}()

	if err := stream.Send(req); err != nil {
		return err
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	wg.Wait()

	return nil

}

func deleteNodeStream(cli tree.NodeReceiverStreamClient) error {
	stream, err := cli.DeleteNodeStream(context.Background())
	if err != nil {
		return err
	}

	req := &tree.DeleteNodeRequest{
		Node: &tree.Node{
			Path: "/test2.txt",
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}

			fmt.Println(resp)
		}
	}()

	if err := stream.Send(req); err != nil {
		return err
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	wg.Wait()

	return nil
}

func readNodeStream(cli tree.NodeProviderStreamerClient) error {
	stream, err := cli.ReadNodeStream(context.Background())
	if err != nil {
		return err
	}

	req := &tree.ReadNodeRequest{
		Node: &tree.Node{
			Path: "/",
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			resp, err := stream.Recv()
			if err != nil {
				return
			}

			fmt.Println(resp)
		}
	}()

	if err := stream.Send(req); err != nil {
		return err
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	wg.Wait()

	return nil
}
