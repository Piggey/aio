package aio

type group struct {
	tag          string
	endpointOpts []endpointOpt
	endpoints    []endpoint
}
