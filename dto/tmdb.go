package dto

// TMDBShowResponse represents the response from TMDB show details API
type TMDBShowResponse struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	OriginalName string           `json:"original_name"`
	Status       string           `json:"status"`
	FirstAirDate string           `json:"first_air_date"`
	Overview     string           `json:"overview"`
	PosterPath   string           `json:"poster_path"`
	BackdropPath string           `json:"backdrop_path"`
	Genres       []TMDBGenre      `json:"genres"`
	Popularity   float64          `json:"popularity"`
	VoteAverage  float32          `json:"vote_average"`
	VoteCount    int              `json:"vote_count"`
	Seasons      []TMDBSeasonInfo `json:"seasons"`
}

// TMDBGenre represents a genre
type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TMDBSeasonInfo represents season information
type TMDBSeasonInfo struct {
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	AirDate      string `json:"air_date"`
}

// TMDBSeasonResponse represents the response from TMDB season details API
type TMDBSeasonResponse struct {
	ID           int           `json:"id"`
	SeasonNumber int           `json:"season_number"`
	Name         string        `json:"name"`
	Overview     string        `json:"overview"`
	Episodes     []TMDBEpisode `json:"episodes"`
}

// TMDBEpisode represents an episode from TMDB
type TMDBEpisode struct {
	ID            int     `json:"id"`
	SeasonNumber  int     `json:"season_number"`
	EpisodeNumber int     `json:"episode_number"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	AirDate       string  `json:"air_date"`
	StillPath     string  `json:"still_path"`
	Runtime       int     `json:"runtime"`
	VoteAverage   float32 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
}

// TMDBSearchResponse represents the response from TMDB search API
type TMDBSearchResponse struct {
	Page         int              `json:"page"`
	Results      []TMDBShowResult `json:"results"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
}

// TMDBShowResult represents a show in search results
type TMDBShowResult struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	PosterPath   string  `json:"poster_path"`
	FirstAirDate string  `json:"first_air_date"`
	Overview     string  `json:"overview"`
	VoteAverage  float32 `json:"vote_average"`
}

// TMDBErrorResponse represents an error response from TMDB
type TMDBErrorResponse struct {
	StatusCode    int    `json:"status_code"`
	StatusMessage string `json:"status_message"`
	Success       bool   `json:"success"`
}
