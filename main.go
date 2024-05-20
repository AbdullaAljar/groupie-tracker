package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

type Artist struct {
	ID              int      `json:"id"`
	Image           string   `json:"image"`
	Name            string   `json:"name"`
	Members         []string `json:"members"`
	CreationDate    int      `json:"creationDate"`
	FirstAlbum      string   `json:"firstAlbum"`
	LocationsURL    string   `json:"locations"`
	Locations       []string
	ConcertDatesURL string `json:"concertDates"`
	ConcertDates    []string
	RelationsURL    string `json:"relations"`
	Relations       map[string][]string
}

var portNumber string = "2156" //change to the preferred port number

func main() {
	artists := fetchArtists()

	artistHandler := func(w http.ResponseWriter, r *http.Request) {
		artistTemplate, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Error rendering home page", http.StatusInternalServerError)
			return
		}
		artistTemplate.Execute(w, artists)
	}

	detailsHandler := func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			http.Error(w, "Missing artist ID", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid artist ID", http.StatusBadRequest)
			return
		}

		artist, err := fetchArtistDetails(id)
		if err != nil {
			http.Error(w, "Artist not found", http.StatusNotFound)
			return
		}

		detailsTemplate, err := template.ParseFiles("templates/details.html")
		if err != nil {
			http.Error(w, "Error rendering artist details", http.StatusInternalServerError)
			return
		}
		detailsTemplate.Execute(w, artist)
	}

	// errorHandler := func(w http.ResponseWriter, r *http.Request) {
	// 	//TO-DO
	// }

	fs := http.FileServer(http.Dir("css"))
	http.Handle("/css/", http.StripPrefix("/css/", fs))

	// fsFavicon := http.FileServer(http.Dir("favicon"))
	// http.Handle("/favicon/", http.StripPrefix("/favicon/", fsFavicon))

	http.HandleFunc("/", artistHandler)
	http.HandleFunc("/artist", detailsHandler)

	log.Printf("Server starting on http://localhost:%s", portNumber)
	err := http.ListenAndServe(":"+portNumber, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func fetchArtists() []Artist {
	url := "https://groupietrackers.herokuapp.com/api/artists"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching data: %s", err)
	}
	defer resp.Body.Close()

	var artists []Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		log.Fatalf("Error decoding JSON: %s", err)
	}

	return artists
}

func fetchArtistDetails(id int) (*Artist, error) {
	// Fetch artist basic details
	url := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/artists/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching artist details: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch artist details, status code: %d", resp.StatusCode)
	}

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		return nil, fmt.Errorf("error decoding artist details JSON: %s", err)
	}

	// Fetch artist's locations
	locURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/locations/%d", id)
	artist.LocationsURL = locURL
	locResp, err := http.Get(locURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching artist locations: %s", err)
	}
	defer locResp.Body.Close()

	if locResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch artist locations, status code: %d", locResp.StatusCode)
	}

	var locations struct {
		Locations []string `json:"locations"`
	}
	err = json.NewDecoder(locResp.Body).Decode(&locations)
	if err != nil {
		return nil, fmt.Errorf("error decoding artist locations JSON: %s", err)
	}
	artist.Locations = locations.Locations

	// Fetch artist's dates
	dateURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/dates/%d", id)
	artist.ConcertDatesURL = dateURL
	dateResp, err := http.Get(dateURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching artist dates: %s", err)
	}
	defer dateResp.Body.Close()

	if dateResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch artist dates, status code: %d", dateResp.StatusCode)
	}

	var dates struct {
		Dates []string `json:"dates"`
	}
	err = json.NewDecoder(dateResp.Body).Decode(&dates)
	if err != nil {
		return nil, fmt.Errorf("error decoding artist dates JSON: %s", err)
	}
	artist.ConcertDates = dates.Dates

	// Fetch artist's relations
	relURL := fmt.Sprintf("https://groupietrackers.herokuapp.com/api/relation/%d", id)
	artist.RelationsURL = relURL
	relResp, err := http.Get(relURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching artist relations: %s", err)
	}
	defer relResp.Body.Close()

	if relResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch artist relations, status code: %d", relResp.StatusCode)
	}

	var relations struct {
		DatesLocations map[string][]string `json:"datesLocations"`
	}
	err = json.NewDecoder(relResp.Body).Decode(&relations)
	if err != nil {
		return nil, fmt.Errorf("error decoding artist relations JSON: %s", err)
	}

	artist.Relations = relations.DatesLocations

	return &artist, nil
}
