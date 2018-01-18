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

package fxcontext

import (
	gcontext "context"

	"go.uber.org/fx"
	"go.uber.org/fx/service"
	"go.uber.org/fx/ulog"
)

type contextKey int

type store struct {
	log ulog.Log
}

const _fxContextStore contextKey = iota

var _ fx.Context = &Context{}

// Context embeds Host and Go stdlib context for use
type Context struct {
	gcontext.Context
}

// New always returns Context for use in the service
func New(ctx gcontext.Context, host service.Host) fx.Context {
	if host != nil {
		ctx = gcontext.WithValue(ctx, _fxContextStore, store{
			log: host.Logger(),
		})
	}
	return &Context{
		Context: ctx,
	}
}

// Logger returns context based logger. If logger is absent from the context,
// the function updates the context with a new context based logger
func (c *Context) Logger() ulog.Log {
	return c.getStore().log
}

func (c *Context) getStore() store {
	fxctxStore := c.Context.Value(_fxContextStore)
	if fxctxStore == nil {
		fxctxStore = store{
			log: ulog.Logger(),
		}
		c.Context = gcontext.WithValue(c.Context, _fxContextStore, fxctxStore)
	}
	return fxctxStore.(store)
}
