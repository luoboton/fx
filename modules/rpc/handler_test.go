// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.uber.org/fx"
	"go.uber.org/fx/auth"
	"go.uber.org/fx/internal/fxcontext"
	"go.uber.org/fx/service"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/encoding/thrift"
	"go.uber.org/yarpc/transport"
)

type fakeEnveloper struct {
	serviceName string
}

func (f fakeEnveloper) MethodName() string {
	return f.serviceName
}

func (f fakeEnveloper) EnvelopeType() wire.EnvelopeType {
	return wire.Reply
}

func (f fakeEnveloper) ToWire() (wire.Value, error) {
	return wire.NewValueStruct(wire.Struct{}), nil
}

func UnaryFakeHandlerFunc(ctx fx.Context, meta yarpc.ReqMeta, val wire.Value) (thrift.Response, error) {
	return thrift.Response{
		Body: fakeEnveloper{},
	}, nil
}

func OnewayFakeHandlerFunc(ctx fx.Context, meta yarpc.ReqMeta, val wire.Value) error {
	return nil
}

func OnewayFakeHandlerFuncWithError(ctx fx.Context, meta yarpc.ReqMeta, val wire.Value) error {
	return errors.New("mocking error")
}

func TestWrapUnary(t *testing.T) {
	handlerFunc := WrapUnary(UnaryFakeHandlerFunc)
	assert.NotNil(t, handlerFunc)
	resp, err := handlerFunc.Handle(context.Background(), nil, wire.Value{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestWrapOneway(t *testing.T) {
	handlerFunc := WrapOneway(OnewayFakeHandlerFunc)
	assert.NotNil(t, handlerFunc)
	err := handlerFunc.HandleOneway(context.Background(), nil, wire.Value{})
	assert.NoError(t, err)
}

func TestWrapOneway_error(t *testing.T) {
	handlerFunc := WrapOneway(OnewayFakeHandlerFuncWithError)
	assert.NotNil(t, handlerFunc)
	err := handlerFunc.HandleOneway(context.Background(), nil, wire.Value{})
	assert.Error(t, err)
}

func TestInboundMiddleware_fxContext(t *testing.T) {
	unary := fxContextInboundMiddleware{
		Host: service.NullHost(),
	}
	err := unary.Handle(context.Background(), &transport.Request{}, nil, &fakeUnaryHandler{t: t})
	assert.Equal(t, "handle", err.Error())
}

func TestOnewayInboundMiddleware_fxContext(t *testing.T) {
	oneway := fxContextOnewayInboundMiddleware{
		Host: service.NullHost(),
	}
	err := oneway.HandleOneway(context.Background(), &transport.Request{}, &fakeOnewayHandler{t: t})
	assert.Equal(t, "oneway handle", err.Error())
}

func TestInboundMiddleware_auth(t *testing.T) {
	withAuthClient(t, nil, func() {
		unary := authInboundMiddleware{
			Host: service.NullHost(),
		}
		err := unary.Handle(context.Background(), &transport.Request{}, nil, &fakeUnaryHandler{t: t})
		assert.EqualError(t, err, "handle")
	})
}

func TestInboundMiddleware_authFailure(t *testing.T) {
	withAuthClient(t, auth.FakeFailureClient, func() {
		unary := authInboundMiddleware{
			Host: service.NullHost(),
		}
		err := unary.Handle(context.Background(), &transport.Request{}, nil, &fakeUnaryHandler{t: t})
		assert.EqualError(t, err, "Error authorizing the service")
	})
}

func TestOnewayInboundMiddleware_auth(t *testing.T) {
	withAuthClient(t, nil, func() {
		oneway := authOnewayInboundMiddleware{
			Host: service.NullHost(),
		}
		err := oneway.HandleOneway(context.Background(), &transport.Request{}, &fakeOnewayHandler{t: t})
		assert.EqualError(t, err, "oneway handle")

	})
}

func TestOnewayInboundMiddleware_authFailure(t *testing.T) {
	withAuthClient(t, auth.FakeFailureClient, func() {
		oneway := authOnewayInboundMiddleware{
			Host: service.NullHost(),
		}
		err := oneway.HandleOneway(context.Background(), &transport.Request{}, &fakeOnewayHandler{t: t})
		assert.EqualError(t, err, "Error authorizing the service")
	})
}

func withAuthClient(t *testing.T, registerFunc auth.RegisterFunc, fn func()) {
	auth.UnregisterClient()
	auth.RegisterClient(registerFunc)
	auth.SetupClient(service.NullHost())
	fn()
}

type fakeUnaryHandler struct {
	t *testing.T
}

func (f fakeUnaryHandler) Handle(ctx context.Context, _param1 *transport.Request, _param2 transport.ResponseWriter) error {
	assert.NotNil(f.t, fxcontext.Context{
		Context: ctx,
	})
	return errors.New("handle")
}

type fakeOnewayHandler struct {
	t *testing.T
}

func (f fakeOnewayHandler) HandleOneway(ctx context.Context, p *transport.Request) error {
	assert.NotNil(f.t, fxcontext.Context{
		Context: ctx,
	})
	return errors.New("oneway handle")
}
