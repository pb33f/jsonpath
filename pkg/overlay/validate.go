package overlay

import (
	"github.com/pb33f/jsonpath/pkg/jsonpath"
	"github.com/pb33f/jsonpath/pkg/jsonpath/config"
	"net/url"
	"strings"
)

type ValidationErrors []error

func (v ValidationErrors) Error() string {
	msgs := make([]string, len(v))
	for i, err := range v {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}

func (v ValidationErrors) Return() error {
	if len(v) > 0 {
		return v
	}
	return nil
}

func (o *Overlay) Validate() error {
	errs := make(ValidationErrors, 0)
	if o.Version != "1.0.0" {
		errs = append(errs, &OverlayVersionError{})
	}

	if o.Info.Title == "" {
		errs = append(errs, &OverlayTitleError{})
	}
	if o.Info.Version == "" {
		errs = append(errs, &OverlayVersionFieldError{})
	}

	if o.Extends != "" {
		_, err := url.Parse(o.Extends)
		if err != nil {
			errs = append(errs, &OverlayExtendsURLError{Cause: err})
		}
	}

	for i, action := range o.Actions {
		if action.Target == "" {
			errs = append(errs, &ActionTargetMissingError{Index: i})
		}

		if action.Remove && !action.Update.IsZero() {
			errs = append(errs, &ActionRemoveUpdateConflictError{Index: i})
		}

		if action.Upsert {
			if action.Remove {
				errs = append(errs, &ActionUpsertRemoveConflictError{Index: i})
			} else if !action.Update.IsZero() {
				p, err := jsonpath.NewPath(action.Target, config.WithPropertyNameExtension())
				if err != nil {
					errs = append(errs, &ActionInvalidTargetPathError{Index: i, Cause: err})
				} else if !p.IsSingular() {
					errs = append(errs, &ActionUpsertNonSingularPathError{Index: i, Target: action.Target})
				}
			}
		}
	}

	return errs.Return()
}
