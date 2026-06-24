package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	
	"github.com/berylm1/iot-analytics-dashboard/internal/models"
	"github.com/berylm1/iot-analytics-dashboard/internal/storage"
)

type Handler struct {
	store storage.Store
}

func NewHandler(store storage.Store) *Handler {
	return &Handler{store: store}
}

// GetAggregations returns aggregated data with optional limit
func (h *Handler) GetAggregations(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	aggregations := h.store.GetRecent(limit)
	
	filtered := make([]models.AggregatedData, 0)
		for _, agg := range aggregations {
		if agg.DeviceCount > 0 {
			filtered = append(filtered, agg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(filtered); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetLatestAggregation returns the most recent aggregation
func (h *Handler) GetLatestAggregation(w http.ResponseWriter, r *http.Request) {
	aggregations := h.store.GetRecent(1)

	w.Header().Set("Content-Type", "application/json")
	if len(aggregations) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No data available",
			"data":    nil,
		})
		return
	}

	if err := json.NewEncoder(w).Encode(aggregations[0]); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetStats returns overall statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	aggregations := h.store.GetRecent(100)

	if len(aggregations) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "No data available",
		})
		return
	}

	var totalTemp, totalHumidity float64
	var minTemp, maxTemp float64
	totalDevices := 0

	minTemp = aggregations[0].MinTemperature
	maxTemp = aggregations[0].MaxTemperature

	for _, agg := range aggregations {
		totalTemp += agg.AvgTemperature
		totalHumidity += agg.AvgHumidity
		totalDevices += agg.DeviceCount

		if agg.MinTemperature < minTemp {
			minTemp = agg.MinTemperature
		}
		if agg.MaxTemperature > maxTemp {
			maxTemp = agg.MaxTemperature
		}
	}

	stats := map[string]interface{}{
		"avgTemperature": totalTemp / float64(len(aggregations)),
		"avgHumidity":    totalHumidity / float64(len(aggregations)),
		"minTemperature": minTemp,
		"maxTemperature": maxTemp,
		"totalDevices":   totalDevices,
		"windowCount":    len(aggregations),
		"averageDevices": float64(totalDevices) / float64(len(aggregations)),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
