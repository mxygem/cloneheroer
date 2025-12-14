package server

import (
	"net/http"
	"strconv"
	"time"

	"cloneheroer/internal/db"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Server wraps Echo and database repo.
type Server struct {
	app  *echo.Echo
	repo *db.Repo
}

// New creates a configured server instance.
func New(repo *db.Repo) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Enable CORS for frontend
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	s := &Server{
		app:  e,
		repo: repo,
	}
	s.registerRoutes()
	return s
}

// Start runs the HTTP server.
func (s *Server) Start(addr string) error {
	return s.app.Start(addr)
}

func (s *Server) registerRoutes() {
	s.app.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
	})

	s.app.GET("/scores", s.handleListScores)
	s.app.GET("/artists", s.handleListArtists)
	s.app.GET("/songs", s.handleListSongs)
	s.app.PATCH("/artists/:id", s.handleUpdateArtist)
	s.app.PATCH("/songs/:id", s.handleUpdateSong)
	s.app.PATCH("/scores/:id", s.handleUpdateScore)
	s.app.PATCH("/players/:id", s.handleUpdatePlayer)

	// Debug route to list all registered routes (useful for troubleshooting)
	s.app.GET("/debug/routes", func(c echo.Context) error {
		routes := []map[string]string{}
		for _, route := range s.app.Routes() {
			routes = append(routes, map[string]string{
				"method": route.Method,
				"path":   route.Path,
				"name":   route.Name,
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"routes": routes,
		})
	})
}

func parseIDParam(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

type updateArtistRequest struct {
	Name *string `json:"name"`
}

func (s *Server) handleUpdateArtist(c echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	req := updateArtistRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if err := s.repo.UpdateArtist(c.Request().Context(), id, req.Name); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

type updateSongRequest struct {
	Name     *string  `json:"name"`
	ArtistID *int64   `json:"artist_id"`
	Charters []string `json:"charters"`
}

func (s *Server) handleUpdateSong(c echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	req := updateSongRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if err := s.repo.UpdateSong(c.Request().Context(), id, req.Name, req.ArtistID, req.Charters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

type updateScoreRequest struct {
	TotalScore *int64  `json:"total_score"`
	Stars      *int    `json:"stars_achieved"`
	Charter    *string `json:"charter"`
}

func (s *Server) handleUpdateScore(c echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	req := updateScoreRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if err := s.repo.UpdateScore(c.Request().Context(), id, req.TotalScore, req.Stars, req.Charter); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

type updatePlayerRequest struct {
	Name       *string  `json:"name"`
	Instrument *string  `json:"instrument"`
	Difficulty *string  `json:"difficulty"`
	Score      *int64   `json:"score"`
	Combo      *int     `json:"combo"`
	Accuracy   *float64 `json:"accuracy"`
	Misses     *int     `json:"misses"`
	Rank       *int     `json:"rank"`
}

func (s *Server) handleUpdatePlayer(c echo.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	req := updatePlayerRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}
	if err := s.repo.UpdatePlayer(
		c.Request().Context(),
		id,
		req.Name,
		req.Instrument,
		req.Difficulty,
		req.Score,
		req.Combo,
		req.Accuracy,
		req.Misses,
		req.Rank,
	); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *Server) handleListScores(c echo.Context) error {
	limitParam := c.QueryParam("limit")
	offsetParam := c.QueryParam("offset")
	var (
		limit  int32 = 20
		offset int32 = 0
	)
	if limitParam != "" {
		if v, err := strconv.Atoi(limitParam); err == nil {
			limit = int32(v)
		}
	}
	if offsetParam != "" {
		if v, err := strconv.Atoi(offsetParam); err == nil {
			offset = int32(v)
		}
	}

	scores, err := s.repo.ListScores(c.Request().Context(), limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, scores)
}

func (s *Server) handleListArtists(c echo.Context) error {
	limitParam := c.QueryParam("limit")
	offsetParam := c.QueryParam("offset")
	var (
		limit  int32 = 50
		offset int32 = 0
	)
	if limitParam != "" {
		if v, err := strconv.Atoi(limitParam); err == nil {
			limit = int32(v)
		}
	}
	if offsetParam != "" {
		if v, err := strconv.Atoi(offsetParam); err == nil {
			offset = int32(v)
		}
	}

	artists, err := s.repo.ListArtists(c.Request().Context(), limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, artists)
}

func (s *Server) handleListSongs(c echo.Context) error {
	limitParam := c.QueryParam("limit")
	offsetParam := c.QueryParam("offset")
	var (
		limit  int32 = 50
		offset int32 = 0
	)
	if limitParam != "" {
		if v, err := strconv.Atoi(limitParam); err == nil {
			limit = int32(v)
		}
	}
	if offsetParam != "" {
		if v, err := strconv.Atoi(offsetParam); err == nil {
			offset = int32(v)
		}
	}

	songs, err := s.repo.ListSongs(c.Request().Context(), limit, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, songs)
}

