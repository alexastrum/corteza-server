// Code generated by MockGen. DO NOT EDIT.
// Source: messaging/internal/service/message.go

// Package service is a generated GoMock package.
package service

import (
	context "context"
	types "github.com/crusttech/crust/messaging/types"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMessageService is a mock of MessageService interface
type MockMessageService struct {
	ctrl     *gomock.Controller
	recorder *MockMessageServiceMockRecorder
}

// MockMessageServiceMockRecorder is the mock recorder for MockMessageService
type MockMessageServiceMockRecorder struct {
	mock *MockMessageService
}

// NewMockMessageService creates a new mock instance
func NewMockMessageService(ctrl *gomock.Controller) *MockMessageService {
	mock := &MockMessageService{ctrl: ctrl}
	mock.recorder = &MockMessageServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMessageService) EXPECT() *MockMessageServiceMockRecorder {
	return m.recorder
}

// With mocks base method
func (m *MockMessageService) With(ctx context.Context) MessageService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "With", ctx)
	ret0, _ := ret[0].(MessageService)
	return ret0
}

// With indicates an expected call of With
func (mr *MockMessageServiceMockRecorder) With(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockMessageService)(nil).With), ctx)
}

// Find mocks base method
func (m *MockMessageService) Find(filter *types.MessageFilter) (types.MessageSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", filter)
	ret0, _ := ret[0].(types.MessageSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find
func (mr *MockMessageServiceMockRecorder) Find(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockMessageService)(nil).Find), filter)
}

// FindThreads mocks base method
func (m *MockMessageService) FindThreads(filter *types.MessageFilter) (types.MessageSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindThreads", filter)
	ret0, _ := ret[0].(types.MessageSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindThreads indicates an expected call of FindThreads
func (mr *MockMessageServiceMockRecorder) FindThreads(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindThreads", reflect.TypeOf((*MockMessageService)(nil).FindThreads), filter)
}

// Create mocks base method
func (m *MockMessageService) Create(messages *types.Message) (*types.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", messages)
	ret0, _ := ret[0].(*types.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockMessageServiceMockRecorder) Create(messages interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockMessageService)(nil).Create), messages)
}

// Update mocks base method
func (m *MockMessageService) Update(messages *types.Message) (*types.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", messages)
	ret0, _ := ret[0].(*types.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update
func (mr *MockMessageServiceMockRecorder) Update(messages interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockMessageService)(nil).Update), messages)
}

// React mocks base method
func (m *MockMessageService) React(messageID uint64, reaction string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "React", messageID, reaction)
	ret0, _ := ret[0].(error)
	return ret0
}

// React indicates an expected call of React
func (mr *MockMessageServiceMockRecorder) React(messageID, reaction interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "React", reflect.TypeOf((*MockMessageService)(nil).React), messageID, reaction)
}

// RemoveReaction mocks base method
func (m *MockMessageService) RemoveReaction(messageID uint64, reaction string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveReaction", messageID, reaction)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveReaction indicates an expected call of RemoveReaction
func (mr *MockMessageServiceMockRecorder) RemoveReaction(messageID, reaction interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveReaction", reflect.TypeOf((*MockMessageService)(nil).RemoveReaction), messageID, reaction)
}

// MarkAsRead mocks base method
func (m *MockMessageService) MarkAsRead(channelID, threadID, lastReadMessageID uint64) (uint32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkAsRead", channelID, threadID, lastReadMessageID)
	ret0, _ := ret[0].(uint32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarkAsRead indicates an expected call of MarkAsRead
func (mr *MockMessageServiceMockRecorder) MarkAsRead(channelID, threadID, lastReadMessageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkAsRead", reflect.TypeOf((*MockMessageService)(nil).MarkAsRead), channelID, threadID, lastReadMessageID)
}

// Pin mocks base method
func (m *MockMessageService) Pin(messageID uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Pin", messageID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Pin indicates an expected call of Pin
func (mr *MockMessageServiceMockRecorder) Pin(messageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Pin", reflect.TypeOf((*MockMessageService)(nil).Pin), messageID)
}

// RemovePin mocks base method
func (m *MockMessageService) RemovePin(messageID uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemovePin", messageID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemovePin indicates an expected call of RemovePin
func (mr *MockMessageServiceMockRecorder) RemovePin(messageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemovePin", reflect.TypeOf((*MockMessageService)(nil).RemovePin), messageID)
}

// Bookmark mocks base method
func (m *MockMessageService) Bookmark(messageID uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Bookmark", messageID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Bookmark indicates an expected call of Bookmark
func (mr *MockMessageServiceMockRecorder) Bookmark(messageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Bookmark", reflect.TypeOf((*MockMessageService)(nil).Bookmark), messageID)
}

// RemoveBookmark mocks base method
func (m *MockMessageService) RemoveBookmark(messageID uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveBookmark", messageID)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveBookmark indicates an expected call of RemoveBookmark
func (mr *MockMessageServiceMockRecorder) RemoveBookmark(messageID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveBookmark", reflect.TypeOf((*MockMessageService)(nil).RemoveBookmark), messageID)
}

// Delete mocks base method
func (m *MockMessageService) Delete(ID uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ID)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockMessageServiceMockRecorder) Delete(ID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockMessageService)(nil).Delete), ID)
}