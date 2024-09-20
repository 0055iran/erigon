// Code generated by MockGen. DO NOT EDIT.
// Source: ./peer_provider.go
//
// Generated by this command:
//
//	mockgen -typed=true -source=./peer_provider.go -destination=./peer_provider_mock.go -package=p2p
//

// Package p2p is a generated GoMock package.
package p2p

import (
	context "context"
	reflect "reflect"

	sentryproto "github.com/erigontech/erigon-lib/gointerfaces/sentryproto"
	gomock "go.uber.org/mock/gomock"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// MockpeerProvider is a mock of peerProvider interface.
type MockpeerProvider struct {
	ctrl     *gomock.Controller
	recorder *MockpeerProviderMockRecorder
}

// MockpeerProviderMockRecorder is the mock recorder for MockpeerProvider.
type MockpeerProviderMockRecorder struct {
	mock *MockpeerProvider
}

// NewMockpeerProvider creates a new mock instance.
func NewMockpeerProvider(ctrl *gomock.Controller) *MockpeerProvider {
	mock := &MockpeerProvider{ctrl: ctrl}
	mock.recorder = &MockpeerProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockpeerProvider) EXPECT() *MockpeerProviderMockRecorder {
	return m.recorder
}

// Peers mocks base method.
func (m *MockpeerProvider) Peers(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*sentryproto.PeersReply, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Peers", varargs...)
	ret0, _ := ret[0].(*sentryproto.PeersReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Peers indicates an expected call of Peers.
func (mr *MockpeerProviderMockRecorder) Peers(ctx, in any, opts ...any) *MockpeerProviderPeersCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, in}, opts...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Peers", reflect.TypeOf((*MockpeerProvider)(nil).Peers), varargs...)
	return &MockpeerProviderPeersCall{Call: call}
}

// MockpeerProviderPeersCall wrap *gomock.Call
type MockpeerProviderPeersCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockpeerProviderPeersCall) Return(arg0 *sentryproto.PeersReply, arg1 error) *MockpeerProviderPeersCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockpeerProviderPeersCall) Do(f func(context.Context, *emptypb.Empty, ...grpc.CallOption) (*sentryproto.PeersReply, error)) *MockpeerProviderPeersCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockpeerProviderPeersCall) DoAndReturn(f func(context.Context, *emptypb.Empty, ...grpc.CallOption) (*sentryproto.PeersReply, error)) *MockpeerProviderPeersCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}