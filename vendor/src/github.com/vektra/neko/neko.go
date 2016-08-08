package neko

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

// Simple tuple around a test description and the work
type test struct {
	Name string
	Func func()
}

// Keeps track of mocks, setups, and tests so they can be
// later coordinated.

type Organizer struct {
	t *testing.T

	only    *test
	mocks   []*mock.Mock
	setup   []func()
	cleanup []func()
	tests   []test
}

// Create a new Organizer against testing's T interface
func Start(t *testing.T) *Organizer {
	return &Organizer{t: t}
}

// Track a github.com/stretchr/testify/mock.Mock along with the tests
func (o *Organizer) CheckMock(m *mock.Mock) {
	o.mocks = append(o.mocks, m)
}

// Add some work to be done before each test
func (o *Organizer) Setup(f func()) {
	o.setup = append(o.setup, f)
}

// Add some work to be done after each test
func (o *Organizer) Cleanup(f func()) {
	o.cleanup = append(o.cleanup, f)
}

// Add a test.
func (o *Organizer) It(name string, f func()) {
	o.tests = append(o.tests, test{name, f})
}

func (o *Organizer) Only(name string, f func()) {
	o.only = &test{name, f}
}

// Useful by allowing the developer to simply add 'N' before
// It to disable a block.
func (o *Organizer) NIt(name string, f func()) {
	o.tests = append(o.tests, test{name, nil})
}

// Coordinate running the tests with the setups and mocks
func (o *Organizer) Run() {
	if o.only != nil {
		o.runTest(o.only)
		return
	}

	for _, test := range o.tests {
		o.runTest(&test)
	}
}

func (o *Organizer) runTest(test *test) {
	if test.Func == nil {
		o.t.Logf("==== DISABLED: %s ====", test.Name)
		return
	}

	o.t.Logf("==== %s ====", test.Name)

	for _, mock := range o.mocks {
		mock.ExpectedCalls = nil
		mock.Calls = nil
	}

	for _, setup := range o.setup {
		setup()
	}

	defer o.runCleanup()

	test.Func()

	for _, mock := range o.mocks {
		mock.AssertExpectations(o.t)
	}
}

func (o *Organizer) runCleanup() {
	for _, cleanup := range o.cleanup {
		cleanup()
	}
}

// Have fun with neko!
func (o *Organizer) Meow() {
	o.t.Logf("Meow! Neko is on the case! Running %d tests now!", len(o.tests))
	o.Run()
}
