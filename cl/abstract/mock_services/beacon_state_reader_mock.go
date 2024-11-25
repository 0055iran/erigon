// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erigontech/erigon/cl/abstract (interfaces: BeaconStateReader)
//
// Generated by this command:
//
//	mockgen -typed=true -destination=./mock_services/beacon_state_reader_mock.go -package=mock_services . BeaconStateReader
//

// Package mock_services is a generated GoMock package.
package mock_services

import (
	reflect "reflect"

	common "github.com/erigontech/erigon/erigon-lib/common"
	clparams "github.com/erigontech/erigon/cl/clparams"
	solid "github.com/erigontech/erigon/cl/cltypes/solid"
	gomock "go.uber.org/mock/gomock"
)

// MockBeaconStateReader is a mock of BeaconStateReader interface.
type MockBeaconStateReader struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconStateReaderMockRecorder
	isgomock struct{}
}

// MockBeaconStateReaderMockRecorder is the mock recorder for MockBeaconStateReader.
type MockBeaconStateReaderMockRecorder struct {
	mock *MockBeaconStateReader
}

// NewMockBeaconStateReader creates a new mock instance.
func NewMockBeaconStateReader(ctrl *gomock.Controller) *MockBeaconStateReader {
	mock := &MockBeaconStateReader{ctrl: ctrl}
	mock.recorder = &MockBeaconStateReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBeaconStateReader) EXPECT() *MockBeaconStateReaderMockRecorder {
	return m.recorder
}

// CommitteeCount mocks base method.
func (m *MockBeaconStateReader) CommitteeCount(epoch uint64) uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CommitteeCount", epoch)
	ret0, _ := ret[0].(uint64)
	return ret0
}

// CommitteeCount indicates an expected call of CommitteeCount.
func (mr *MockBeaconStateReaderMockRecorder) CommitteeCount(epoch any) *MockBeaconStateReaderCommitteeCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CommitteeCount", reflect.TypeOf((*MockBeaconStateReader)(nil).CommitteeCount), epoch)
	return &MockBeaconStateReaderCommitteeCountCall{Call: call}
}

// MockBeaconStateReaderCommitteeCountCall wrap *gomock.Call
type MockBeaconStateReaderCommitteeCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderCommitteeCountCall) Return(arg0 uint64) *MockBeaconStateReaderCommitteeCountCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderCommitteeCountCall) Do(f func(uint64) uint64) *MockBeaconStateReaderCommitteeCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderCommitteeCountCall) DoAndReturn(f func(uint64) uint64) *MockBeaconStateReaderCommitteeCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GenesisValidatorsRoot mocks base method.
func (m *MockBeaconStateReader) GenesisValidatorsRoot() common.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenesisValidatorsRoot")
	ret0, _ := ret[0].(common.Hash)
	return ret0
}

// GenesisValidatorsRoot indicates an expected call of GenesisValidatorsRoot.
func (mr *MockBeaconStateReaderMockRecorder) GenesisValidatorsRoot() *MockBeaconStateReaderGenesisValidatorsRootCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenesisValidatorsRoot", reflect.TypeOf((*MockBeaconStateReader)(nil).GenesisValidatorsRoot))
	return &MockBeaconStateReaderGenesisValidatorsRootCall{Call: call}
}

// MockBeaconStateReaderGenesisValidatorsRootCall wrap *gomock.Call
type MockBeaconStateReaderGenesisValidatorsRootCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderGenesisValidatorsRootCall) Return(arg0 common.Hash) *MockBeaconStateReaderGenesisValidatorsRootCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderGenesisValidatorsRootCall) Do(f func() common.Hash) *MockBeaconStateReaderGenesisValidatorsRootCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderGenesisValidatorsRootCall) DoAndReturn(f func() common.Hash) *MockBeaconStateReaderGenesisValidatorsRootCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetBeaconProposerIndexForSlot mocks base method.
func (m *MockBeaconStateReader) GetBeaconProposerIndexForSlot(slot uint64) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBeaconProposerIndexForSlot", slot)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBeaconProposerIndexForSlot indicates an expected call of GetBeaconProposerIndexForSlot.
func (mr *MockBeaconStateReaderMockRecorder) GetBeaconProposerIndexForSlot(slot any) *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBeaconProposerIndexForSlot", reflect.TypeOf((*MockBeaconStateReader)(nil).GetBeaconProposerIndexForSlot), slot)
	return &MockBeaconStateReaderGetBeaconProposerIndexForSlotCall{Call: call}
}

// MockBeaconStateReaderGetBeaconProposerIndexForSlotCall wrap *gomock.Call
type MockBeaconStateReaderGetBeaconProposerIndexForSlotCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall) Return(arg0 uint64, arg1 error) *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall) Do(f func(uint64) (uint64, error)) *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall) DoAndReturn(f func(uint64) (uint64, error)) *MockBeaconStateReaderGetBeaconProposerIndexForSlotCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// GetDomain mocks base method.
func (m *MockBeaconStateReader) GetDomain(domainType [4]byte, epoch uint64) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomain", domainType, epoch)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomain indicates an expected call of GetDomain.
func (mr *MockBeaconStateReaderMockRecorder) GetDomain(domainType, epoch any) *MockBeaconStateReaderGetDomainCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomain", reflect.TypeOf((*MockBeaconStateReader)(nil).GetDomain), domainType, epoch)
	return &MockBeaconStateReaderGetDomainCall{Call: call}
}

// MockBeaconStateReaderGetDomainCall wrap *gomock.Call
type MockBeaconStateReaderGetDomainCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderGetDomainCall) Return(arg0 []byte, arg1 error) *MockBeaconStateReaderGetDomainCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderGetDomainCall) Do(f func([4]byte, uint64) ([]byte, error)) *MockBeaconStateReaderGetDomainCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderGetDomainCall) DoAndReturn(f func([4]byte, uint64) ([]byte, error)) *MockBeaconStateReaderGetDomainCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ValidatorForValidatorIndex mocks base method.
func (m *MockBeaconStateReader) ValidatorForValidatorIndex(index int) (solid.Validator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorForValidatorIndex", index)
	ret0, _ := ret[0].(solid.Validator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidatorForValidatorIndex indicates an expected call of ValidatorForValidatorIndex.
func (mr *MockBeaconStateReaderMockRecorder) ValidatorForValidatorIndex(index any) *MockBeaconStateReaderValidatorForValidatorIndexCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorForValidatorIndex", reflect.TypeOf((*MockBeaconStateReader)(nil).ValidatorForValidatorIndex), index)
	return &MockBeaconStateReaderValidatorForValidatorIndexCall{Call: call}
}

// MockBeaconStateReaderValidatorForValidatorIndexCall wrap *gomock.Call
type MockBeaconStateReaderValidatorForValidatorIndexCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderValidatorForValidatorIndexCall) Return(arg0 solid.Validator, arg1 error) *MockBeaconStateReaderValidatorForValidatorIndexCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderValidatorForValidatorIndexCall) Do(f func(int) (solid.Validator, error)) *MockBeaconStateReaderValidatorForValidatorIndexCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderValidatorForValidatorIndexCall) DoAndReturn(f func(int) (solid.Validator, error)) *MockBeaconStateReaderValidatorForValidatorIndexCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ValidatorPublicKey mocks base method.
func (m *MockBeaconStateReader) ValidatorPublicKey(index int) (common.Bytes48, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorPublicKey", index)
	ret0, _ := ret[0].(common.Bytes48)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidatorPublicKey indicates an expected call of ValidatorPublicKey.
func (mr *MockBeaconStateReaderMockRecorder) ValidatorPublicKey(index any) *MockBeaconStateReaderValidatorPublicKeyCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorPublicKey", reflect.TypeOf((*MockBeaconStateReader)(nil).ValidatorPublicKey), index)
	return &MockBeaconStateReaderValidatorPublicKeyCall{Call: call}
}

// MockBeaconStateReaderValidatorPublicKeyCall wrap *gomock.Call
type MockBeaconStateReaderValidatorPublicKeyCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderValidatorPublicKeyCall) Return(arg0 common.Bytes48, arg1 error) *MockBeaconStateReaderValidatorPublicKeyCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderValidatorPublicKeyCall) Do(f func(int) (common.Bytes48, error)) *MockBeaconStateReaderValidatorPublicKeyCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderValidatorPublicKeyCall) DoAndReturn(f func(int) (common.Bytes48, error)) *MockBeaconStateReaderValidatorPublicKeyCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Version mocks base method.
func (m *MockBeaconStateReader) Version() clparams.StateVersion {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Version")
	ret0, _ := ret[0].(clparams.StateVersion)
	return ret0
}

// Version indicates an expected call of Version.
func (mr *MockBeaconStateReaderMockRecorder) Version() *MockBeaconStateReaderVersionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Version", reflect.TypeOf((*MockBeaconStateReader)(nil).Version))
	return &MockBeaconStateReaderVersionCall{Call: call}
}

// MockBeaconStateReaderVersionCall wrap *gomock.Call
type MockBeaconStateReaderVersionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBeaconStateReaderVersionCall) Return(arg0 clparams.StateVersion) *MockBeaconStateReaderVersionCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBeaconStateReaderVersionCall) Do(f func() clparams.StateVersion) *MockBeaconStateReaderVersionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBeaconStateReaderVersionCall) DoAndReturn(f func() clparams.StateVersion) *MockBeaconStateReaderVersionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
