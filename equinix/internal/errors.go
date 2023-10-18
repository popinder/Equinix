package internal

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/packethost/packngo"
)

// FriendlyError improves error messages when the API error is blank or in an
// alternate format (as is the case with invalid token or loadbalancer errors)
func FriendlyError(err error) error {
	if e, ok := err.(*packngo.ErrorResponse); ok {
		resp := e.Response
		errors := Errors(e.Errors)

		if 0 == len(errors) {
			errors = Errors{e.SingleError}
		}

		return convertToFriendlyError(errors, resp)
	}
	return err
}

func FriendlyErrorForMetalGo(err error, resp *http.Response) error {
	errors := Errors([]string{err.Error()})
	return convertToFriendlyError(errors, resp)
}

func convertToFriendlyError(errors Errors, resp *http.Response) error {
	er := &ErrorResponse{
		StatusCode: resp.StatusCode,
		Errors:     errors,
	}
	respHead := resp.Header

	// this checks if the error comes from API (and not from cache/LB)
	if len(errors) > 0 {
		ct := respHead.Get("Content-Type")
		xrid := respHead.Get("X-Request-Id")
		if strings.Contains(ct, "application/json") && len(xrid) > 0 {
			er.IsAPIError = true
		}
	}
	return er
}

func IsForbidden(err error) bool {
	r, ok := err.(*packngo.ErrorResponse)
	if ok && r.Response != nil {
		return r.Response.StatusCode == http.StatusForbidden
	}
	if r, ok := err.(*ErrorResponse); ok {
		return r.StatusCode == http.StatusForbidden
	}
	return false
}

func IsNotFound(err error) bool {
	if r, ok := err.(*ErrorResponse); ok {
		return r.StatusCode == http.StatusNotFound && r.IsAPIError
	}
	if r, ok := err.(*packngo.ErrorResponse); ok && r.Response != nil {
		return r.Response.StatusCode == http.StatusNotFound
	}
	return false
}

type Errors []string

func (e Errors) Error() string {
	return strings.Join(e, "; ")
}

type ErrorResponse struct {
	StatusCode int
	Errors
	IsAPIError bool
}

func (er *ErrorResponse) Error() string {
	ret := ""
	if er.IsAPIError {
		ret += "API Error "
	}
	if er.StatusCode != 0 {
		ret += fmt.Sprintf("HTTP %d ", er.StatusCode)
	}
	ret += er.Errors.Error()
	return ret
}

// setMap sets the map of values to ResourceData, checking and returning the
// errors. Typically d.Set is not error checked. This helper makes checking
// those errors less tedious. Because this works with a map, the order of the
// errors would not be predictable, to avoid this the errors will be sorted.
func setMap(d *schema.ResourceData, m map[string]interface{}) error {
	errs := &multierror.Error{}
	for key, v := range m {
		var err error
		if f, ok := v.(setFn); ok {
			err = f(d, key)
		} else {
			if key == "router" {
				d.Set("gateway", v)
			}
			err = d.Set(key, v)
		}

		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	sort.Sort(errs)

	return errs.ErrorOrNil()
}

type setFn = func(d *schema.ResourceData, key string) error

// IsNotAssigned matches errors reported from unassigned virtual networks
func IsNotAssigned(resp *http.Response, err error) bool {
	if resp.StatusCode != http.StatusUnprocessableEntity {
		return false
	}
	if err, ok := err.(*packngo.ErrorResponse); ok {
		for _, e := range append(err.Errors, err.SingleError) {
			if strings.HasPrefix(e, "Virtual network") && strings.HasSuffix(e, "not assigned") {
				return true
			}
		}
	}
	return false
}

func HttpForbidden(resp *http.Response, err error) bool {
	if resp != nil && (resp.StatusCode != http.StatusForbidden) {
		return false
	}

	switch err := err.(type) {
	case *ErrorResponse, *packngo.ErrorResponse:
		return IsForbidden(err)
	}

	return false
}

func HttpNotFound(resp *http.Response, err error) bool {
	if resp != nil && (resp.StatusCode != http.StatusNotFound) {
		return false
	}

	switch err := err.(type) {
	case *ErrorResponse, *packngo.ErrorResponse:
		return IsNotFound(err)
	}
	return false
}

// IgnoreResponseErrors ignores http response errors when matched by one of the
// provided checks
func IgnoreResponseErrors(ignore ...func(resp *http.Response, err error) bool) func(resp *packngo.Response, err error) error {
	return func(resp *packngo.Response, err error) error {
		var r *http.Response
		if resp != nil && resp.Response != nil {
			r = resp.Response
		}
		mute := false
		for _, ignored := range ignore {
			if ignored(r, err) {
				mute = true
				break
			}
		}

		if mute {
			return nil
		}
		return err
	}
}
