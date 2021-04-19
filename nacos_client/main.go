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
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	pb "grpc-demo/pb/helloworld"
)
var namingClient naming_client.INamingClient
func init() {
	resolver.Register(&exampleResolverBuilder{})
	// 创建clientConfig的另一种方式
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""), //当namespace是public时，此处填空字符串。
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/Users/yc/logs"),
		constant.WithCacheDir("/Users/yc/logs"),
		constant.WithRotateTime("1h"),
		constant.WithMaxAge(3),
		constant.WithLogLevel("debug"),
	)
	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}
	var err error
	namingClient, err = clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	fmt.Println(err)

}
const (
	exampleScheme      = "grpc-demo"
	exampleServiceName = "user"
)

func main() {


	roundrobinConn, err := grpc.Dial(
		fmt.Sprintf("%s://grpc/%s", exampleScheme, exampleServiceName),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`), // This sets the initial balancing policy.
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	fmt.Println(err)
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


	err = namingClient.Unsubscribe(&vo.SubscribeParam{
		ServiceName: "grpc-demo",
		GroupName:   "user",             // 默认值DEFAULT_GROUP
		Clusters:    []string{"grpc"}, // 默认值DEFAULT
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			fmt.Println(services)
		},
	})
	fmt.Println(err)




}

// Following is an example name resolver implementation. Read the name
// resolution example to learn more about it.

type exampleResolverBuilder struct{}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addrs := make([]string, 0)

	instances, err := namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: target.Scheme,
		GroupName:   target.Endpoint,             // 默认值DEFAULT_GROUP
		Clusters:    []string{target.Authority}, // 默认值DEFAULT
		HealthyOnly: true,
	})
	if err == nil {
		for _, instance := range instances{
			addrs = append(addrs, instance.Ip + ":" + strconv.FormatInt(int64(instance.Port), 10))
		}

	}
	fmt.Println(err)
	fmt.Println(addrs)



	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			target.Endpoint: addrs,
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

	err := namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: r.target.Scheme,
		GroupName:   r.target.Endpoint,             // 默认值DEFAULT_GROUP
		Clusters:    []string{r.target.Authority}, // 默认值DEFAULT
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			addrs := make([]string, 0)
			for _, instance := range services{
				addrs = append(addrs, instance.Ip + ":" + strconv.FormatInt(int64(instance.Port), 10))
			}
			r.addrsStore = map[string][]string{
				r.target.Endpoint: addrs,
			}
			fmt.Println(addrs)
			fmt.Println("aaaaaa")
			fmt.Println(err)
			r.updateStatus()

		},
	})
	fmt.Println(err)

}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}

func init() {

}
