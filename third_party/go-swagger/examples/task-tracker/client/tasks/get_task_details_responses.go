// Code generated by go-swagger; DO NOT EDIT.

package tasks

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/go-swagger/go-swagger/examples/task-tracker/models"
)

// GetTaskDetailsReader is a Reader for the GetTaskDetails structure.
type GetTaskDetailsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetTaskDetailsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetTaskDetailsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 422:
		result := NewGetTaskDetailsUnprocessableEntity()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetTaskDetailsDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetTaskDetailsOK creates a GetTaskDetailsOK with default headers values
func NewGetTaskDetailsOK() *GetTaskDetailsOK {
	return &GetTaskDetailsOK{}
}

/* GetTaskDetailsOK describes a response with status code 200, with default header values.

Task details
*/
type GetTaskDetailsOK struct {
	Payload *models.Task
}

// IsSuccess returns true when this get task details o k response returns a 2xx status code
func (o *GetTaskDetailsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get task details o k response returns a 3xx status code
func (o *GetTaskDetailsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get task details o k response returns a 4xx status code
func (o *GetTaskDetailsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get task details o k response returns a 5xx status code
func (o *GetTaskDetailsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get task details o k response returns a 4xx status code
func (o *GetTaskDetailsOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetTaskDetailsOK) Error() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetailsOK  %+v", 200, o.Payload)
}

func (o *GetTaskDetailsOK) String() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetailsOK  %+v", 200, o.Payload)
}

func (o *GetTaskDetailsOK) GetPayload() *models.Task {
	return o.Payload
}

func (o *GetTaskDetailsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Task)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTaskDetailsUnprocessableEntity creates a GetTaskDetailsUnprocessableEntity with default headers values
func NewGetTaskDetailsUnprocessableEntity() *GetTaskDetailsUnprocessableEntity {
	return &GetTaskDetailsUnprocessableEntity{}
}

/* GetTaskDetailsUnprocessableEntity describes a response with status code 422, with default header values.

Validation error
*/
type GetTaskDetailsUnprocessableEntity struct {
	Payload *models.ValidationError
}

// IsSuccess returns true when this get task details unprocessable entity response returns a 2xx status code
func (o *GetTaskDetailsUnprocessableEntity) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get task details unprocessable entity response returns a 3xx status code
func (o *GetTaskDetailsUnprocessableEntity) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get task details unprocessable entity response returns a 4xx status code
func (o *GetTaskDetailsUnprocessableEntity) IsClientError() bool {
	return true
}

// IsServerError returns true when this get task details unprocessable entity response returns a 5xx status code
func (o *GetTaskDetailsUnprocessableEntity) IsServerError() bool {
	return false
}

// IsCode returns true when this get task details unprocessable entity response returns a 4xx status code
func (o *GetTaskDetailsUnprocessableEntity) IsCode(code int) bool {
	return code == 422
}

func (o *GetTaskDetailsUnprocessableEntity) Error() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetailsUnprocessableEntity  %+v", 422, o.Payload)
}

func (o *GetTaskDetailsUnprocessableEntity) String() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetailsUnprocessableEntity  %+v", 422, o.Payload)
}

func (o *GetTaskDetailsUnprocessableEntity) GetPayload() *models.ValidationError {
	return o.Payload
}

func (o *GetTaskDetailsUnprocessableEntity) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ValidationError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTaskDetailsDefault creates a GetTaskDetailsDefault with default headers values
func NewGetTaskDetailsDefault(code int) *GetTaskDetailsDefault {
	return &GetTaskDetailsDefault{
		_statusCode: code,
	}
}

/* GetTaskDetailsDefault describes a response with status code -1, with default header values.

Error response
*/
type GetTaskDetailsDefault struct {
	_statusCode int
	XErrorCode  string

	Payload *models.Error
}

// Code gets the status code for the get task details default response
func (o *GetTaskDetailsDefault) Code() int {
	return o._statusCode
}

// IsSuccess returns true when this get task details default response returns a 2xx status code
func (o *GetTaskDetailsDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this get task details default response returns a 3xx status code
func (o *GetTaskDetailsDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this get task details default response returns a 4xx status code
func (o *GetTaskDetailsDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this get task details default response returns a 5xx status code
func (o *GetTaskDetailsDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this get task details default response returns a 4xx status code
func (o *GetTaskDetailsDefault) IsCode(code int) bool {
	return o._statusCode == code
}

func (o *GetTaskDetailsDefault) Error() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetails default  %+v", o._statusCode, o.Payload)
}

func (o *GetTaskDetailsDefault) String() string {
	return fmt.Sprintf("[GET /tasks/{id}][%d] getTaskDetails default  %+v", o._statusCode, o.Payload)
}

func (o *GetTaskDetailsDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetTaskDetailsDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// hydrates response header X-Error-Code
	hdrXErrorCode := response.GetHeader("X-Error-Code")

	if hdrXErrorCode != "" {
		o.XErrorCode = hdrXErrorCode
	}

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
