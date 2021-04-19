package nacos

type UserRegister struct {
	BaseRegister
}

func NewUserRegister(ip string, port uint64, logDir string) (*UserRegister, error){
	userRegister := new(UserRegister)
	userRegister.clusterName = USER_CLUSTER_NAME
	userRegister.groupName = USER_GROUP_NAEME
	userRegister.serviceName = USER_SERVICE_NAME
	userRegister.logDir = logDir
	userRegister.ip = ip
	userRegister.port = port
	err := userRegister.initRegisterClient()

	return userRegister, err
}
