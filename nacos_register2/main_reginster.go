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
	"errors"
	"fmt"
	"google.golang.org/grpc"
	pb "grpc-demo/pb/helloworld"
	"grpc-demo/register/nacos"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	port = uint64(50052)
	portStr = ":50052"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}
func GetMachineIp() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "" ,err
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no addirs")
}
// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	fmt.Println(in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName() + "_" + portStr}, nil
}

func main() {
	ip, err := GetMachineIp()
	fmt.Println(ip)
	fmt.Println(err)
	userRegister, err := nacos.NewUserRegister(ip, port,"/Users/yc/logs")
	fmt.Println(err)
	lis, err := net.Listen("tcp", portStr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	err = userRegister.Register()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		ch := make(chan os.Signal, 10)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
		for {
			<-ch
			fmt.Println("quit")
			userRegister.UnRegister()
			os.Exit(0)
		}
	}()
	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}


}

