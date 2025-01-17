// Code generated by mockery v1.0.0. DO NOT EDIT.

package upgrade_agent

import mock "github.com/stretchr/testify/mock"

// MockDependencies is an autogenerated mock type for the Dependencies type
type MockDependencies struct {
	mock.Mock
}

// ExecutePrivileged provides a mock function with given fields: command, args
func (_m *MockDependencies) ExecutePrivileged(command string, args ...string) (string, string, int) {
	_va := make([]interface{}, len(args))
	for _i := range args {
		_va[_i] = args[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, command)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, ...string) string); ok {
		r0 = rf(command, args...)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, ...string) string); ok {
		r1 = rf(command, args...)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 int
	if rf, ok := ret.Get(2).(func(string, ...string) int); ok {
		r2 = rf(command, args...)
	} else {
		r2 = ret.Get(2).(int)
	}

	return r0, r1, r2
}
