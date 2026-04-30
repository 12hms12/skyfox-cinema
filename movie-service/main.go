package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Movie represents a movie structure
type Movie struct {
	Title      string   `json:"title"`
	Year       string   `json:"year"`
	Rated      string   `json:"rated"`
	Released   string   `json:"released"`
	Runtime    string   `json:"runtime"`
	Genre      string   `json:"genre"`
	Director   string   `json:"director"`
	Writer     string   `json:"writer"`
	Actors     string   `json:"actors"`
	Plot       string   `json:"plot"`
	Language   string   `json:"language"`
	Country    string   `json:"country"`
	Awards     string   `json:"awards"`
	Poster     string   `json:"poster"`
	Ratings    []Rating `json:"ratings"`
	Metascore  string   `json:"metascore"`
	ImdbRating string   `json:"imdbrating"`
	ImdbVotes  string   `json:"imdbvotes"`
	ImdbID     string   `json:"imdbid"`
	Type       string   `json:"type"`
	DVD        string   `json:"dvd"`
	BoxOffice  string   `json:"boxoffice"`
	Production string   `json:"production"`
	Website    string   `json:"website"`
	Response   string   `json:"response"`
}

// Rating represents a movie rating
type Rating struct {
	Source string `json:"source"`
	Value  string `json:"value"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

var movies []Movie

// loadMovies loads movies from JSON file
func loadMovies() error {
	data, err := ioutil.ReadFile("movies.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &movies)
}

// getMovies returns all movies
func getMovies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movies)
}

// getMovieByID returns a single movie by ID
func getMovieByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	for _, movie := range movies {
		if movie.ImdbID == id {
			json.NewEncoder(w).Encode(movie)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(ErrorResponse{Error: "Movie with requested ID not found"})
}

// getRoot returns a random quote with color
func getRoot(w http.ResponseWriter, r *http.Request) {
	quotes := []string{
		"The only way to do great work is to love what you do - Steve Jobs",
		"Innovation distinguishes between a leader and a follower - Steve Jobs",
		"Stay hungry, stay foolish - Steve Jobs",
		"Life is what happens when you're busy making other plans - John Lennon",
		"The future belongs to those who believe in the beauty of their dreams - Eleanor Roosevelt",
	}

	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#FFA07A", "#98D8C8", "#F7DC6F", "#BB8FCE"}

	rand.Seed(time.Now().UnixNano())
	quote := quotes[rand.Intn(len(quotes))]
	color := colors[rand.Intn(len(colors))]

	html := fmt.Sprintf(`
		<html>
		<body style="display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; font-family: Arial, sans-serif;">
			<p style="text-align: center; font-size: 2em; color: %s; max-width: 80%%;">
				%s
			</p>
		</body>
		</html>
	`, color, quote)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// notFound handles 404 errors
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "There is nothing to do here.! 404!")
}

func main() {
	// Load movies from JSON file
	if err := loadMovies(); err != nil {
		log.Fatal("Error loading movies:", err)
	}

	log.Printf("Loaded %d movies from movies.json\n", len(movies))

	// Create router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/", getRoot).Methods("GET")
	router.HandleFunc("/movies", getMovies).Methods("GET")
	router.HandleFunc("/movies/{id}", getMovieByID).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(notFound)

	// Start server
	port := "4567"
	log.Printf("Movie service starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
