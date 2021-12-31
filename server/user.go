package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/msksgm/go-realworld-msksgm-copy/conduit"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fid reflect.StructField) string {
		name := strings.SplitN(fid.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			name = ""
		}
		return name
	})
}

func (s *Server) createUser() http.HandlerFunc {
	type Input struct {
		User struct {
			Email    string `json:"email" validate:"required,email"`
			Username string `json:"username" validate:"required,min=2"`
			Password string `json:"password" validate:"required,min=8,max=72"`
		} `json:"user" validate:"required"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		input := &Input{}

		if err := readJSON(r.Body, &input); err != nil {
			errorResponse(w, http.StatusUnprocessableEntity, err)
			return
		}

		if err := validate.Struct(input.User); err != nil {
			validationError(w, err)
			return
		}

		user := conduit.User{
			Email:    input.User.Email,
			Username: input.User.Username,
		}

		user.SetPassword(input.User.Password)

		if err := s.userService.CreateUser(r.Context(), &user); err != nil {
			switch {
			case errors.Is(err, conduit.ErrDuplicateEmail):
				err = ErrorM{"email": []string{"this email is already in use"}}
				errorResponse(w, http.StatusConflict, err)
			case errors.Is(err, conduit.ErrDuplicateUsername):
				err = ErrorM{"username": []string{"this username is already in use"}}
				errorResponse(w, http.StatusConflict, err)
			default:
				serverError(w, err)
			}
			return
		}

		writeJSON(w, http.StatusCreated, M{"user": user})
	}
}

func (s *Server) loginUser() http.HandlerFunc {
	type Input struct {
		User struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"user"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		input := Input{}

		if err := readJSON(r.Body, &input); err != nil {
			errorResponse(w, http.StatusUnprocessableEntity, err)
			return
		}

		user, err := s.userService.Authenticate(r.Context(), input.User.Email, input.User.Password)

		if err != nil || user == nil {
			invalidUserCredentialsError(w)
			return
		}

		token, err := generateUserToken(user)
		if err != nil {
			serverError(w, err)
			return
		}

		user.Token = token

		writeJSON(w, http.StatusOK, M{"user": user})
	}
}

func (s *Server) getCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := userFromContext(ctx)
		user.Token = userTokenFromContext(ctx)

		fmt.Println(ctx)
		writeJSON(w, http.StatusOK, M{"user": user})
	}
}
