// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/erigontech/erigon/v3/polygon/heimdall (interfaces: HeimdallClient)
//
// Generated by this command:
//
//	mockgen -typed=true -destination=./client_mock.go -package=heimdall . HeimdallClient
//

// Package heimdall is a generated GoMock package.
package heimdall

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockHeimdallClient is a mock of HeimdallClient interface.
type MockHeimdallClient struct {
	ctrl     *gomock.Controller
	recorder *MockHeimdallClientMockRecorder
	isgomock struct{}
}

// MockHeimdallClientMockRecorder is the mock recorder for MockHeimdallClient.
type MockHeimdallClientMockRecorder struct {
	mock *MockHeimdallClient
}

// NewMockHeimdallClient creates a new mock instance.
func NewMockHeimdallClient(ctrl *gomock.Controller) *MockHeimdallClient {
	mock := &MockHeimdallClient{ctrl: ctrl}
	mock.recorder = &MockHeimdallClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHeimdallClient) EXPECT() *MockHeimdallClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockHeimdallClient) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockHeimdallClientMockRecorder) Close() *MockHeimdallClientCloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockHeimdallClient)(nil).Close))
	return &MockHeimdallClientCloseCall{Call: call}
}

// MockHeimdallClientCloseCall wrap *gomock.Call
type MockHeimdallClientCloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientCloseCall) Return() *MockHeimdallClientCloseCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientCloseCall) Do(f func()) *MockHeimdallClientCloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientCloseCall) DoAndReturn(f func()) *MockHeimdallClientCloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchCheckpoint mocks base method.
func (m *MockHeimdallClient) FetchCheckpoint(ctx context.Context, number int64) (*Checkpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchCheckpoint", ctx, number)
	ret0, _ := ret[0].(*Checkpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCheckpoint indicates an expected call of FetchCheckpoint.
func (mr *MockHeimdallClientMockRecorder) FetchCheckpoint(ctx, number any) *MockHeimdallClientFetchCheckpointCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCheckpoint", reflect.TypeOf((*MockHeimdallClient)(nil).FetchCheckpoint), ctx, number)
	return &MockHeimdallClientFetchCheckpointCall{Call: call}
}

// MockHeimdallClientFetchCheckpointCall wrap *gomock.Call
type MockHeimdallClientFetchCheckpointCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchCheckpointCall) Return(arg0 *Checkpoint, arg1 error) *MockHeimdallClientFetchCheckpointCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchCheckpointCall) Do(f func(context.Context, int64) (*Checkpoint, error)) *MockHeimdallClientFetchCheckpointCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchCheckpointCall) DoAndReturn(f func(context.Context, int64) (*Checkpoint, error)) *MockHeimdallClientFetchCheckpointCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchCheckpointCount mocks base method.
func (m *MockHeimdallClient) FetchCheckpointCount(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchCheckpointCount", ctx)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCheckpointCount indicates an expected call of FetchCheckpointCount.
func (mr *MockHeimdallClientMockRecorder) FetchCheckpointCount(ctx any) *MockHeimdallClientFetchCheckpointCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCheckpointCount", reflect.TypeOf((*MockHeimdallClient)(nil).FetchCheckpointCount), ctx)
	return &MockHeimdallClientFetchCheckpointCountCall{Call: call}
}

// MockHeimdallClientFetchCheckpointCountCall wrap *gomock.Call
type MockHeimdallClientFetchCheckpointCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchCheckpointCountCall) Return(arg0 int64, arg1 error) *MockHeimdallClientFetchCheckpointCountCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchCheckpointCountCall) Do(f func(context.Context) (int64, error)) *MockHeimdallClientFetchCheckpointCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchCheckpointCountCall) DoAndReturn(f func(context.Context) (int64, error)) *MockHeimdallClientFetchCheckpointCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchCheckpoints mocks base method.
func (m *MockHeimdallClient) FetchCheckpoints(ctx context.Context, page, limit uint64) ([]*Checkpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchCheckpoints", ctx, page, limit)
	ret0, _ := ret[0].([]*Checkpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchCheckpoints indicates an expected call of FetchCheckpoints.
func (mr *MockHeimdallClientMockRecorder) FetchCheckpoints(ctx, page, limit any) *MockHeimdallClientFetchCheckpointsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchCheckpoints", reflect.TypeOf((*MockHeimdallClient)(nil).FetchCheckpoints), ctx, page, limit)
	return &MockHeimdallClientFetchCheckpointsCall{Call: call}
}

// MockHeimdallClientFetchCheckpointsCall wrap *gomock.Call
type MockHeimdallClientFetchCheckpointsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchCheckpointsCall) Return(arg0 []*Checkpoint, arg1 error) *MockHeimdallClientFetchCheckpointsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchCheckpointsCall) Do(f func(context.Context, uint64, uint64) ([]*Checkpoint, error)) *MockHeimdallClientFetchCheckpointsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchCheckpointsCall) DoAndReturn(f func(context.Context, uint64, uint64) ([]*Checkpoint, error)) *MockHeimdallClientFetchCheckpointsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchFirstMilestoneNum mocks base method.
func (m *MockHeimdallClient) FetchFirstMilestoneNum(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchFirstMilestoneNum", ctx)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchFirstMilestoneNum indicates an expected call of FetchFirstMilestoneNum.
func (mr *MockHeimdallClientMockRecorder) FetchFirstMilestoneNum(ctx any) *MockHeimdallClientFetchFirstMilestoneNumCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchFirstMilestoneNum", reflect.TypeOf((*MockHeimdallClient)(nil).FetchFirstMilestoneNum), ctx)
	return &MockHeimdallClientFetchFirstMilestoneNumCall{Call: call}
}

// MockHeimdallClientFetchFirstMilestoneNumCall wrap *gomock.Call
type MockHeimdallClientFetchFirstMilestoneNumCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchFirstMilestoneNumCall) Return(arg0 int64, arg1 error) *MockHeimdallClientFetchFirstMilestoneNumCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchFirstMilestoneNumCall) Do(f func(context.Context) (int64, error)) *MockHeimdallClientFetchFirstMilestoneNumCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchFirstMilestoneNumCall) DoAndReturn(f func(context.Context) (int64, error)) *MockHeimdallClientFetchFirstMilestoneNumCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchLastNoAckMilestone mocks base method.
func (m *MockHeimdallClient) FetchLastNoAckMilestone(ctx context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchLastNoAckMilestone", ctx)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchLastNoAckMilestone indicates an expected call of FetchLastNoAckMilestone.
func (mr *MockHeimdallClientMockRecorder) FetchLastNoAckMilestone(ctx any) *MockHeimdallClientFetchLastNoAckMilestoneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchLastNoAckMilestone", reflect.TypeOf((*MockHeimdallClient)(nil).FetchLastNoAckMilestone), ctx)
	return &MockHeimdallClientFetchLastNoAckMilestoneCall{Call: call}
}

// MockHeimdallClientFetchLastNoAckMilestoneCall wrap *gomock.Call
type MockHeimdallClientFetchLastNoAckMilestoneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchLastNoAckMilestoneCall) Return(arg0 string, arg1 error) *MockHeimdallClientFetchLastNoAckMilestoneCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchLastNoAckMilestoneCall) Do(f func(context.Context) (string, error)) *MockHeimdallClientFetchLastNoAckMilestoneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchLastNoAckMilestoneCall) DoAndReturn(f func(context.Context) (string, error)) *MockHeimdallClientFetchLastNoAckMilestoneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchLatestSpan mocks base method.
func (m *MockHeimdallClient) FetchLatestSpan(ctx context.Context) (*Span, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchLatestSpan", ctx)
	ret0, _ := ret[0].(*Span)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchLatestSpan indicates an expected call of FetchLatestSpan.
func (mr *MockHeimdallClientMockRecorder) FetchLatestSpan(ctx any) *MockHeimdallClientFetchLatestSpanCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchLatestSpan", reflect.TypeOf((*MockHeimdallClient)(nil).FetchLatestSpan), ctx)
	return &MockHeimdallClientFetchLatestSpanCall{Call: call}
}

// MockHeimdallClientFetchLatestSpanCall wrap *gomock.Call
type MockHeimdallClientFetchLatestSpanCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchLatestSpanCall) Return(arg0 *Span, arg1 error) *MockHeimdallClientFetchLatestSpanCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchLatestSpanCall) Do(f func(context.Context) (*Span, error)) *MockHeimdallClientFetchLatestSpanCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchLatestSpanCall) DoAndReturn(f func(context.Context) (*Span, error)) *MockHeimdallClientFetchLatestSpanCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchMilestone mocks base method.
func (m *MockHeimdallClient) FetchMilestone(ctx context.Context, number int64) (*Milestone, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchMilestone", ctx, number)
	ret0, _ := ret[0].(*Milestone)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchMilestone indicates an expected call of FetchMilestone.
func (mr *MockHeimdallClientMockRecorder) FetchMilestone(ctx, number any) *MockHeimdallClientFetchMilestoneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchMilestone", reflect.TypeOf((*MockHeimdallClient)(nil).FetchMilestone), ctx, number)
	return &MockHeimdallClientFetchMilestoneCall{Call: call}
}

// MockHeimdallClientFetchMilestoneCall wrap *gomock.Call
type MockHeimdallClientFetchMilestoneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchMilestoneCall) Return(arg0 *Milestone, arg1 error) *MockHeimdallClientFetchMilestoneCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchMilestoneCall) Do(f func(context.Context, int64) (*Milestone, error)) *MockHeimdallClientFetchMilestoneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchMilestoneCall) DoAndReturn(f func(context.Context, int64) (*Milestone, error)) *MockHeimdallClientFetchMilestoneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchMilestoneCount mocks base method.
func (m *MockHeimdallClient) FetchMilestoneCount(ctx context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchMilestoneCount", ctx)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchMilestoneCount indicates an expected call of FetchMilestoneCount.
func (mr *MockHeimdallClientMockRecorder) FetchMilestoneCount(ctx any) *MockHeimdallClientFetchMilestoneCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchMilestoneCount", reflect.TypeOf((*MockHeimdallClient)(nil).FetchMilestoneCount), ctx)
	return &MockHeimdallClientFetchMilestoneCountCall{Call: call}
}

// MockHeimdallClientFetchMilestoneCountCall wrap *gomock.Call
type MockHeimdallClientFetchMilestoneCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchMilestoneCountCall) Return(arg0 int64, arg1 error) *MockHeimdallClientFetchMilestoneCountCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchMilestoneCountCall) Do(f func(context.Context) (int64, error)) *MockHeimdallClientFetchMilestoneCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchMilestoneCountCall) DoAndReturn(f func(context.Context) (int64, error)) *MockHeimdallClientFetchMilestoneCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchMilestoneID mocks base method.
func (m *MockHeimdallClient) FetchMilestoneID(ctx context.Context, milestoneID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchMilestoneID", ctx, milestoneID)
	ret0, _ := ret[0].(error)
	return ret0
}

// FetchMilestoneID indicates an expected call of FetchMilestoneID.
func (mr *MockHeimdallClientMockRecorder) FetchMilestoneID(ctx, milestoneID any) *MockHeimdallClientFetchMilestoneIDCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchMilestoneID", reflect.TypeOf((*MockHeimdallClient)(nil).FetchMilestoneID), ctx, milestoneID)
	return &MockHeimdallClientFetchMilestoneIDCall{Call: call}
}

// MockHeimdallClientFetchMilestoneIDCall wrap *gomock.Call
type MockHeimdallClientFetchMilestoneIDCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchMilestoneIDCall) Return(arg0 error) *MockHeimdallClientFetchMilestoneIDCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchMilestoneIDCall) Do(f func(context.Context, string) error) *MockHeimdallClientFetchMilestoneIDCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchMilestoneIDCall) DoAndReturn(f func(context.Context, string) error) *MockHeimdallClientFetchMilestoneIDCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchNoAckMilestone mocks base method.
func (m *MockHeimdallClient) FetchNoAckMilestone(ctx context.Context, milestoneID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchNoAckMilestone", ctx, milestoneID)
	ret0, _ := ret[0].(error)
	return ret0
}

// FetchNoAckMilestone indicates an expected call of FetchNoAckMilestone.
func (mr *MockHeimdallClientMockRecorder) FetchNoAckMilestone(ctx, milestoneID any) *MockHeimdallClientFetchNoAckMilestoneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchNoAckMilestone", reflect.TypeOf((*MockHeimdallClient)(nil).FetchNoAckMilestone), ctx, milestoneID)
	return &MockHeimdallClientFetchNoAckMilestoneCall{Call: call}
}

// MockHeimdallClientFetchNoAckMilestoneCall wrap *gomock.Call
type MockHeimdallClientFetchNoAckMilestoneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchNoAckMilestoneCall) Return(arg0 error) *MockHeimdallClientFetchNoAckMilestoneCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchNoAckMilestoneCall) Do(f func(context.Context, string) error) *MockHeimdallClientFetchNoAckMilestoneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchNoAckMilestoneCall) DoAndReturn(f func(context.Context, string) error) *MockHeimdallClientFetchNoAckMilestoneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchSpan mocks base method.
func (m *MockHeimdallClient) FetchSpan(ctx context.Context, spanID uint64) (*Span, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchSpan", ctx, spanID)
	ret0, _ := ret[0].(*Span)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchSpan indicates an expected call of FetchSpan.
func (mr *MockHeimdallClientMockRecorder) FetchSpan(ctx, spanID any) *MockHeimdallClientFetchSpanCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchSpan", reflect.TypeOf((*MockHeimdallClient)(nil).FetchSpan), ctx, spanID)
	return &MockHeimdallClientFetchSpanCall{Call: call}
}

// MockHeimdallClientFetchSpanCall wrap *gomock.Call
type MockHeimdallClientFetchSpanCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchSpanCall) Return(arg0 *Span, arg1 error) *MockHeimdallClientFetchSpanCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchSpanCall) Do(f func(context.Context, uint64) (*Span, error)) *MockHeimdallClientFetchSpanCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchSpanCall) DoAndReturn(f func(context.Context, uint64) (*Span, error)) *MockHeimdallClientFetchSpanCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchSpans mocks base method.
func (m *MockHeimdallClient) FetchSpans(ctx context.Context, page, limit uint64) ([]*Span, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchSpans", ctx, page, limit)
	ret0, _ := ret[0].([]*Span)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchSpans indicates an expected call of FetchSpans.
func (mr *MockHeimdallClientMockRecorder) FetchSpans(ctx, page, limit any) *MockHeimdallClientFetchSpansCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchSpans", reflect.TypeOf((*MockHeimdallClient)(nil).FetchSpans), ctx, page, limit)
	return &MockHeimdallClientFetchSpansCall{Call: call}
}

// MockHeimdallClientFetchSpansCall wrap *gomock.Call
type MockHeimdallClientFetchSpansCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchSpansCall) Return(arg0 []*Span, arg1 error) *MockHeimdallClientFetchSpansCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchSpansCall) Do(f func(context.Context, uint64, uint64) ([]*Span, error)) *MockHeimdallClientFetchSpansCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchSpansCall) DoAndReturn(f func(context.Context, uint64, uint64) ([]*Span, error)) *MockHeimdallClientFetchSpansCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchStateSyncEvent mocks base method.
func (m *MockHeimdallClient) FetchStateSyncEvent(ctx context.Context, id uint64) (*EventRecordWithTime, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchStateSyncEvent", ctx, id)
	ret0, _ := ret[0].(*EventRecordWithTime)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchStateSyncEvent indicates an expected call of FetchStateSyncEvent.
func (mr *MockHeimdallClientMockRecorder) FetchStateSyncEvent(ctx, id any) *MockHeimdallClientFetchStateSyncEventCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchStateSyncEvent", reflect.TypeOf((*MockHeimdallClient)(nil).FetchStateSyncEvent), ctx, id)
	return &MockHeimdallClientFetchStateSyncEventCall{Call: call}
}

// MockHeimdallClientFetchStateSyncEventCall wrap *gomock.Call
type MockHeimdallClientFetchStateSyncEventCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchStateSyncEventCall) Return(arg0 *EventRecordWithTime, arg1 error) *MockHeimdallClientFetchStateSyncEventCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchStateSyncEventCall) Do(f func(context.Context, uint64) (*EventRecordWithTime, error)) *MockHeimdallClientFetchStateSyncEventCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchStateSyncEventCall) DoAndReturn(f func(context.Context, uint64) (*EventRecordWithTime, error)) *MockHeimdallClientFetchStateSyncEventCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// FetchStateSyncEvents mocks base method.
func (m *MockHeimdallClient) FetchStateSyncEvents(ctx context.Context, fromId uint64, to time.Time, limit int) ([]*EventRecordWithTime, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchStateSyncEvents", ctx, fromId, to, limit)
	ret0, _ := ret[0].([]*EventRecordWithTime)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchStateSyncEvents indicates an expected call of FetchStateSyncEvents.
func (mr *MockHeimdallClientMockRecorder) FetchStateSyncEvents(ctx, fromId, to, limit any) *MockHeimdallClientFetchStateSyncEventsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchStateSyncEvents", reflect.TypeOf((*MockHeimdallClient)(nil).FetchStateSyncEvents), ctx, fromId, to, limit)
	return &MockHeimdallClientFetchStateSyncEventsCall{Call: call}
}

// MockHeimdallClientFetchStateSyncEventsCall wrap *gomock.Call
type MockHeimdallClientFetchStateSyncEventsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHeimdallClientFetchStateSyncEventsCall) Return(arg0 []*EventRecordWithTime, arg1 error) *MockHeimdallClientFetchStateSyncEventsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHeimdallClientFetchStateSyncEventsCall) Do(f func(context.Context, uint64, time.Time, int) ([]*EventRecordWithTime, error)) *MockHeimdallClientFetchStateSyncEventsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHeimdallClientFetchStateSyncEventsCall) DoAndReturn(f func(context.Context, uint64, time.Time, int) ([]*EventRecordWithTime, error)) *MockHeimdallClientFetchStateSyncEventsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
