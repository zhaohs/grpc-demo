/*
 *
 * Copyright 2015 gRPC authors.
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

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	pb "grpc-demo/pb/helloworld"
	"time"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	fmt.Println(in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName() + "_" + port}, nil
}

var conn *zk.Conn
var zkPath = "/zktest"
func main() {
	var err error
	conn, _, err = zk.Connect([]string{"127.0.0.1:2181"}, time.Second*60)//测试所以写的60s
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	Register()
	go func() {
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
		for {
			<-ch
			fmt.Println("quit")
			Unregister("localhost:50051")
			os.Exit(0)
		}
	}()
	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func Register() {

	if err := createPath(zkPath + "/localhost:50051", nil, conn); err != nil {
		fmt.Println(err)
		panic(err)
	}


}

//删除注册的节点
func Unregister(registerPath string) {
	registerPath = zkPath + "/" + registerPath
	if conn != nil {
		exist, _, err := conn.Exists(registerPath)
		if exist {
			err = conn.Delete(registerPath, 0)
			if err != nil {
				fmt.Println("Delete ZK Node:", err)
			}
		}
		//conn.Close()//如果节点是自动注册的，则使用该操作，直接Close掉conn即可。
	}
}

func createPath(path string, data []byte, client *zk.Conn) error {
	exists, _, err := client.Exists(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if exists {
		return nil
	}

	name := "/"
	p := strings.Split(path, "/")

	for _, v := range p[1 : len(p)-1] {
		name += v
		e, _, _ := client.Exists(name)
		fmt.Println(name)
		if !e {
			_, err = client.Create(name, []byte{}, int32(0), zk.WorldACL(zk.PermAll))
			if err != nil {
				fmt.Println(err)
				return err
			}
		}
		name += "/"
	}

	_, err = client.Create(path, data, int32(0), zk.WorldACL(zk.PermAll))
	return err
}
