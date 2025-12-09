package main

import (
	"context"
	api "mgr-demo2/kitex_gen/thrift/kitex/api"
)

// AppServiceImpl implements the last service interface defined in the IDL.
type AppServiceImpl struct{}

// Action implements the AppServiceImpl interface.
func (s *AppServiceImpl) Action(ctx context.Context, req *api.MgrReq) (resp *api.Response, err error) {
	// TODO: Your code here...
	return
}
