package nacos

import "google.golang.org/grpc/resolver"

type UserResolverBuilder struct {
	BaseResolverBuilder
}




const USER_SERVICE_NAME = "userProfile"
const USER_CLUSTER_NAME = "aaaaa"
const USER_GROUP_NAEME = "user"
// env = stage/prod
func NewUserResolverBuilder(logDir string) error {

	userResolver := new(UserResolverBuilder)
	userResolver.clusterName = USER_CLUSTER_NAME
	userResolver.groupName = USER_GROUP_NAEME
	userResolver.serviceName = USER_SERVICE_NAME
	userResolver.logDir = logDir
	err := userResolver.initRegisterClient()
	if err == nil {
		resolver.Register(userResolver)
	}
	return err

}

