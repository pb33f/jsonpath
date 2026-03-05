package overlay

import "fmt"

type UpsertError struct {
	Path  string
	Msg   string
	Index int64
	Len   int
}

func (e *UpsertError) Error() string {
	if e.Index >= 0 && e.Len >= 0 {
		return fmt.Sprintf("upsert error at %q: %s (index %d, length %d)", e.Path, e.Msg, e.Index, e.Len)
	}
	if e.Path != "" {
		return fmt.Sprintf("upsert error at %q: %s", e.Path, e.Msg)
	}
	return fmt.Sprintf("upsert error: %s", e.Msg)
}

type ArrayIndexOutOfBoundsError struct {
	Index     int64
	Length    int
	CanAppend bool
}

func (e *ArrayIndexOutOfBoundsError) Error() string {
	if e.CanAppend {
		return fmt.Sprintf("array index %d out of bounds (length %d); only appending at index %d is allowed",
			e.Index, e.Length, e.Length)
	}
	return fmt.Sprintf("array index %d out of bounds (length %d)", e.Index, e.Length)
}

type NegativeArrayIndexError struct {
	Index int64
}

func (e *NegativeArrayIndexError) Error() string {
	return fmt.Sprintf("negative array index not supported: %d", e.Index)
}

type TypeMismatchError struct {
	Expected string
	Actual   string
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Actual)
}

type CannotCreateArrayAtIndexError struct {
	Index int64
}

func (e *CannotCreateArrayAtIndexError) Error() string {
	return fmt.Sprintf("cannot create new array at non-zero index %d", e.Index)
}

type UnknownSegmentKindError struct {
	Kind interface{}
}

func (e *UnknownSegmentKindError) Error() string {
	return fmt.Sprintf("unknown segment kind: %v", e.Kind)
}

type EmptyPathError struct{}

func (e *EmptyPathError) Error() string {
	return "empty path for upsert"
}

type OverlayVersionError struct{}

func (e *OverlayVersionError) Error() string {
	return "overlay version must be 1.0.0"
}

type OverlayTitleError struct{}

func (e *OverlayTitleError) Error() string {
	return "overlay info title must be defined"
}

type OverlayVersionFieldError struct{}

func (e *OverlayVersionFieldError) Error() string {
	return "overlay info version must be defined"
}

type OverlayExtendsURLError struct {
	Cause error
}

func (e *OverlayExtendsURLError) Error() string {
	return fmt.Sprintf("overlay extends must be a valid URL")
}

func (e *OverlayExtendsURLError) Unwrap() error {
	return e.Cause
}

type ActionTargetMissingError struct {
	Index int
}

func (e *ActionTargetMissingError) Error() string {
	return fmt.Sprintf("overlay action at index %d target must be defined", e.Index)
}

type ActionRemoveUpdateConflictError struct {
	Index int
}

func (e *ActionRemoveUpdateConflictError) Error() string {
	return fmt.Sprintf("overlay action at index %d should not both set remove and define update", e.Index)
}

type ActionUpsertRemoveConflictError struct {
	Index int
}

func (e *ActionUpsertRemoveConflictError) Error() string {
	return fmt.Sprintf("overlay action at index %d should not both set upsert and remove", e.Index)
}

type ActionInvalidTargetPathError struct {
	Index int
	Cause error
}

func (e *ActionInvalidTargetPathError) Error() string {
	return fmt.Sprintf("overlay action at index %d has invalid target path", e.Index)
}

func (e *ActionInvalidTargetPathError) Unwrap() error {
	return e.Cause
}

type ActionUpsertNonSingularPathError struct {
	Index  int
	Target string
}

func (e *ActionUpsertNonSingularPathError) Error() string {
	return fmt.Sprintf("overlay action at index %d upsert requires a singular path (no wildcards, recursive descent, slices, or filters): %s", e.Index, e.Target)
}
