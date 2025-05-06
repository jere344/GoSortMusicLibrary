package main

import (
	"encoding/json"
	"fmt" // Import fmt for error formatting
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings" // Import strings
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/jere344/gosortmusiclibrary/internal/sorter" // Import the sorter package
)

type App struct {
	Router *mux.Router
}

type SortRequest struct {
	SourceFolder      string `json:"sourceFolder"`
	DestinationFolder string `json:"destinationFolder"`
	Script            string `json:"script"`
	FileOperationMode string `json:"fileOperationMode"`
}

type SortResponse struct {
	Logs []string `json:"logs"`
}

func main() {
	app := &App{}
	app.Initialize()
	app.Run(":8080")
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// Serve static files
	staticDir := "/static/"
	a.Router.
		PathPrefix(staticDir).
		Handler(http.StripPrefix(staticDir, http.FileServer(http.Dir("./web/static/"))))

	a.Router.HandleFunc("/", a.homePage).Methods("GET")
	a.Router.HandleFunc("/sort", a.sortMusicLibrary).Methods("POST")
}

func (a *App) homePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	tmpl.Execute(w, nil)
}

func (a *App) sortMusicLibrary(w http.ResponseWriter, r *http.Request) {
	var req SortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Println("Error decoding request body:", err)
		// Send error log back to client
		resp := SortResponse{Logs: []string{fmt.Sprintf("Server Error: Invalid request body - %v", err)}}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer r.Body.Close()

	// Basic validation
	req.FileOperationMode = strings.ToLower(req.FileOperationMode)
	if req.SourceFolder == "" || req.DestinationFolder == "" || req.Script == "" ||
		(req.FileOperationMode != "move" && req.FileOperationMode != "copy" && req.FileOperationMode != "preview") {
		errMsg := "Missing required fields or invalid file operation mode (must be 'preview', 'move', or 'copy')"
		http.Error(w, errMsg, http.StatusBadRequest)
		log.Println(errMsg, "in request")
		// Send error log back to client
		resp := SortResponse{Logs: []string{fmt.Sprintf("Client Error: %s", errMsg)}}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Received sort request: Source='%s', Destination='%s', Mode='%s'", req.SourceFolder, req.DestinationFolder, req.FileOperationMode)

	// Create a temporary file for the script
	tmpDir := os.TempDir()
	scriptFileName := filepath.Join(tmpDir, "sortscript_"+time.Now().Format("20060102150405")+".script")
	err := os.WriteFile(scriptFileName, []byte(req.Script), 0644)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create temporary script file: %v", err)
		http.Error(w, errMsg, http.StatusInternalServerError)
		log.Println("Error creating temp script file:", err)
		// Send error log back to client
		resp := SortResponse{Logs: []string{fmt.Sprintf("Server Error: %s", errMsg)}}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer os.Remove(scriptFileName) // Clean up the temporary file

	// Create and configure the sorter
	s := sorter.NewSorter(req.SourceFolder, scriptFileName, req.DestinationFolder, req.FileOperationMode) // Pass mode

	// Execute the sort operation and get logs
	logs, err := s.ExecuteSort() // Capture logs and error
	if err != nil {
		// Log the server-side error
		log.Println("Error executing sort:", err)
		// Append the final error message to the logs being sent to the client
		logs = append(logs, fmt.Sprintf("Error during sorting process: %v", err))
		// Send logs (including the error) back to the client with an error status
		resp := SortResponse{Logs: logs}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Indicate server error
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Send successful logs back to the client
	resp := SortResponse{Logs: logs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	log.Println("Sort request processed successfully.")
}

func (a *App) Run(addr string) {
	log.Println("Starting server on", addr)
	if err := http.ListenAndServe(addr, a.Router); err != nil {
		log.Fatal(err)
	}
}
