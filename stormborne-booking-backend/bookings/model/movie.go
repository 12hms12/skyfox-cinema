package model

type Movie struct {
	MovieId  string `json:"id"`
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Plot     string `json:"plot"`
	Poster   string `json:"poster"`
	Genre	 string `json:"genre"`
	ImdbRating string `json:"imdbRating"`
	Rated    string `json:"rated"`
	ImdbVotes string `json:"imdbVotes"`
}

func NewMovie(id string, name string, duration string, plot string, poster string, genre string, imdbrating string,rated string,imdbvotes string) *Movie {
	return &Movie{
		MovieId:  id,
		Name:     name,
		Duration: duration,
		Plot:     plot,
		Poster:   poster,
		Genre:	  genre,
		ImdbRating: imdbrating,
		Rated: rated,
		ImdbVotes: imdbvotes,
	}
}
