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
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"grpc-demo/register/nacos"
	"strconv"
	"time"

	"google.golang.org/grpc"
	pb "grpc-demo/pb/helloworld"
)
var namingClient naming_client.INamingClient


func main() {

	err := nacos.NewUserResolverBuilder("/Users/yc/logs")
	if err != nil {
		fmt.Println(err)
	}
	roundrobinConn, err := grpc.Dial(
		fmt.Sprintf("%s://%s/%s", nacos.USER_CLUSTER_NAME, nacos.USER_GROUP_NAEME, nacos.USER_SERVICE_NAME),
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

}

