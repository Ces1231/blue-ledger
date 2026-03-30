package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/google/uuid"
)

// Handler handles all auth HTTP endpoints.
type Handler struct {
	svc              Service
	loginBonusRepo   *LoginBonusRepository
}

// NewHandler creates a new auth handler.
func NewHandler(svc Service, loginBonusRepo *LoginBonusRepository) *Handler {
	return &Handler{
		svc:            svc,
		loginBonusRepo: loginBonusRepo,
	}
}

// RegisterRoutes mounts auth routes on the given Echo group.
func (h *Handler) RegisterRoutes(g *echo.Group, jwtMiddleware echo.MiddlewareFunc) {
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/magic-link", h.SendMagicLink)
	g.GET("/magic", h.ConsumeMagicLink)
	g.POST("/refresh", h.Refresh)
	g.POST("/logout", h.Logout, jwtMiddleware)
	g.GET("/me", h.Me, jwtMiddleware)
}

// --- Request / Response types ---

type registerRequest struct {
	ChapterName  string `json:"chapter_name" validate:"required,min=3"`
	GreekLetters string `json:"greek_letters" validate:"required"`
	City         string `json:"city" validate:"required"`
	StateCode    string `json:"state_code" validate:"required,len=2"`
	University   string `json:"university"`
	FirstName    string `json:"first_name" validate:"required"`
	LastName     string `json:"last_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type magicLinkRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register handles POST /auth/register
// Creates a new chapter + admin user and returns auth tokens.
func (h *Handler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	input := RegisterInput{
		ChapterName:  req.ChapterName,
		GreekLetters: req.GreekLetters,
		City:         req.City,
		StateCode:    req.StateCode,
		University:   req.University,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Password:     req.Password,
	}

	resp, err := h.svc.Register(c.Request().Context(), input, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return echo.NewHTTPError(http.StatusConflict, "email address is already registered")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "registration failed")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{"data": resp})
}

// Login handles POST /auth/login
func (h *Handler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.svc.Login(c.Request().Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}, c.Request().UserAgent(), c.RealIP())

	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		}
		if errors.Is(err, ErrAccountLocked) {
			return echo.NewHTTPError(http.StatusTooManyRequests, "account is temporarily locked")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
	}

	// Award daily login bonus
	if resp != nil && resp.User.ID != "" {
		userID, err := uuid.Parse(resp.User.ID)
		if err == nil {
			// Don't block login if bonus fails, just log it
			_, _ = h.loginBonusRepo.AwardDailyLoginBonus(c.Request().Context(), userID)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": resp})
}

// SendMagicLink handles POST /auth/magic-link
func (h *Handler) SendMagicLink(c echo.Context) error {
	var req magicLinkRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	// Always return success to avoid email enumeration
	_ = h.svc.SendMagicLink(c.Request().Context(), req.Email)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "If that email is registered, a magic link has been sent"},
	})
}

// ConsumeMagicLink handles GET /auth/magic?token=<raw>
func (h *Handler) ConsumeMagicLink(c echo.Context) error {
	rawToken := c.QueryParam("token")
	if rawToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}

	resp, err := h.svc.ConsumeMagicLink(c.Request().Context(), rawToken, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenUsed) || errors.Is(err, ErrTokenExpired) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid, expired, or already-used link")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "magic link validation failed")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": resp})
}

// Refresh handles POST /auth/refresh
func (h *Handler) Refresh(c echo.Context) error {
	var req refreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	resp, err := h.svc.RefreshToken(c.Request().Context(), req.RefreshToken, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrTokenUsed) || errors.Is(err, ErrTokenExpired) {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "token refresh failed")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": resp})
}

// Logout handles POST /auth/logout
func (h *Handler) Logout(c echo.Context) error {
	var req logoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	_ = h.svc.Logout(c.Request().Context(), req.RefreshToken)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": map[string]string{"message": "logged out"},
	})
}

// Me handles GET /auth/me — returns the current authenticated user's profile.
func (h *Handler) Me(c echo.Context) error {
	userID := GetUserID(c)
	if userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	user, err := h.svc.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"data": user})
}
