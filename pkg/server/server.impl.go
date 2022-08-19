package server

import (
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MovieServer struct {
	Lock sync.Mutex
	db   *gorm.DB
}

func NewMovieServer(db *gorm.DB) *MovieServer {
	return &MovieServer{
		db: db,
	}
}

func (m *MovieServer) UploadMovie(ctx echo.Context) error {

	var newMovie Movie
	err := ctx.Bind(&newMovie)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid request body")
	}
	m.Lock.Lock()
	defer m.Lock.Unlock()

	tx := m.db.Create(&newMovie)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error.Error())
	}
	return ctx.JSON(http.StatusOK, "")
}

func (m *MovieServer) GetMovieByCastMember(ctx echo.Context, castmember string) error {

	var movie []Movie
	tx := m.db.Where(&Movie{Cast: []string{
		castmember,
	}}).Find(&movie)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error.Error())
	}
	if len(movie) == 0 {
		return ctx.JSON(http.StatusNotFound, "")
	}

	return nil
}

func (m *MovieServer) GetMovieBygenre(ctx echo.Context, genre string) error {

	var movie []Movie
	tx := m.db.Where(&Movie{Genres: []string{
		genre,
	}}).Find(&movie)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error.Error())
	}
	if len(movie) == 0 {
		return ctx.JSON(http.StatusNotFound, "")
	}

	return ctx.JSON(http.StatusOK, movie)
}

func (m *MovieServer) GetMovieByName(ctx echo.Context, name string) error {

	var movie []Movie
	tx := m.db.Where(&Movie{Title: name}).Find(&movie)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error.Error())
	}
	if len(movie) == 0 {
		return ctx.JSON(http.StatusNotFound, "")
	}
	return ctx.JSON(http.StatusOK, movie)
}

func (m *MovieServer) GetMovieByYear(ctx echo.Context, year int64) error {

	var movie []Movie
	tx := m.db.Where(&Movie{Year: year}).Find(&movie)
	if tx.Error != nil {
		return ctx.JSON(http.StatusBadRequest, tx.Error.Error())
	}
	if len(movie) == 0 {
		return ctx.JSON(http.StatusNotFound, "")
	}

	return ctx.JSON(http.StatusOK, movie)
}
