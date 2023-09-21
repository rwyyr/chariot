// MIT License
//
// Copyright (c) 2023 Roman Homoliako
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package chariot

import (
	"context"
	"os"
)

// Option is an option that may be passed to the 'New' function.
type Option func(*options)

// RunOption is an option that may be passed to the 'App.Run' method.
type RunOption func(*options)

// ShutdownOption is an option that may be passed to the 'App.Shutdown' method.
type ShutdownOption func(*options)

// With provides initializers.
func With(initializers ...interface{}) Option {
	return func(options *options) {
		options.initializers = append(options.initializers, initializers...)
	}
}

// WithComponents allows to provide a component as a value, not via a constructor. Note, however, that because of how
// interfaces work in Go one can't provide a component of an interface type. To bypass the limitation use a constructor
// instead.
func WithComponents(components ...interface{}) Option {
	return func(options *options) {
		options.components = append(options.components, components...)
	}
}

// WithSignals provides other signals in addition to the default one to cancel the context associated with an app.
func WithSignals(signals ...os.Signal) Option {
	return func(options *options) {
		options.signals = append(options.signals, signals...)
	}
}

// WithInitContext provides an alternative context to be available as a component upon the initialization rather than
// the default one. The latter doesn't cease to be taken into account though. This means, the context provided this way
// will be cancelled as soon as the default one is. The option finds its place when, say, one needs to implement a
// timed initialization.
func WithInitContext(ctx context.Context) Option {
	return func(options *options) {
		options.ctx = ctx
	}
}

// WithOptions provides an amalgamation of options.
func WithOptions(funcOptions ...func(*options)) Option {
	return func(options *options) {
		for _, option := range funcOptions {
			option(options)
		}
	}
}

// WithRunContext provides an alternative context to be used as a parent context for the context passed to runners.
// Without the option, the context associated with an app acts as a parent one. It doesn't cease to be taken into
// account though when the option is provided.
func WithRunContext(ctx context.Context) RunOption {
	return func(options *options) {
		options.ctx = ctx
	}
}

// WithErrHandler provides an error handler.
func WithErrHandler(handler func(context.Context, error)) RunOption {
	return func(options *options) {
		options.handler = handler
	}
}

// WithShutdownContext provides an alternative context to be used as a parent context for the context passed to
// shutdowners. Without the option, the context associated with an app acts as a parent one. It doesn't cease to be
// taken into account though when the option is provided.
func WithShutdownContext(ctx context.Context) ShutdownOption {
	return func(options *options) {
		options.ctx = ctx
	}
}

type options struct {
	initializers []interface{}
	components   []interface{}
	signals      []os.Signal
	ctx          context.Context
	handler      func(context.Context, error)
}
