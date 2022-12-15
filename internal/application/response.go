package application

import (
	"dgb/meter.oil.readings/internal/configuration"
	"encoding/json"
	"net/http"
)

type Response struct {
	config configuration.Configuration
}

type ResponseParams struct {
	W      http.ResponseWriter
	Result any
}

func (response *Response) Ok(p ResponseParams) {
	response.Write(p.W, 200, p.Result)
}

func (response *Response) Created(p ResponseParams) {
	response.Write(p.W, 201, p.Result)
}

func (response *Response) NotFound(p ResponseParams) {
	response.Write(p.W, 404, p.Result)
}

func (response *Response) BadRequest(p ResponseParams) {
	response.Write(p.W, 403, p.Result)
}

func (response *Response) ServerError(p ResponseParams) {
	response.Write(p.W, 500, p.Result)
}

func (response *Response) Write(w http.ResponseWriter, statusCode int, result any) {
	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", response.config.CORS_CLIENTS)
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Headers", "Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "*")

	w.WriteHeader(statusCode)

	if result != nil {
		json.NewEncoder(w).Encode(result)
	}
}

func NewResponse(configuration configuration.Configuration) *Response {
	return &Response{
		configuration,
	}
}
