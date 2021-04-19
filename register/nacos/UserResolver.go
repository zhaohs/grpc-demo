package nacos

import (
	"fmt"
	"google.golang.org/grpc/resolver"
)

type UserResolverBuilder struct {
	BaseResolverBuilder
}




const USER_SERVICE_NAME = "user"
const USER_CLUSTER_NAME = "orcatalk"
const USER_GROUP_NAEME = "user"
// env = stage/prod
func NewUserResolverBuilder(logDir string) (*UserResolverBuilder,error) {

	userResolver := new(UserResolverBuilder)
	userResolver.clusterName = USER_CLUSTER_NAME
	userResolver.groupName = USER_GROUP_NAEME
	userResolver.serviceName = USER_SERVICE_NAME
	userResolver.logDir = logDir
	err := userResolver.initRegisterClient()
	if err == nil {
		resolver.Register(userResolver)
	}
	return userResolver, err

}

func (resolver *UserResolverBuilder) GetTargetUrl()string {
	return fmt.Sprintf("%s://%s/%s", USER_CLUSTER_NAME, USER_GROUP_NAEME, USER_SERVICE_NAME)
}

