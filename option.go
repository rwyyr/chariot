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

// Option is an option one can provide to the New function.
type Option func(*options)

// RunOption is an option one can provide to the App's Run method.
type RunOption func(*options)

// ShutdownOption is an option that may be passed to the 'App.Shutdown' method.
type ShutdownOption func(*options)

// With provides initializers.
func With(initializers ...interface{}) Option {
	return func(options *options) {
		options.initializers = append(options.initializers, initializers...)
	}
}

// WithComponents provides a component as a value, not via a constructor. Note, however, that
// because of intricacies of interface assignment one can't provide a component of an interface
// type this way. Resort to using a constructor to bypass the limitation.
func WithComponents(components ...interface{}) Option {
	return func(options *options) {
		options.components = append(options.components, components...)
	}
}

// WithSignals provides additional signals to extend the list that controls the behavior of a
// prepackaged context.
func WithSignals(signals ...os.Signal) Option {
	return func(options *options) {
		options.signals = append(options.signals, signals...)
	}
}

// WithContext provides a replacement to a prepackaged context for the duration of the function.
// Note, however, that the latter is still taken into account even if a replacement is provided.
func WithContext(ctx context.Context) Option {
	return func(options *options) {
		options.ctx = ctx
	}
}

// WithOptions provides a combination of options.
func WithOptions(funcOptions ...func(*options)) Option {
	return func(options *options) {
		for _, option := range funcOptions {
			option(options)
		}
	}
}

// WithRunContext provides a replacement to a context provided to the components. Otherwise, the
// one associated with an app is used. Note, however, that the latter is still taken into account
// even if a replacement is provided.
func WithRunContext(ctx context.Context) RunOption {
	return func(options *options) {
		options.ctx = ctx
	}
}

// WithShutdownContext provides an alternative context to be used as a parent context for the
// context passed to shutdowners. Without the option, the context associated with an app acts as a
// parent one. It doesn't cease to be taken into account though when the option is provided.
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
