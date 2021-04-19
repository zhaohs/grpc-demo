package nacos

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"google.golang.org/grpc/resolver"
	"strconv"
)

type Base struct {
	namingClient naming_client.INamingClient
	clusterName string
	serviceName string
	groupName string
	logDir string
}
// 注册中心
type BaseRegister struct {
	Base
	ip string
	port uint64
}

func(register *BaseRegister) Register() error {
	_, err := register.namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          register.ip,
		Port:        register.port,
		ServiceName: register.serviceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		//Metadata:    map[string]string{"idc":"xinjiapo"},
		ClusterName: register.clusterName, // 默认值DEFAULT
		GroupName:   register.groupName,   // 默认值DEFAULT_GROUP
	})
	return err
}


func(register *BaseRegister) UnRegister() error {
	_, err := register.namingClient.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          register.ip,
		Port:        register.port,
		ServiceName: register.serviceName,
		Ephemeral:   true,
		Cluster: register.clusterName, // 默认值DEFAULT
		GroupName:   register.groupName,   // 默认值DEFAULT_GROUP
	})
	return err
}

// 生成builder
type BaseResolverBuilder struct {
	Base
}

type BaseResolver struct {
	namingClient naming_client.INamingClient
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}


func (base *Base) initRegisterClient() error {
	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(""), //当namespace是public时，此处填空字符串。
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(base.logDir),
		constant.WithCacheDir(base.logDir),
		constant.WithRotateTime("24h"),
		constant.WithMaxAge(3),
		constant.WithLogLevel("info"),
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
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		return err
	}
	base.namingClient = namingClient
	return nil
}

func(builder *BaseResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	addrs := make([]string, 0)

	instances, err := builder.namingClient.SelectInstances(vo.SelectInstancesParam{
		ServiceName: target.Endpoint,
		GroupName:   target.Authority,             // 默认值DEFAULT_GROUP
		Clusters:    []string{target.Scheme}, // 默认值DEFAULT
		HealthyOnly: true,
	})
	fmt.Println(instances)
	if err == nil {
		for _, instance := range instances{
			addrs = append(addrs, instance.Ip + ":" + strconv.FormatInt(int64(instance.Port), 10))
		}
		fmt.Println(addrs)
		baseResolver := new(BaseResolver)
		baseResolver.addrsStore = map[string][]string{
			baseResolver.target.Endpoint: addrs,
		}
		baseResolver.namingClient = builder.namingClient
		baseResolver.cc = cc

		baseResolver.updateStatus()
		go baseResolver.watcher()
		return baseResolver, nil
	}
	return nil, err

}

func (builder *BaseResolverBuilder) Scheme() string { return builder.clusterName }

func (baseResolver *BaseResolver) watcher() {

	baseResolver.namingClient.Subscribe(&vo.SubscribeParam{
		ServiceName: baseResolver.target.Scheme,
		GroupName:   baseResolver.target.Endpoint,             // 默认值DEFAULT_GROUP
		Clusters:    []string{baseResolver.target.Authority}, // 默认值DEFAULT
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			addrs := make([]string, 0)
			for _, instance := range services{
				addrs = append(addrs, instance.Ip + ":" + strconv.FormatInt(int64(instance.Port), 10))
			}
			baseResolver.addrsStore = map[string][]string{
				baseResolver.target.Endpoint: addrs,
			}
			baseResolver.updateStatus()
		},
	})
}

func (baseResolver *BaseResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (baseResolver *BaseResolver) Close(){}

func (baseResolver *BaseResolver) updateStatus() {
	addrStrs := baseResolver.addrsStore[baseResolver.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	baseResolver.cc.UpdateState(resolver.State{Addresses: addrs})
}