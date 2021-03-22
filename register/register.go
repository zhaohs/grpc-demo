package register

var (
	// m is a map from name to balancer builder.
	m = make(map[string]Register)
)

type Register interface {

	Connect()
}

