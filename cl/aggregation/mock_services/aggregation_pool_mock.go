// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erigontech/erigon/cl/aggregation (interfaces: AggregationPool)
//
// Generated by this command:
//
//	mockgen -typed=true -destination=./mock_services/aggregation_pool_mock.go -package=mock_services . AggregationPool
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	common "github.com/erigontech/erigon/erigon-lib/common"
	solid "github.com/erigontech/erigon/cl/cltypes/solid"
	gomock "go.uber.org/mock/gomock"
)

// MockAggregationPool is a mock of AggregationPool interface.
type MockAggregationPool struct {
	ctrl     *gomock.Controller
	recorder *MockAggregationPoolMockRecorder
	isgomock struct{}
}

// MockAggregationPoolMockRecorder is the mock recorder for MockAggregationPool.
type MockAggregationPoolMockRecorder struct {
	mock *MockAggregationPool
}

// NewMockAggregationPool creates a new mock instance.
func NewMockAggregationPool(ctrl *gomock.Controller) *MockAggregationPool {
	mock := &MockAggregationPool{ctrl: ctrl}
	mock.recorder = &MockAggregationPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAggregationPool) EXPECT() *MockAggregationPoolMockRecorder {
	return m.recorder
}

// AddAttestation mocks base method.
func (m *MockAggregationPool) AddAttestation(att *solid.Attestation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddAttestation", att)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddAttestation indicates an expected call of AddAttestation.
func (mr *MockAggregationPoolMockRecorder) AddAttestation(att any) *MockAggregationPoolAddAttestationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddAttestation", reflect.TypeOf((*MockAggregationPool)(nil).AddAttestation), att)
	return &MockAggregationPoolAddAttestationCall{Call: call}
}

// MockAggregationPoolAddAttestationCall wrap *gomock.Call
type MockAggregationPoolAddAttestationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAggregationPoolAddAttestationCall) Return(arg0 error) *MockAggregationPoolAddAttestationCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAggregationPoolAddAttestationCall) Do(f func(*solid.Attestation) error) *MockAggregationPoolAddAttestationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAggregationPoolAddAttestationCall) DoAndReturn(f func(*solid.Attestation) error) *MockAggregationPoolAddAttestationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetAggregatationByRoot mocks base method.
func (m *MockAggregationPool) GetAggregatationByRoot(root common.Hash) *solid.Attestation {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAggregatationByRoot", root)
	ret0, _ := ret[0].(*solid.Attestation)
	return ret0
}

// GetAggregatationByRoot indicates an expected call of GetAggregatationByRoot.
func (mr *MockAggregationPoolMockRecorder) GetAggregatationByRoot(root any) *MockAggregationPoolGetAggregatationByRootCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAggregatationByRoot", reflect.TypeOf((*MockAggregationPool)(nil).GetAggregatationByRoot), root)
	return &MockAggregationPoolGetAggregatationByRootCall{Call: call}
}

// MockAggregationPoolGetAggregatationByRootCall wrap *gomock.Call
type MockAggregationPoolGetAggregatationByRootCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAggregationPoolGetAggregatationByRootCall) Return(arg0 *solid.Attestation) *MockAggregationPoolGetAggregatationByRootCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAggregationPoolGetAggregatationByRootCall) Do(f func(common.Hash) *solid.Attestation) *MockAggregationPoolGetAggregatationByRootCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAggregationPoolGetAggregatationByRootCall) DoAndReturn(f func(common.Hash) *solid.Attestation) *MockAggregationPoolGetAggregatationByRootCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
