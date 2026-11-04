package auth

import (
	"net/http"

	"github.com/rpmoulton/mint-alternative/internal/httpx"
	"github.com/rpmoulton/mint-alternative/internal/jwt"
)

func (h *Handler) Routes() *httpx.Group {
	g := httpx.NewGroup(httpx.Defaults{Tags: []string{"auth"}})

	httpx.Register(g, httpx.Operation[RegisterInput, AuthResponse]{
		Method:             http.MethodPost,
		Path:               "/register",
		Summary:            "Register a user",
		BodyDescription:    "New user registration details",
		SuccessDescription: "User registered and tokens issued",
		Success:            httpx.Enveloped(http.StatusCreated),
		Problems: []httpx.ProblemKind{
			httpx.Conflict().Described("A user with this email already exists"),
		},
		Handler: h.Register,
	})

	httpx.Register(g, httpx.Operation[LoginInput, AuthResponse]{
		Method:             http.MethodPost,
		Path:               "/login",
		Summary:            "Log in",
		BodyDescription:    "User credentials",
		SuccessDescription: "Authentication successful and tokens issued",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.Unauthorized().Described("The credentials are invalid"),
		},
		Handler: h.Login,
	})

	httpx.Register(g, httpx.Operation[EmptyInput, UserResponse]{
		Method:             http.MethodGet,
		Path:               "/me",
		Summary:            "Get current user profile",
		SuccessDescription: "Current user profile returned",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.NotFound().Described("The authenticated user no longer exists"),
			httpx.BadRequest().Described("Invalid user token subject"),
		},
		Guards:  []httpx.Guard{jwt.AccessGuard()},
		Handler: h.Me,
	})

	httpx.Register(g, httpx.Operation[EmptyInput, jwt.Tokens]{
		Method:             http.MethodPost,
		Path:               "/refresh",
		Summary:            "Refresh authentication tokens",
		SuccessDescription: "Fresh authentication tokens issued",
		Success:            httpx.Enveloped(http.StatusOK),
		Problems: []httpx.ProblemKind{
			httpx.Unauthorized().Described("The refresh identity is invalid"),
		},
		Guards:  []httpx.Guard{jwt.RefreshGuard()},
		Handler: h.Refresh,
	})

	httpx.Register(g, httpx.Operation[EmptyInput, httpx.NoBody]{
		Method:             http.MethodPost,
		Path:               "/logout",
		Summary:            "Log out current session",
		SuccessDescription: "Session terminated",
		Success:            httpx.Message(http.StatusOK),
		Guards:             []httpx.Guard{jwt.AccessGuard()},
		Handler:            h.Logout,
	})

	return g
}
