package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
	"time"
)
var zkPath = "/zktest"
func main() {
	conn, _, err := zk.Connect([]string{"127.0.0.1:2181"}, time.Second*60)//测试所以写的60s
	if err != nil {
		fmt.Println(err)
		return
	}

	/*if err := createPath(zkPath + "/test1", nil, conn); err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err := createPath(zkPath + "/test2", nil, conn); err != nil {
		fmt.Println(err)
		panic(err)
	}*/
	time.Sleep(time.Second * 2)
	if ips, _, err := conn.Children(zkPath); err == nil {
		fmt.Println("aaaaa")
		fmt.Println(ips)
	} else {
		fmt.Println("bbb")
		fmt.Println(err)
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
