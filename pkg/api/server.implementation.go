package api

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ServerImplementation struct {
	Lock sync.Mutex
	DB   *gorm.DB
}

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func (s *ServerImplementation) GetMovieByName(ctx echo.Context, name string) error {
	return nil
}

func (s *ServerImplementation) GetMovieBygenre(ctx echo.Context, genre string) error {
	return nil
}

func (s *ServerImplementation) GetMovieByCastMember(ctx echo.Context, castmember string) error {
	return nil
}

func (s *ServerImplementation) GetMovieByYear(ctx echo.Context, year int64) error {

	var movies []Movie
	tx := s.DB.Where("year = ?", year).Find(&movies)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error)
	}

	return ctx.JSON(http.StatusOK, movies)
}

func (s *ServerImplementation) UploadMovie(ctx echo.Context) error {

	var newMovie Movie
	err := ctx.Bind(&newMovie)
	if err != nil {
		return err
	}
	s.Lock.Lock()
	defer s.Lock.Unlock()

	tx := s.DB.Create(&newMovie)
	if tx.Error != nil {
		return tx.Error
	}

	return ctx.JSON(http.StatusOK, "")
}
