package main

import (
	"net/http"
)

// Handler to return application health status
func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	// Create a map to hold the health status data
	data := envelope{
		"status":      "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverError(w, r, err)
	}
}