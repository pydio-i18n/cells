// Code generated by go-swagger; DO NOT EDIT.

package provisioning

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/pydio/pydio-sdk-go/models"
)

// NewAdminEditWorkspaceParams creates a new AdminEditWorkspaceParams object
// with the default values initialized.
func NewAdminEditWorkspaceParams() *AdminEditWorkspaceParams {
	var ()
	return &AdminEditWorkspaceParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewAdminEditWorkspaceParamsWithTimeout creates a new AdminEditWorkspaceParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewAdminEditWorkspaceParamsWithTimeout(timeout time.Duration) *AdminEditWorkspaceParams {
	var ()
	return &AdminEditWorkspaceParams{

		timeout: timeout,
	}
}

// NewAdminEditWorkspaceParamsWithContext creates a new AdminEditWorkspaceParams object
// with the default values initialized, and the ability to set a context for a request
func NewAdminEditWorkspaceParamsWithContext(ctx context.Context) *AdminEditWorkspaceParams {
	var ()
	return &AdminEditWorkspaceParams{

		Context: ctx,
	}
}

// NewAdminEditWorkspaceParamsWithHTTPClient creates a new AdminEditWorkspaceParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewAdminEditWorkspaceParamsWithHTTPClient(client *http.Client) *AdminEditWorkspaceParams {
	var ()
	return &AdminEditWorkspaceParams{
		HTTPClient: client,
	}
}

/*AdminEditWorkspaceParams contains all the parameters to send to the API endpoint
for the admin edit workspace operation typically these are written to a http.Request
*/
type AdminEditWorkspaceParams struct {

	/*Payload
	  Repository details

	*/
	Payload *models.AdminWorkspace
	/*WorkspaceID
	  Id or Alias / Update details for this workspace

	*/
	WorkspaceID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the admin edit workspace params
func (o *AdminEditWorkspaceParams) WithTimeout(timeout time.Duration) *AdminEditWorkspaceParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the admin edit workspace params
func (o *AdminEditWorkspaceParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the admin edit workspace params
func (o *AdminEditWorkspaceParams) WithContext(ctx context.Context) *AdminEditWorkspaceParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the admin edit workspace params
func (o *AdminEditWorkspaceParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the admin edit workspace params
func (o *AdminEditWorkspaceParams) WithHTTPClient(client *http.Client) *AdminEditWorkspaceParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the admin edit workspace params
func (o *AdminEditWorkspaceParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPayload adds the payload to the admin edit workspace params
func (o *AdminEditWorkspaceParams) WithPayload(payload *models.AdminWorkspace) *AdminEditWorkspaceParams {
	o.SetPayload(payload)
	return o
}

// SetPayload adds the payload to the admin edit workspace params
func (o *AdminEditWorkspaceParams) SetPayload(payload *models.AdminWorkspace) {
	o.Payload = payload
}

// WithWorkspaceID adds the workspaceID to the admin edit workspace params
func (o *AdminEditWorkspaceParams) WithWorkspaceID(workspaceID string) *AdminEditWorkspaceParams {
	o.SetWorkspaceID(workspaceID)
	return o
}

// SetWorkspaceID adds the workspaceId to the admin edit workspace params
func (o *AdminEditWorkspaceParams) SetWorkspaceID(workspaceID string) {
	o.WorkspaceID = workspaceID
}

// WriteToRequest writes these params to a swagger request
func (o *AdminEditWorkspaceParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Payload != nil {
		if err := r.SetBodyParam(o.Payload); err != nil {
			return err
		}
	}

	// path param workspaceId
	if err := r.SetPathParam("workspaceId", o.WorkspaceID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
