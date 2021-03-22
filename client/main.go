/*
 *
 * Copyright 2018 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Binary client is an example client.
package main

import (
	"context"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	pb "grpc-demo/pb/helloworld"
)

const (
	exampleScheme      = "example"
	exampleServiceName = "/zktest"
)
var zkPath = "/zktest"
var conn *zk.Conn
func main() {
	var err error
	conn, _, err = zk.Connect([]string{"127.0.0.1:2181"}, time.Second*60)//测试所以写的60s
	if err != nil {
		fmt.Println(err)
		return
	}

	roundrobinConn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // This sets the initial balancing policy.
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
/*
	roundrobinConn, err := grpc.Dial(
		"localhost:50051",
		//grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // This sets the initial balancing policy.
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)*/
	defer roundrobinConn.Close()
	pbClient := pb.NewGreeterClient(roundrobinConn)
	for i := 0; i < 50; i++ {
		r, err := pbClient.SayHello(context.Background(),&pb.HelloRequest{Name:"aaa" + strconv.Itoa(i)})
		fmt.Println(r.Message)
		fmt.Println(err)
		time.Sleep(time.Second)
	}




}

// Following is an example name resolver implementation. Read the name
// resolution example to learn more about it.

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addrs := make([]string, 0)
	if ips, _, err := conn.Children(zkPath); err == nil {
		addrs = ips
	}
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: addrs,
		},
	}
	r.updateStatus()
	go r.watcher()
	return r, nil
}
func (*exampleResolverBuilder) Scheme() string { return exampleScheme }

type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) updateStatus() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *exampleResolver) watcher() {


	children, _, child_ch, err := conn.ChildrenW(zkPath)
	if err != nil {
		fmt.Println(err)

		//return
	}
	fmt.Println(children)

	for {
		select {
		case ch_event := <-child_ch:
			{
				if ch_event.Type == zk.EventNodeCreated {
					fmt.Printf("has node[%s] detete\n", ch_event.Path)
					children, _, child_ch, err = conn.ChildrenW(zkPath)
				} else if ch_event.Type == zk.EventNodeDeleted {
					fmt.Printf("has new node[%d] create\n", ch_event.Path)
					children, _, child_ch, err = conn.ChildrenW(zkPath)
				} else if ch_event.Type == zk.EventNodeDataChanged {
					fmt.Printf("has node[%d] data changed\n", ch_event.Path)
					children, _, child_ch, err = conn.ChildrenW(zkPath)
				} else if ch_event.Type == zk.EventNodeChildrenChanged {
					fmt.Printf("has children node[%d] data changed\n", ch_event.Path)
					children, _, child_ch, err = conn.ChildrenW(zkPath)
				}
			}
		}
		if len(children) != len(r.addrsStore[r.target.Endpoint]) {
			fmt.Println(fmt.Sprintf("origin %d  new %d", len(r.addrsStore[r.target.Endpoint]), len(children)))
			r.addrsStore[r.target.Endpoint] = children
			r.updateStatus()

		}
		fmt.Println("aaa")
		time.Sleep(time.Millisecond * 2)
	}

}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}

func init() {
	resolver.Register(&exampleResolverBuilder{})
}
