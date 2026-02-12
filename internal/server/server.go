// Package server implements the tishi HTTP API server.
package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

// Server is the tishi HTTP API server.
type Server struct {
	pool   *pgxpool.Pool
	log    *zap.Logger
	cfg    *config.Config
	router chi.Router
}

// New creates a new Server instance.
func New(pool *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *Server {
	s := &Server{
		pool: pool,
		log:  log,
		cfg:  cfg,
	}
	s.router = s.setupRouter()
	return s
}

// Router returns the chi.Router for the server.
func (s *Server) Router() chi.Router {
	return s.router
}

// setupRouter configures all routes and middleware.
func (s *Server) setupRouter() chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(zapLoggerMiddleware(s.log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/healthz", s.handleHealthz)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Throttle(100))

		r.Get("/rankings", s.handleGetRankings)
		r.Get("/projects", s.handleListProjects)
		r.Get("/projects/{id}", s.handleGetProject)
		r.Get("/projects/{id}/trends", s.handleGetProjectTrends)
		r.Get("/posts", s.handleListPosts)
		r.Get("/posts/{slug}", s.handleGetPost)
		r.Get("/categories", s.handleListCategories)
	})

	return r
}

// --- Response helpers ---

// apiResponse is the standard JSON response wrapper.
type apiResponse struct {
	Data interface{} `json:"data"`
	Meta *apiMeta    `json:"meta,omitempty"`
}

type apiMeta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	resp := apiError{}
	resp.Error.Code = code
	resp.Error.Message = message
	writeJSON(w, status, resp)
}

// parsePagination extracts page and per_page from query params.
func parsePagination(r *http.Request) (page, perPage, offset int) {
	page = 1
	perPage = 20

	if v := r.URL.Query().Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := r.URL.Query().Get("per_page"); v != "" {
		if pp, err := strconv.Atoi(v); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}
	offset = (page - 1) * perPage
	return
}

// --- Handlers ---

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	var dbStatus string
	if s.pool == nil {
		dbStatus = "disconnected"
	} else if err := s.pool.Ping(r.Context()); err != nil {
		dbStatus = "disconnected"
	} else {
		dbStatus = "connected"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "ok",
		"version":  "0.1.0-dev",
		"database": dbStatus,
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleGetRankings(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := parsePagination(r)
	maxRank := 100

	rows, err := s.pool.Query(r.Context(), `
		SELECT p.id, p.full_name, p.description, p.language, p.license,
			p.stargazers_count, p.forks_count, p.open_issues_count,
			p.score, p.rank, p.topics, p.pushed_at
		FROM projects p
		WHERE p.rank IS NOT NULL AND p.rank <= $1 AND p.is_archived = FALSE
		ORDER BY p.rank ASC
		LIMIT $2 OFFSET $3`, maxRank, perPage, offset)
	if err != nil {
		s.log.Error("查询排行榜失败", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var (
			id, stars, forks, issues, rank           int
			score                                    float64
			fullName, description, language, license string
			topics                                   []string
			pushedAt                                 *time.Time
		)
		if err := rows.Scan(&id, &fullName, &description, &language, &license,
			&stars, &forks, &issues, &score, &rank, &topics, &pushedAt); err != nil {
			s.log.Error("扫描行失败", zap.Error(err))
			continue
		}

		project := map[string]interface{}{
			"id":          id,
			"full_name":   fullName,
			"description": description,
			"language":    language,
			"license":     license,
			"stars":       stars,
			"forks":       forks,
			"open_issues": issues,
			"score":       score,
			"rank":        rank,
			"topics":      topics,
			"pushed_at":   pushedAt,
		}
		projects = append(projects, project)
	}

	if projects == nil {
		projects = []map[string]interface{}{}
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, apiResponse{
		Data: projects,
		Meta: &apiMeta{
			Total:      len(projects),
			Page:       page,
			PerPage:    perPage,
			TotalPages: (maxRank + perPage - 1) / perPage,
		},
	})
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := parsePagination(r)

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, full_name, description, language, stargazers_count, score, rank
		FROM projects
		WHERE is_archived = FALSE
		ORDER BY score DESC
		LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var (
			id, stars   int
			rank        *int
			score       float64
			fullName    string
			description *string
			language    *string
		)
		if err := rows.Scan(&id, &fullName, &description, &language, &stars, &score, &rank); err != nil {
			continue
		}
		projects = append(projects, map[string]interface{}{
			"id":          id,
			"full_name":   fullName,
			"description": description,
			"language":    language,
			"stars":       stars,
			"score":       score,
			"rank":        rank,
		})
	}

	if projects == nil {
		projects = []map[string]interface{}{}
	}

	// Get total count
	var total int
	s.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM projects WHERE is_archived = FALSE`).Scan(&total)

	writeJSON(w, http.StatusOK, apiResponse{
		Data: projects,
		Meta: &apiMeta{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: (total + perPage - 1) / perPage,
		},
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid project ID")
		return
	}

	var (
		projectID, stars, forks, issues, watchers int
		rank                                      *int
		score                                     float64
		githubID                                  int64
		fullName, language, license               string
		description, homepage                     *string
		topics                                    []string
		pushedAt, createdAtGH, firstSeenAt        *time.Time
	)

	err = s.pool.QueryRow(r.Context(), `
		SELECT id, github_id, full_name, description, language, license,
			topics, homepage, created_at_gh, pushed_at,
			stargazers_count, forks_count, open_issues_count, watchers_count,
			score, rank, first_seen_at
		FROM projects WHERE id = $1`, id).Scan(
		&projectID, &githubID, &fullName, &description, &language, &license,
		&topics, &homepage, &createdAtGH, &pushedAt,
		&stars, &forks, &issues, &watchers,
		&score, &rank, &firstSeenAt,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Project not found")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, apiResponse{
		Data: map[string]interface{}{
			"id":            projectID,
			"github_id":     githubID,
			"full_name":     fullName,
			"description":   description,
			"language":      language,
			"license":       license,
			"topics":        topics,
			"homepage":      homepage,
			"stars":         stars,
			"forks":         forks,
			"open_issues":   issues,
			"watchers":      watchers,
			"score":         score,
			"rank":          rank,
			"pushed_at":     pushedAt,
			"created_at_gh": createdAtGH,
			"first_seen_at": firstSeenAt,
			"github_url":    "https://github.com/" + fullName,
		},
	})
}

func (s *Server) handleGetProjectTrends(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "Invalid project ID")
		return
	}

	days := 30
	if v := r.URL.Query().Get("days"); v != "" {
		if d, err := strconv.Atoi(v); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	rows, err := s.pool.Query(r.Context(), `
		SELECT snapshot_date, stargazers_count, forks_count, open_issues_count, score, rank
		FROM daily_snapshots
		WHERE project_id = $1
		  AND snapshot_date >= CURRENT_DATE - $2::INT
		ORDER BY snapshot_date ASC`, id, days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}
	defer rows.Close()

	var trends []map[string]interface{}
	for rows.Next() {
		var (
			date                 time.Time
			stars, forks, issues int
			score                *float64
			rank                 *int
		)
		if err := rows.Scan(&date, &stars, &forks, &issues, &score, &rank); err != nil {
			continue
		}
		trends = append(trends, map[string]interface{}{
			"date":        date.Format("2006-01-02"),
			"stars":       stars,
			"forks":       forks,
			"open_issues": issues,
			"score":       score,
			"rank":        rank,
		})
	}
	if trends == nil {
		trends = []map[string]interface{}{}
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, apiResponse{Data: trends})
}

func (s *Server) handleListPosts(w http.ResponseWriter, r *http.Request) {
	page, perPage, offset := parsePagination(r)

	rows, err := s.pool.Query(r.Context(), `
		SELECT id, title, slug, post_type, published_at, LEFT(content, 200) AS excerpt
		FROM blog_posts
		WHERE published_at IS NOT NULL
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2`, perPage, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}
	defer rows.Close()

	var posts []map[string]interface{}
	for rows.Next() {
		var (
			id          int
			title, slug string
			postType    string
			publishedAt *time.Time
			excerpt     *string
		)
		if err := rows.Scan(&id, &title, &slug, &postType, &publishedAt, &excerpt); err != nil {
			continue
		}
		posts = append(posts, map[string]interface{}{
			"id":           id,
			"title":        title,
			"slug":         slug,
			"type":         postType,
			"published_at": publishedAt,
			"excerpt":      excerpt,
		})
	}
	if posts == nil {
		posts = []map[string]interface{}{}
	}

	var total int
	s.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM blog_posts WHERE published_at IS NOT NULL`).Scan(&total)

	w.Header().Set("Cache-Control", "public, max-age=86400")
	writeJSON(w, http.StatusOK, apiResponse{
		Data: posts,
		Meta: &apiMeta{Total: total, Page: page, PerPage: perPage, TotalPages: (total + perPage - 1) / perPage},
	})
}

func (s *Server) handleGetPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var (
		id          int
		title       string
		content     string
		postType    string
		publishedAt *time.Time
		createdAt   time.Time
	)
	err := s.pool.QueryRow(r.Context(), `
		SELECT id, title, content, post_type, published_at, created_at
		FROM blog_posts WHERE slug = $1`, slug).Scan(
		&id, &title, &content, &postType, &publishedAt, &createdAt,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Post not found")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	writeJSON(w, http.StatusOK, apiResponse{
		Data: map[string]interface{}{
			"id":           id,
			"title":        title,
			"slug":         slug,
			"content":      content,
			"type":         postType,
			"published_at": publishedAt,
			"created_at":   createdAt,
		},
	})
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT c.id, c.name, c.slug, c.description,
			COUNT(pc.project_id) AS project_count
		FROM categories c
		LEFT JOIN project_categories pc ON c.id = pc.category_id
		GROUP BY c.id
		ORDER BY c.sort_order ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}
	defer rows.Close()

	var categories []map[string]interface{}
	for rows.Next() {
		var (
			id, count   int
			name, slug  string
			description *string
		)
		if err := rows.Scan(&id, &name, &slug, &description, &count); err != nil {
			continue
		}
		categories = append(categories, map[string]interface{}{
			"id":            id,
			"name":          name,
			"slug":          slug,
			"description":   description,
			"project_count": count,
		})
	}
	if categories == nil {
		categories = []map[string]interface{}{}
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	writeJSON(w, http.StatusOK, apiResponse{Data: categories})
}
