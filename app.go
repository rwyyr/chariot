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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
)

// App is a DI container supplemented with a compact set of related logic aimed to facilitate the
// process of initialization of applications composed of multiple components or modules.
type App struct {
	ctx         context.Context
	cancel      func()
	components  map[reflect.Type]*component
	runners     []Runner
	shutdowners []Shutdowner
}

type (
	// Runner stands for any conformant component that is collected during its initialization by an
	// app so later to be concurrently run by the app upon invocation of its Run method.
	Runner interface {
		Run(context.Context) error
	}

	// Shutdowner stands for any conformant component that is collected during its initialization
	// by an app so to be invoked either when its Shutdown method is called or the app itself fails
	// to initalize.
	Shutdowner interface {
		Shutdown(context.Context)
	}
)

// FuncRunner is a quick way to introduce a Runner-conformant component.
type FuncRunner func(context.Context) error

// New initializes a new app. Its main work is, if given a list of initializers, to invoke them in
// that order that ensures dependencies between them are resolved correctly. An initializer is a
// function which takes 0..N components and returns 0..N components. The last value returned, if
// complies with the 'error' type, isn't treated as a component but as an initialization failure.
// Variadic params are ignored. Initializers that don't produce any components are grouped together
// and invoked after ones that do. The latter ones are called "constructors". The former ones
// received the name "inits". The process in some way resembles the way how vars and init funcs
// work in Go (what is initialized or called first). Components are objects of distinct types.
// There must not be a duplicating type, and an error is returned in case of any. A dependency
// between initializers is established when one takes a component returned by another. If a
// dependency is missing an error is returned. Circular dependencies are caught, and an error is
// returned in case of any. A context (of type 'context.Context') is provided out of the box. It is
// associated with the app and continues to be taken into account upon all operations with it even
// if they may require a separate context. It is cancelled either if the 'SIGINT' signal (or other
// registered signals, see the corresponding option) is caught or the app's 'Shutdown' method is
// called. Components that comply with the 'Runner' interface are collected and stored for later
// use with the 'Run' method. Components that comply with the 'Shutdowner' interface are collected
// and stored for later use with the 'Shutdown' method. The method is called when the function
// returns an error.
func New(funcOptions ...Option) (_ App, err error) {
	var options options
	for _, option := range funcOptions {
		option(&options)
	}

	app := App{
		components: make(map[reflect.Type]*component, len(options.initializers)+1),
	}

	app.initializeCtx(options.signals)
	cancel := app.setCtxComponent(options.ctx)
	defer cancel()
	defer app.resetCtxComponent()

	defer func() {
		if err == nil {
			return
		}
		var ctx context.Context
		app.Retrieve(&ctx)
		app.Shutdown(WithShutdownContext(ctx))
	}()

	inits, err := app.initializeComponents(
		app.mergeComponentsInitializers(options.components, options.initializers),
	)
	if err != nil {
		return App{}, err
	}
	if err := app.invokeInits(inits); err != nil {
		return App{}, err
	}

	return app, nil
}

// Run runs previously collected Runner-conformant components in a concurrent manner with respect
// to errors returned by them in the process. In case of any the context provided to them is
// cancelled and the method waits till other components finish their work. Errors returned at this
// stage are collected and an aggregated error is returned (placing the one that triggered the
// event at the head of the underlying list). In case there was no error the method returns nil.
func (a App) Run(funcOptions ...RunOption) error {
	var options options
	for _, option := range funcOptions {
		option(&options)
	}

	var (
		ctx    context.Context
		cancel func()
	)
	if options.ctx != nil {
		ctx, cancel = context.WithCancel(options.ctx)
		go func() {
			select {
			case <-a.ctx.Done():
				cancel()
			case <-ctx.Done():
			}
		}()
	} else {
		ctx, cancel = context.WithCancel(a.ctx)
	}
	defer cancel()

	var (
		finished  sync.WaitGroup
		runErrors = make(chan error, len(a.runners))
	)
	finished.Add(len(a.runners))
	for _, runner := range a.runners {
		go func(runner Runner) {
			defer finished.Done()
			if err := runner.Run(ctx); err != nil {
				runErrors <- err
			}
		}(runner)
	}
	go func() {
		finished.Wait()
		close(runErrors)
	}()
	err, ok := <-runErrors
	if !ok {
		return nil
	}
	cancel()
	subsequentErrors := make([]error, 0, len(a.runners)-1)
	for err := range runErrors {
		subsequentErrors = append(subsequentErrors, err)
	}

	return errors.Join(append([]error{err}, subsequentErrors...)...)
}

// Shutdown releases resources associated with an app and invokes Shutdowner-conformant components
// collected during the initialization of the app in the reverse order they were collected. The
// latter is akin to the common way of releasing resources of multiple objects in defer statements.
// Once shut down the app is rendered unusable afterwards.
func (a App) Shutdown(funcOptions ...ShutdownOption) {
	var options options
	for _, option := range funcOptions {
		option(&options)
	}

	defer a.cancel()

	var (
		ctx    context.Context
		cancel func()
	)
	if options.ctx != nil {
		ctx, cancel = context.WithCancel(options.ctx)
		go func() {
			select {
			case <-a.ctx.Done():
				cancel()
			case <-ctx.Done():
			}
		}()
	} else {
		ctx, cancel = context.WithCancel(a.ctx)
	}
	defer cancel()
	for i := len(a.shutdowners) - 1; i >= 0; i-- {
		a.shutdowners[i].Shutdown(ctx)
	}
}

// Retrieve retrieves a component. A valid value is a pointer to the type of the component.
func (a App) Retrieve(ptr interface{}) bool {
	value := reflect.ValueOf(ptr).Elem()

	component, found := a.components[value.Type()]
	if !found {
		return false
	}

	value.Set(component.value)

	return true
}

// Run delegates the execution to the receiver.
func (r FuncRunner) Run(ctx context.Context) error {
	return r(ctx)
}

func (a *App) initializeCtx(signals []os.Signal) {
	a.ctx, a.cancel = signal.NotifyContext(context.Background(), append(signals, os.Interrupt)...)
}

func (a App) setCtxComponent(ctx context.Context) func() {
	cancel := func() {}
	if ctx != nil {
		ctx, cancel = context.WithCancel(ctx)
		go func() {
			select {
			case <-a.ctx.Done():
				cancel()
			case <-ctx.Done():
			}
		}()
	} else {
		ctx = a.ctx
	}
	a.components[reflect.TypeOf((*context.Context)(nil)).Elem()] = &component{
		value: reflect.ValueOf(ctx),
	}

	return cancel
}

func (a App) resetCtxComponent() {
	a.components[reflect.TypeOf((*context.Context)(nil)).Elem()] = &component{
		value: reflect.ValueOf(a.ctx),
	}
}

func (App) mergeComponentsInitializers(components, initializers []interface{}) []interface{} {
	for _, component := range components {
		constructor := reflect.MakeFunc(
			reflect.FuncOf(
				nil,
				[]reflect.Type{
					reflect.TypeOf(component),
				},
				false,
			),
			func([]reflect.Value) []reflect.Value {
				return []reflect.Value{
					reflect.ValueOf(component),
				}
			},
		)
		initializers = append(initializers, constructor.Interface())
	}

	return initializers
}

func (a *App) collectComponents(initializers []interface{}) ([]initFunc, error) {
	var inits []initFunc
	for _, initializer := range initializers {
		initializerType := reflect.TypeOf(initializer)

		num := initializerType.NumIn()
		if initializerType.IsVariadic() {
			num--
		}
		var dependencies []reflect.Type
		for i := 0; i < num; i++ {
			dependencies = append(dependencies, initializerType.In(i))
		}

		num = initializerType.NumOut()
		if last := num - 1; last >= 0 && initializerType.Out(last).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			num = last
		}
		if num == 0 {
			inits = append(inits, initFunc{
				dependencies: dependencies,
				init:         reflect.ValueOf(initializer),
			})

			continue
		}

		for i := 0; i < num; i++ {
			componentType := initializerType.Out(i)

			if _, ok := a.components[componentType]; ok {
				return nil, fmt.Errorf("duplicating component '%s'", componentType)
			}

			a.components[componentType] = &component{
				dependencies: dependencies,
				constructor:  reflect.ValueOf(initializer),
			}
		}
	}

	return inits, nil
}

func (a *App) initializeComponents(initializers []interface{}) ([]initFunc, error) {
	inits, err := a.collectComponents(initializers)
	if err != nil {
		return nil, err
	}

	cycle := map[reflect.Type]struct{}{}
	for componentType, component := range a.components {
		cycle[componentType] = struct{}{}

		if err := a.initializeComponent(component, cycle); err != nil {
			return nil, err
		}

		delete(cycle, componentType)
	}

	return inits, nil
}

func (a *App) initializeComponent(component *component, cycle map[reflect.Type]struct{}) error {
	if component.value.IsValid() {
		return nil
	}

	ins, err := a.ins(component, cycle)
	if err != nil {
		return err
	}

	outs := component.constructor.Call(ins)

	last := outs[len(outs)-1]
	if last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !last.IsNil() {
			return last.Interface().(error)
		}
		outs = outs[:len(outs)-1]
	}

	for _, out := range outs {
		a.components[out.Type()].value = out

		if runner, ok := out.Interface().(Runner); ok {
			a.runners = append(a.runners, runner)
		}

		if shutdowner, ok := out.Interface().(Shutdowner); ok {
			a.shutdowners = append(a.shutdowners, shutdowner)
		}
	}

	return nil
}

func (a *App) ins(component *component, cycle map[reflect.Type]struct{}) ([]reflect.Value, error) {
	var ins []reflect.Value

	for _, dependencyType := range component.dependencies {
		dependency, ok := a.components[dependencyType]
		if !ok {
			return nil, fmt.Errorf("missing dependency '%s'", dependencyType)
		}

		if _, ok := cycle[dependencyType]; ok {
			return nil, errors.New("dependency cycle detected")
		}
		cycle[dependencyType] = struct{}{}

		if err := a.initializeComponent(dependency, cycle); err != nil {
			return nil, err
		}
		ins = append(ins, dependency.value)

		delete(cycle, dependencyType)
	}

	return ins, nil
}

func (a App) invokeInits(inits []initFunc) error {
	for _, init := range inits {
		var ins []reflect.Value
		for _, dependency := range init.dependencies {
			component, ok := a.components[dependency]
			if !ok {
				return fmt.Errorf("missing dependency '%s'", dependency)
			}

			ins = append(ins, component.value)
		}

		outs := init.init.Call(ins)

		if len(outs) == 0 {
			continue
		}

		if err := outs[0]; !err.IsNil() {
			return err.Interface().(error)
		}
	}

	return nil
}

type initFunc struct {
	dependencies []reflect.Type
	init         reflect.Value
}

type component struct {
	dependencies []reflect.Type
	constructor  reflect.Value
	value        reflect.Value
}
