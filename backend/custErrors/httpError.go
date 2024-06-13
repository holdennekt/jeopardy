package custErrors

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ErrInternal error = errors.New("internal server error")

type HttpError interface {
	Code() int
	Body() map[string]any
	Error() string
}

type httpError struct {
	code int
	body map[string]any
}

func NewHttpError(code int, body map[string]any) httpError {
	return httpError{code, body}
}

func (he httpError) Code() int {
	return he.code
}

func (he httpError) Body() map[string]any {
	return he.body
}

func (he httpError) Error() string {
	err, _ := json.Marshal(he.body)
	return string(err)
}

func NewInternalError(err error) httpError {
	return NewHttpError(
		http.StatusInternalServerError,
		gin.H{"error": errors.Join(ErrInternal, err).Error()},
	)
}

func AbortWithError(c *gin.Context, httpErr HttpError) {
	log.Println(httpErr.Body())
	c.AbortWithStatusJSON(
		httpErr.Code(),
		httpErr.Body(),
	)
}

func AbortWithInternalError(c *gin.Context, err error) {
	body := gin.H{"error": errors.Join(ErrInternal, err).Error()}
	log.Println(body)
	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		body,
	)
}
