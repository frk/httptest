package httptest

// StateHandler can be implemented by the client code to provide
// a way to manage the application state of individual tests.
type StateHandler interface {
	// Init can be used to initialize (i.e., set up) a test's state.
	// Init will be invoked BEFORE the state's test is executed.
	Init(State) error
	// Check can be used to check whether a test's state is as expected.
	// Check will be invoked AFTER the state's test is executed.
	Check(State) error
	// Cleanup can be used to clean up (i.e., tear down) a test's state.
	// Cleanup will be invoked AFTER the state's test is executed.
	Cleanup(State) error
}

// A State is a value of any type that the client code can use, together with
// an implementation of the StateHandler, to manage the state of individual tests.
type State any
