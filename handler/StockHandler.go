package handler

import (
	"log"
	"net/http"
	"text/template"

	"github.com/ethanjameslong1/GoCloudProject.git/analysis"
	"github.com/ethanjameslong1/GoCloudProject.git/database"
)

type PageData struct {
	UserData     UserLoginData
	Error        error
	StockWeights analysis.StockWeights
	Interval     string
}

func (h *Handler) StockRequestPageHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Error: User not found in context for StockHandler. Redirecting to login.")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	data := PageData{
		UserData: UserLoginData{
			Name: user.Username,
		},
		Error:    nil,
		Interval: "Weekly",
	}

	tmpl, err := template.ParseFiles("static/stockAnalysisRequest.html")
	if err != nil {
		log.Printf("Error parsing stock template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) StockRequestHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Error: User not found in context for StockHandler. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusFound) // StatusFound (302) is common for redirection
		return
	}
	uData := UserLoginData{Name: user.Username}
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	sData := analysis.StockWeights{OpenPriceWeight: r.FormValue("weightCurrentOpen"), HighPriceWeight: r.FormValue("weightCurrentHigh"), ClosePriceWeight: r.FormValue("weightCurrentClose"), LowPriceWeight: r.FormValue("weightCurrentLow"), VolumeWeight: r.FormValue("weightCurrentVolume"), PercChangeWeight: r.FormValue("weightCloseOpenPctChange"), PercRangeWeight: r.FormValue("weightHighLowPctRange")}
	tmpl, err := template.ParseFiles("static/stockAnalysisRequestComplete.html")
	if err != nil {
		log.Printf("Error parsing login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	dataSlice, err := analysis.MakeWeeklyDataSlice(r.Context(), analysis.AlphaVantageSymbols)
	if err != nil {
		log.Printf("Error creating data slice for analysis: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	analysis.StoreWeeklyData(dataSlice, "", sData)

	err = tmpl.Execute(w, PageData{UserData: uData, StockWeights: sData, Error: nil})
	if err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}
