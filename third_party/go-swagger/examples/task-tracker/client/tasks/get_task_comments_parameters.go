// Code generated by go-swagger; DO NOT EDIT.

package tasks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewGetTaskCommentsParams creates a new GetTaskCommentsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetTaskCommentsParams() *GetTaskCommentsParams {
	return &GetTaskCommentsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetTaskCommentsParamsWithTimeout creates a new GetTaskCommentsParams object
// with the ability to set a timeout on a request.
func NewGetTaskCommentsParamsWithTimeout(timeout time.Duration) *GetTaskCommentsParams {
	return &GetTaskCommentsParams{
		timeout: timeout,
	}
}

// NewGetTaskCommentsParamsWithContext creates a new GetTaskCommentsParams object
// with the ability to set a context for a request.
func NewGetTaskCommentsParamsWithContext(ctx context.Context) *GetTaskCommentsParams {
	return &GetTaskCommentsParams{
		Context: ctx,
	}
}

// NewGetTaskCommentsParamsWithHTTPClient creates a new GetTaskCommentsParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetTaskCommentsParamsWithHTTPClient(client *http.Client) *GetTaskCommentsParams {
	return &GetTaskCommentsParams{
		HTTPClient: client,
	}
}

/* GetTaskCommentsParams contains all the parameters to send to the API endpoint
   for the get task comments operation.

   Typically these are written to a http.Request.
*/
type GetTaskCommentsParams struct {

	/* ID.

	   The id of the item

	   Format: int64
	*/
	ID int64

	/* PageSize.

	   Amount of items to return in a single page

	   Format: int32
	   Default: 20
	*/
	PageSize *int32

	/* Since.

	   The created time of the oldest seen comment

	   Format: date-time
	*/
	Since *strfmt.DateTime

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get task comments params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskCommentsParams) WithDefaults() *GetTaskCommentsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get task comments params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskCommentsParams) SetDefaults() {
	var (
		pageSizeDefault = int32(20)
	)

	val := GetTaskCommentsParams{
		PageSize: &pageSizeDefault,
	}

	val.timeout = o.timeout
	val.Context = o.Context
	val.HTTPClient = o.HTTPClient
	*o = val
}

// WithTimeout adds the timeout to the get task comments params
func (o *GetTaskCommentsParams) WithTimeout(timeout time.Duration) *GetTaskCommentsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get task comments params
func (o *GetTaskCommentsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get task comments params
func (o *GetTaskCommentsParams) WithContext(ctx context.Context) *GetTaskCommentsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get task comments params
func (o *GetTaskCommentsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get task comments params
func (o *GetTaskCommentsParams) WithHTTPClient(client *http.Client) *GetTaskCommentsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get task comments params
func (o *GetTaskCommentsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the get task comments params
func (o *GetTaskCommentsParams) WithID(id int64) *GetTaskCommentsParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get task comments params
func (o *GetTaskCommentsParams) SetID(id int64) {
	o.ID = id
}

// WithPageSize adds the pageSize to the get task comments params
func (o *GetTaskCommentsParams) WithPageSize(pageSize *int32) *GetTaskCommentsParams {
	o.SetPageSize(pageSize)
	return o
}

// SetPageSize adds the pageSize to the get task comments params
func (o *GetTaskCommentsParams) SetPageSize(pageSize *int32) {
	o.PageSize = pageSize
}

// WithSince adds the since to the get task comments params
func (o *GetTaskCommentsParams) WithSince(since *strfmt.DateTime) *GetTaskCommentsParams {
	o.SetSince(since)
	return o
}

// SetSince adds the since to the get task comments params
func (o *GetTaskCommentsParams) SetSince(since *strfmt.DateTime) {
	o.Since = since
}

// WriteToRequest writes these params to a swagger request
func (o *GetTaskCommentsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", swag.FormatInt64(o.ID)); err != nil {
		return err
	}

	if o.PageSize != nil {

		// query param pageSize
		var qrPageSize int32

		if o.PageSize != nil {
			qrPageSize = *o.PageSize
		}
		qPageSize := swag.FormatInt32(qrPageSize)
		if qPageSize != "" {

			if err := r.SetQueryParam("pageSize", qPageSize); err != nil {
				return err
			}
		}
	}

	if o.Since != nil {

		// query param since
		var qrSince strfmt.DateTime

		if o.Since != nil {
			qrSince = *o.Since
		}
		qSince := qrSince.String()
		if qSince != "" {

			if err := r.SetQueryParam("since", qSince); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
