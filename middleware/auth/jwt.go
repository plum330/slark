package auth

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/middleware"
	"github.com/go-slark/slark/pkg"
	"github.com/golang-jwt/jwt"
)

type Option func(*option)

type option struct {
	signingMethod jwt.SigningMethod
	claims        func() jwt.Claims
}

func SigningMethod(method jwt.SigningMethod) Option {
	return func(o *option) {
		o.signingMethod = method
	}
}

func Claims(f func() jwt.Claims) Option {
	return func(o *option) {
		o.claims = f
	}
}

const msg = "unauthorized"

var (
	jwtKeyFuncMiss        = errors.Unauthorized(msg, "jwt key func miss")
	jwtTokenMiss          = errors.Unauthorized(msg, "jwt token miss")
	jwtTokenInvalid       = errors.Unauthorized(msg, "jwt token invalid")
	jwtTokenExpired       = errors.Unauthorized(msg, "jwt token expired")
	jwtTokenParseError    = errors.Unauthorized(msg, "jwt token parse error")
	jwtSigningMethodError = errors.Unauthorized(msg, "jwt signing method error")
)

func Authorize(f jwt.Keyfunc, opts ...Option) middleware.Middleware {
	o := &option{
		signingMethod: jwt.SigningMethodHS256,
	}
	for _, opt := range opts {
		opt(o)
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if f == nil {
				return nil, jwtKeyFuncMiss
			}
			authorization, ok := ctx.Value(utils.Authorization).(string)
			if !ok {
				return nil, jwtTokenMiss
			}
			var (
				token *jwt.Token
				err   error
			)
			if o.claims != nil {
				token, err = jwt.ParseWithClaims(authorization, o.claims(), f)
			} else {
				token, err = jwt.Parse(authorization, f)
			}
			if err != nil {
				ve, ok := err.(*jwt.ValidationError)
				if !ok {
					return nil, errors.Unauthorized(msg, err.Error())
				}
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					return nil, jwtTokenInvalid
				}
				if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
					return nil, jwtTokenExpired
				}
				return nil, jwtTokenParseError
			}
			if !token.Valid {
				return nil, jwtTokenInvalid
			}
			if token.Method != o.signingMethod {
				return nil, jwtSigningMethodError
			}
			// TODO
			ctx = context.WithValue(ctx, utils.Claims, token.Claims)
			return handler(ctx, req)
		}
	}
}
