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
	Predictions  []database.Prediction
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
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return
	}
	uData := UserLoginData{Name: user.Username}
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Printf("error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	symbolList := r.PostForm["stocks"]
	tmpl, err := template.ParseFiles("static/stockAnalysisRequestComplete.html")
	if err != nil {
		log.Printf("Error parsing login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	dataSlice, err := analysis.MakeWeeklyDataSlice(r.Context(), symbolList)
	if err != nil {
		log.Printf("Error creating data slice for analysis: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	DataMap, err := analysis.StoreWeeklyDataV1(dataSlice, "")
	if err != nil {
		log.Printf("Error colelcting weekly stock data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	Pred, err := analysis.AnalyzeStoredDataV1(DataMap)
	if err != nil {
		log.Printf("Error analyzing stored stock data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	for _, prediction := range Pred {
		log.Printf("AddPrediction begin called with %s (predictable), %s (predictor) and %f (correlation)", prediction.PredictableSym, prediction.PredictorSym, prediction.Correlation)
		h.StockDBService.AddPrediction(r.Context(), prediction.PredictableSym, prediction.PredictorSym, prediction.Correlation, "First Draft", user.ID)
	}
	err = tmpl.Execute(w, PageData{UserData: uData, Error: nil})
	if err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) ShowPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(database.User)
	if !ok {
		log.Printf("Error: User not found in context for StockHandler. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusNotFound)
		return
	}

	predictions, err := h.StockDBService.GetAllPredictionsForUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error fetching predictions: %v", err)
		http.Error(w, "Could not retrieve predictions.", http.StatusInternalServerError)
		return
	}
	data := PageData{
		Predictions: predictions,
	}
	tmpl, err := template.ParseFiles("static/show_predictions.html")
	if err != nil {
		log.Printf("Error parsing show_predictions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing show_predictions template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
