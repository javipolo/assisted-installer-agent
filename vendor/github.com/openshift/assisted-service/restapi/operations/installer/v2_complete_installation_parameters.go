// Code generated by go-swagger; DO NOT EDIT.

package installer

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"

	"github.com/openshift/assisted-service/models"
)

// NewV2CompleteInstallationParams creates a new V2CompleteInstallationParams object
//
// There are no default values defined in the spec.
func NewV2CompleteInstallationParams() V2CompleteInstallationParams {

	return V2CompleteInstallationParams{}
}

// V2CompleteInstallationParams contains all the bound params for the v2 complete installation operation
// typically these are obtained from a http.Request
//
// swagger:parameters v2CompleteInstallation
type V2CompleteInstallationParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*The cluster whose installation is being completing.
	  Required: true
	  In: path
	*/
	ClusterID strfmt.UUID
	/*The final status of the cluster installation.
	  Required: true
	  In: body
	*/
	CompletionParams *models.CompletionParams
	/*The software version of the discovery agent that is completing the installation.
	  In: header
	*/
	DiscoveryAgentVersion *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewV2CompleteInstallationParams() beforehand.
func (o *V2CompleteInstallationParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rClusterID, rhkClusterID, _ := route.Params.GetOK("cluster_id")
	if err := o.bindClusterID(rClusterID, rhkClusterID, route.Formats); err != nil {
		res = append(res, err)
	}

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.CompletionParams
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("completionParams", "body", ""))
			} else {
				res = append(res, errors.NewParseError("completionParams", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			ctx := validate.WithOperationRequest(context.Background())
			if err := body.ContextValidate(ctx, route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.CompletionParams = &body
			}
		}
	} else {
		res = append(res, errors.Required("completionParams", "body", ""))
	}

	if err := o.bindDiscoveryAgentVersion(r.Header[http.CanonicalHeaderKey("discovery_agent_version")], true, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindClusterID binds and validates parameter ClusterID from path.
func (o *V2CompleteInstallationParams) bindClusterID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	// Format: uuid
	value, err := formats.Parse("uuid", raw)
	if err != nil {
		return errors.InvalidType("cluster_id", "path", "strfmt.UUID", raw)
	}
	o.ClusterID = *(value.(*strfmt.UUID))

	if err := o.validateClusterID(formats); err != nil {
		return err
	}

	return nil
}

// validateClusterID carries on validations for parameter ClusterID
func (o *V2CompleteInstallationParams) validateClusterID(formats strfmt.Registry) error {

	if err := validate.FormatOf("cluster_id", "path", "uuid", o.ClusterID.String(), formats); err != nil {
		return err
	}
	return nil
}

// bindDiscoveryAgentVersion binds and validates parameter DiscoveryAgentVersion from header.
func (o *V2CompleteInstallationParams) bindDiscoveryAgentVersion(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.DiscoveryAgentVersion = &raw

	return nil
}
