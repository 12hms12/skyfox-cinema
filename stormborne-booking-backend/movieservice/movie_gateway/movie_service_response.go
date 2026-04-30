package movieservice

import (
	"fmt"
	"skyfox/bookings/constants"
	"skyfox/bookings/model"
	"skyfox/common/logger"
	ae "skyfox/error"
	"strings"
	"time"
)

type MovieServiceResponse struct {
	ImdbId  string `json:"imdbid"`
	Title   string `json:"title"`
	RunTime string `json:"runtime"`
	Plot    string `json:"plot"`
	Poster  string `json:"poster"`
	Genre	 string `json:"genre"`
	ImdbRating string `json:"imdbRating"`
	Rated   string   `json:"rated"`
	ImdbVotes string `json:"imdbVotes"`
}

func (m MovieServiceResponse) ToMovie() (*model.Movie, error) {
	runtime := strings.Split(m.RunTime, " ")
	if m.RunTime == "" || runtime[0] == "" {
		return &model.Movie{}, ae.UnProcessableError("MovieCreationFailed", "Invalid runtime format", nil)
	}
	duration, err := time.ParseDuration(runtime[0] + "m")

	if err != nil {
		if logger.GetLogger() != nil {
			logger.Error(fmt.Sprintf("failed to get the run time of the movie %s", m.Title), err)
		}
		return &model.Movie{}, ae.UnProcessableError("MovieCreationFailed", "Movie creation failed due to unknown reason", err)
	}

	if m.Poster == "" || m.Poster == "N/A" {
		m.Poster = constants.StockImage
	}

	movie := model.NewMovie(m.ImdbId, m.Title, duration.String(), m.Plot, m.Poster, m.Genre, m.ImdbRating,m.Rated,m.ImdbVotes)
	return movie, nil
}
