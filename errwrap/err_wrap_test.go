package errwrap

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	fooError struct{ error }
	barError struct{ error }
)

// notModifiedError is a custom error type that contains HTTP headers
type notModifiedError struct {
	headers http.Header
}

func (e notModifiedError) Error() string {
	return "not modified"
}

func (e notModifiedError) Headers() http.Header {
	return e.headers
}

// Constructor for notModifiedError
func newNotModifiedError(headers http.Header) notModifiedError {
	return notModifiedError{headers: headers}
}

// Is performs comparison of two notModifiedError instances.
// Any error should be Comparable, http.Header is not comparable,
// hence, we need to compare headers manually.
func (nm notModifiedError) Is(target error) bool {
	m, ok := target.(notModifiedError)
	return ok && reflect.DeepEqual(nm.headers, m.headers)
}

func TestInnerErrorWrapperIs(t *testing.T) {
	fooInnerErr := &fooError{errors.New("inner error")}
	barInnerErr := &barError{errors.New("inner error")}

	assert.Equal(t, "inner error", fooInnerErr.Error())
	require.NotErrorIs(t, fooInnerErr, errors.New("inner error"))
	require.NotErrorIs(t, fooInnerErr, barInnerErr)
}

func TestInnerErrorWrapperAs(t *testing.T) {
	fooInnerErr := fooError{errors.New("foo error")}
	barInnerErr := barError{errors.New("foo error")}

	var ie fooError

	require.ErrorAs(t, fooInnerErr, &ie)
	require.NotErrorAs(t, barInnerErr, &ie)
	assert.Equal(t, "foo error", ie.Error())
}

func TestNew(t *testing.T) {
	err := Errorf(0, "test error %d", 123)

	assert.Equal(t, "test error 123", err.Error())
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
	assert.Equal(t, "Internal error", err.PublicMessage())
	assert.True(t, err.ShouldReport())
	require.Error(t, err.Unwrap())
	assert.Empty(t, err.messages) // No additional messages
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")

	wrappedErr := Wrap(originalErr)

	assert.Equal(t, "original error", wrappedErr.Error())
	assert.Equal(t, originalErr, wrappedErr.Unwrap())
	assert.Equal(t, http.StatusInternalServerError, wrappedErr.StatusCode())
	assert.Equal(t, "Internal error", wrappedErr.PublicMessage())
	assert.True(t, wrappedErr.ShouldReport())
}

func TestWrapNil(t *testing.T) {
	wrappedErr := Wrap(nil)
	assert.Nil(t, wrappedErr)

	wrappedErr = Wrapf(nil, "some message")
	assert.Nil(t, wrappedErr)
}

func TestWrapAlreadyWrapped(t *testing.T) {
	originalErr := New("original error", 0)

	wrappedErr := Wrap(originalErr)
	assert.NotSame(t, originalErr, wrappedErr)
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("database error")

	wrappedErr := Wrapf(originalErr, "failed to save user %d", 123)

	assert.Equal(t, "database error: failed to save user 123", wrappedErr.Error())
	assert.Equal(t, originalErr, wrappedErr.Unwrap())
	assert.Len(t, wrappedErr.messages, 1)
}

func TestWrapfExistingErrWrap(t *testing.T) {
	originalErr := New("database error", 0)

	// First wrap
	firstWrap := Wrapf(originalErr, "failed to query")

	// Second wrap - should create new instance, not modify original
	secondWrap := Wrapf(firstWrap, "failed to get user")

	// Verify that Clone() was called
	assert.NotSame(t, firstWrap, secondWrap)
	assert.NotSame(t, originalErr, firstWrap)
	assert.NotSame(t, originalErr, secondWrap)

	assert.Equal(t, "database error", originalErr.Error())
	assert.Empty(t, originalErr.messages)

	// Check first wrap
	assert.Equal(t, "database error: failed to query", firstWrap.Error())
	assert.Len(t, firstWrap.messages, 1)
	assert.Equal(t, "failed to query", firstWrap.messages[0])

	// Check second wrap
	assert.Equal(t, "database error: failed to query: failed to get user", secondWrap.Error())
	assert.Len(t, secondWrap.messages, 2)
	assert.Equal(t, "failed to query", secondWrap.messages[0])
	assert.Equal(t, "failed to get user", secondWrap.messages[1])
}

func TestWithStatusCode(t *testing.T) {
	originalErr := New("test error", 0)
	modifiedErr := originalErr.WithStatusCode(http.StatusNotFound)

	assert.Equal(t, http.StatusInternalServerError, originalErr.StatusCode())
	assert.Equal(t, http.StatusNotFound, modifiedErr.StatusCode())
}

func TestWithPublicMessage(t *testing.T) {
	originalErr := New("internal database error", 0)
	modifiedErr := originalErr.WithPublicMessage("Service temporarily unavailable")

	assert.Equal(t, "Internal error", originalErr.PublicMessage())
	assert.Equal(t, "Service temporarily unavailable", modifiedErr.PublicMessage())
}

func TestWithShouldReport(t *testing.T) {
	originalErr := New("test error", 0)
	modifiedErr := originalErr.WithShouldReport(false)

	assert.True(t, originalErr.ShouldReport())
	assert.False(t, modifiedErr.ShouldReport())
	assert.NotSame(t, originalErr, modifiedErr)
}

func TestChaining(t *testing.T) {
	baseErr := errors.New("database connection failed")

	finalErr := Wrapf(baseErr, "failed to save user %d", 123).
		WithStatusCode(http.StatusInternalServerError).
		WithPublicMessage("Unable to save changes").
		WithShouldReport(true)

	assert.Equal(t, "database connection failed: failed to save user 123", finalErr.Error())
	assert.Equal(t, http.StatusInternalServerError, finalErr.StatusCode())
	assert.Equal(t, "Unable to save changes", finalErr.PublicMessage())
	assert.True(t, finalErr.ShouldReport())
	assert.Equal(t, baseErr, finalErr.Unwrap())
}

func TestErrorsIs(t *testing.T) {
	baseErr := errors.New("base error")
	otherErr := errors.New("other error")

	wrappedErr := Wrapf(baseErr, "wrapped error")

	require.ErrorIs(t, wrappedErr, baseErr)
	assert.NotErrorIs(t, wrappedErr, otherErr)
}

func TestErrorsAs(t *testing.T) {
	baseErr := New("base error", 0).WithStatusCode(http.StatusAccepted)
	wrappedErr := Wrapf(baseErr, "wrapped error")

	var extractedErr *ErrWrap
	require.ErrorAs(t, wrappedErr, &extractedErr)
	assert.NotNil(t, extractedErr)

	assert.Equal(t, baseErr.StatusCode(), extractedErr.StatusCode())
	assert.Equal(t, baseErr.PublicMessage(), extractedErr.PublicMessage())
	assert.Equal(t, baseErr.ShouldReport(), extractedErr.ShouldReport())
}

func TestStackTracePreservation(t *testing.T) {
	originalErr := New("original", 0)
	wrappedErr := Wrapf(originalErr, "wrapped")

	assert.NotNil(t, originalErr.stack)
	assert.NotNil(t, wrappedErr.stack)

	// When wrapping an existing ErrWrap, it should preserve the original stack
	assert.Equal(t, originalErr.stack, wrappedErr.stack)

	// When wrapping a regular error, it should capture new stack
	regularErr := errors.New("regular")
	wrappedRegular := Wrapf(regularErr, "wrapped regular")
	assert.NotNil(t, wrappedRegular.stack)
}

func TestNotModifiedError(t *testing.T) {
	headers := make(http.Header)
	headers.Set("Cache-Control", "no-cache")
	headers.Set("ETag", `"abc123"`)
	headers.Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")

	originalErr := newNotModifiedError(headers)
	wrappedErr := Wrapf(originalErr, "cache validation failed for resource %s", "/api/users/123")

	assert.Equal(t, "not modified: cache validation failed for resource /api/users/123", wrappedErr.Error())
	assert.Equal(t, originalErr, wrappedErr.Unwrap())

	require.ErrorIs(t, wrappedErr, originalErr)

	differentHeaders := make(http.Header)
	differentHeaders.Set("Cache-Control", "public")
	differentErr := newNotModifiedError(differentHeaders)
	require.NotErrorIs(t, wrappedErr, differentErr)

	var extractedNotModified notModifiedError
	require.ErrorAs(t, wrappedErr, &extractedNotModified)

	extractedHeaders := extractedNotModified.Headers()
	assert.Equal(t, "no-cache", extractedHeaders.Get("Cache-Control"))
	assert.Equal(t, `"abc123"`, extractedHeaders.Get("ETag"))
	assert.Equal(t, "Wed, 21 Oct 2015 07:28:00 GMT", extractedHeaders.Get("Last-Modified"))
}

func TestWrapStdErr(t *testing.T) {
	err := Wrap(context.DeadlineExceeded)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
