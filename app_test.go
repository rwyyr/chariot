// MIT License
//
// Copyright (c) 2022 Roman Homoliako
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

package chariot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rwyyr/chariot"
)

type (
	A struct {
		mocks struct {
			Run      func(context.Context) error
			Shutdown func(context.Context)
		}
	}

	B struct {
		mocks struct {
			Run      func(context.Context) error
			Shutdown func(context.Context)
		}
	}

	C struct{}

	D struct{}
)

type (
	E interface {
		Foo()
	}

	F struct{}
)

type Error struct{}

func TestNewApp(t *testing.T) {

	t.Run("common-case", func(t *testing.T) {

		testA, testB, testC := new(A), new(B), new(C)

		app, err := chariot.New(
			chariot.With(
				func(b *B) (*A, error) {

					return testA, nil
				},
				func(c *C) (*B, error) {

					return testB, nil
				},
				func() *C {

					return testC
				},
			),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var (
			a *A
			b *B
			c *C
		)
		switch {
		case !app.Retrieve(&a):
			t.FailNow()
		case !app.Retrieve(&b):
			t.FailNow()
		case !app.Retrieve(&c):
			t.FailNow()
		case a != testA:
			t.FailNow()
		case b != testB:
			t.FailNow()
		case c != testC:
			t.FailNow()
		}
	})

	t.Run("multiple-components", func(t *testing.T) {

		testA, testB := new(A), new(B)

		app, err := chariot.New(
			chariot.With(func() (*A, *B, error) {

				return testA, testB, nil
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var (
			a *A
			b *B
		)
		switch {
		case !app.Retrieve(&a):
			t.FailNow()
		case !app.Retrieve(&b):
			t.FailNow()
		case a != testA:
			t.FailNow()
		case b != testB:
			t.FailNow()
		}
	})

	t.Run("missing-dependency", func(t *testing.T) {

		app, err := chariot.New(
			chariot.With(func(*B) *A {

				return new(A)
			}),
		)
		if err != nil {
			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("override-component", func(t *testing.T) {

		app, err := chariot.New(
			chariot.With(
				func() (*A, *B, *C, error) {

					return new(A), new(B), new(C), nil
				},
				func() *B {

					return new(B)
				},
			),
		)
		if err != nil {
			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("constructor-error", func(t *testing.T) {

		testErr := errors.New("test")

		app, err := chariot.New(
			chariot.With(func() (*A, error) {

				return nil, testErr
			}),
		)
		if err != nil {
			if !errors.Is(err, testErr) {
				t.Fatal(err)
			}

			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("initializer", func(t *testing.T) {

		var called bool

		app, err := chariot.New(chariot.With(func() {

			called = true
		}))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if !called {
			t.FailNow()
		}
	})

	t.Run("initializer-error", func(t *testing.T) {

		testErr := errors.New("test")

		app, err := chariot.New(chariot.With(func() error {

			return testErr
		}))
		if err != nil {
			if !errors.Is(err, testErr) {
				t.Fatal(err)
			}

			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("constructor-initializer", func(t *testing.T) {

		testA := new(A)

		var called bool

		app, err := chariot.New(chariot.With(
			func() *A {

				return testA
			},
			func(a *A) {

				called = true

				if a != testA {
					t.FailNow()
				}
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if !called {
			t.FailNow()
		}
	})

	t.Run("constructors-first", func(t *testing.T) {

		testA := new(A)

		var sequence [3]int
		sequenceSlice := sequence[:0]

		app, err := chariot.New(chariot.With(
			func(a *A) *B {

				sequenceSlice = append(sequenceSlice, 1)

				if a != testA {
					t.FailNow()
				}

				return new(B)
			},
			func(a *A) error {

				sequenceSlice = append(sequenceSlice, 2)

				if a != testA {
					t.FailNow()
				}

				return nil
			},
			func() *A {

				sequenceSlice = append(sequenceSlice, 0)

				return testA
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		expectedSequence1 := [...]int{
			0, 1, 2,
		}
		expectedSequence2 := [...]int{
			1, 0, 2,
		}
		if !(sequence == expectedSequence1 || sequence == expectedSequence2) {
			t.Fatal(sequence)
		}
	})

	t.Run("constructor-custom-error", func(t *testing.T) {

		testErr := new(Error)

		app, err := chariot.New(chariot.With(func() (*A, *Error) {

			return new(A), testErr
		}))
		if err != nil {
			if !errors.Is(err, testErr) {
				t.Fatal(err)
			}

			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("initializer-custom-error", func(t *testing.T) {

		testErr := new(Error)

		app, err := chariot.New(chariot.With(func() *Error {

			return testErr
		}))
		if err != nil {
			if !errors.Is(err, testErr) {
				t.Fatal(err)
			}

			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("double-error", func(t *testing.T) {

		testErr := errors.New("test error")

		app, err := chariot.New(chariot.With(func() (error, error) {

			return testErr, nil
		}))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var retrievedErr error
		switch {
		case !app.Retrieve(&retrievedErr):
			t.FailNow()
		case retrievedErr != testErr:
			t.Fatal(retrievedErr)
		}
	})

	t.Run("cycle", func(t *testing.T) {

		app, err := chariot.New(
			chariot.With(
				func(B) *A {

					return new(A)
				},
				func(*C) (B, error) {

					return B{}, nil
				},
				func(*D) (*C, error) {

					return new(C), nil
				},
				func(B) *D {

					return new(D)
				},
			),
		)
		if err != nil {
			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("default-context", func(t *testing.T) {

		testA := new(A)

		app, err := chariot.New(
			chariot.With(func(ctx context.Context) *A {

				if ctx == nil {
					t.FailNow()
				}

				return testA
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var a *A
		switch {
		case !app.Retrieve(&a):
			t.FailNow()
		case a != testA:
			t.FailNow()
		}
	})

	t.Run("with-ctx", func(t *testing.T) {

		key := new(struct{})
		ctx := context.WithValue(context.Background(), key, key)

		testA := new(A)

		app, err := chariot.New(
			chariot.WithInitContext(ctx),
			chariot.With(func(ctx context.Context) *A {

				if value := ctx.Value(key); value != key {
					t.Fatal(value)
				}

				return testA
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var a *A
		switch {
		case !app.Retrieve(&a):
			t.FailNow()
		case a != testA:
			t.FailNow()
		}
	})

	t.Run("component", func(t *testing.T) {

		testA := new(A)

		app, err := chariot.New(chariot.WithComponents(testA))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var a *A
		switch {
		case !app.Retrieve(&a):
			t.FailNow()
		case a != testA:
			t.FailNow()
		}
	})

	t.Run("component-override", func(t *testing.T) {

		app, err := chariot.New(
			chariot.With(func() *A {

				return new(A)
			}),
			chariot.WithComponents(new(A)),
		)
		if err != nil {
			return
		}
		defer app.Shutdown()

		t.FailNow()
	})

	t.Run("interface-component", func(t *testing.T) {

		var testE E = new(F)

		app, err := chariot.New(chariot.WithComponents(testE))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var f *F
		switch {
		case !app.Retrieve(&f):
			t.FailNow()
		case f != testE:
			t.FailNow()
		}
	})

	t.Run("variadic", func(t *testing.T) {

		testC := new(C)

		app, err := chariot.New(
			chariot.WithComponents(new(A)),
			chariot.With(func(*A, ...*B) *C {

				return testC
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var c *C
		switch {
		case !app.Retrieve(&c):
			t.FailNow()
		case c != testC:
			t.FailNow()
		}
	})
}

func TestAppRun(t *testing.T) {

	t.Run("simple-case", func(t *testing.T) {

		var aCalled, bCalled bool

		app, err := chariot.New(
			chariot.With(
				func() A {

					var a A
					a.mocks.Run = func(context.Context) error {

						aCalled = true

						return nil
					}

					return a
				},
				func() *B {

					var b B
					b.mocks.Run = func(context.Context) error {

						bCalled = true

						return nil
					}

					return &b
				},
			),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if err := app.Run(); err != nil {
			t.Fatal(err)
		}

		switch {
		case !aCalled:
			t.FailNow()
		case !bCalled:
			t.FailNow()
		}
	})

	t.Run("no-runners", func(t *testing.T) {

		app, err := chariot.New(chariot.With(
			func() (*C, error) {

				return new(C), nil
			},
			func(*C) D {

				return D{}
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if err := app.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error-returned", func(t *testing.T) {

		testErr := errors.New("test error")

		app, err := chariot.New(chariot.With(
			func() (*A, error) {

				var a A
				a.mocks.Run = func(context.Context) error {

					return testErr
				}

				return &a, nil
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if err := app.Run(); !errors.Is(err, testErr) {
			t.Fatal(err)
		}
	})

	t.Run("two-errors", func(t *testing.T) {

		testErr1, testErr2 := errors.New("test error 1"), errors.New("test error 2")

		app, err := chariot.New(chariot.With(
			func() (A, *B) {

				var a A
				a.mocks.Run = func(context.Context) error {

					return testErr1
				}

				var b B
				b.mocks.Run = func(context.Context) error {

					return testErr2
				}

				return a, &b
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		var secondErr error

		if err := app.Run(chariot.WithErrHandler(func(_ context.Context, err error) {

			secondErr = err
		})); err != nil {
			if !(errors.Is(err, testErr1) || errors.Is(err, testErr2)) {
				t.Fatal(err)
			}

			if !(errors.Is(secondErr, testErr1) || errors.Is(secondErr, testErr2)) {
				t.Fatal(err)
			}

			if err == secondErr {
				t.FailNow()
			}

			return
		}

		t.FailNow()
	})

	t.Run("error-cancel", func(t *testing.T) {

		testErr := errors.New("test error")

		var called bool

		app, err := chariot.New(chariot.With(
			func() A {

				var a A
				a.mocks.Run = func(context.Context) error {

					return testErr
				}

				return a
			},
			func() B {

				var b B
				b.mocks.Run = func(ctx context.Context) error {

					called = true

					<-ctx.Done()

					return nil
				}

				return b
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if err := app.Run(); err != nil {
			if !errors.Is(err, testErr) {
				t.Fatal(err)
			}

			if !called {
				t.FailNow()
			}

			return
		}

		t.FailNow()
	})

	t.Run("with-context", func(t *testing.T) {

		var aCalled bool

		keyValue := new(struct{})

		app, err := chariot.New(
			chariot.With(func() A {

				var a A
				a.mocks.Run = func(ctx context.Context) error {

					aCalled = true

					if value := ctx.Value(keyValue); value != keyValue {
						t.Fatal(value)
					}

					return nil
				}

				return a
			}),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()

		if err := app.Run(
			chariot.WithRunContext(
				context.WithValue(context.Background(), keyValue, keyValue),
			),
		); err != nil {
			t.Fatal(err)
		}

		if !aCalled {
			t.FailNow()
		}
	})
}

func TestAppShutdown(t *testing.T) {

	t.Run("simple-case", func(t *testing.T) {

		var orderData [2]int
		order := orderData[:0]

		defer func() {

			if orderData != [...]int{
				1, 2,
			} {
				t.Fatal(orderData)
			}
		}()

		app, err := chariot.New(chariot.With(
			func() A {

				var a A
				a.mocks.Shutdown = func(context.Context) {

					order = append(order, 2)
				}

				return a
			},
			func(A) B {

				var b B
				b.mocks.Shutdown = func(context.Context) {

					order = append(order, 1)
				}

				return b
			},
		))
		if err != nil {
			t.Fatal(err)
		}
		defer app.Shutdown()
	})

	t.Run("upon-new", func(t *testing.T) {

		var orderData [2]int
		order := orderData[:0]

		testErr := errors.New("test error")

		_, err := chariot.New(chariot.With(
			func() A {

				var a A
				a.mocks.Shutdown = func(context.Context) {

					order = append(order, 2)
				}

				return a
			},
			func(A) *B {

				var b B
				b.mocks.Shutdown = func(context.Context) {

					order = append(order, 1)
				}

				return &b
			},
			func(A, *B) (C, error) {

				return C{}, testErr
			},
		))
		if err == nil {
			t.FailNow()
		}

		if !errors.Is(err, testErr) {
			t.Fatal(err)
		}

		if orderData != [...]int{
			1, 2,
		} {
			t.Fatal(orderData)
		}
	})
}

func (a A) Run(ctx context.Context) (_ error) {

	if a.mocks.Run != nil {
		return a.mocks.Run(ctx)
	}

	return
}

func (a A) Shutdown(ctx context.Context) {

	if a.mocks.Shutdown != nil {
		a.mocks.Shutdown(ctx)
	}
}

func (b B) Run(ctx context.Context) (_ error) {

	if b.mocks.Run != nil {
		return b.mocks.Run(ctx)
	}

	return
}

func (b B) Shutdown(ctx context.Context) {

	if b.mocks.Shutdown != nil {
		b.mocks.Shutdown(ctx)
	}
}

func (*F) Foo() {}

func (Error) Error() (_ string) {

	return
}
