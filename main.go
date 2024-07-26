package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type GameStats struct {
	GameName    string `json:"game_name"`
	LaunchCount int    `json:"launch_count"`
}

type MonthStats map[string][]GameStats

type YearStats map[string]MonthStats

type Stats struct {
	Month map[string][]GameStats `json:"month"`
	Year  map[string][]GameStats `json:"year"`
	All   YearStats              `json:"all"`
}

type Top10Stats struct {
	Period string      `json:"period"`
	Games  []GameStats `json:"games"`
}

var (
	logDir     string
	logPattern = regexp.MustCompile(`synchronet: term Node \d+ <\S+> running external program: (.+)`)
	stats      *Stats
	statsMutex sync.RWMutex
)

func main() {
	flag.StringVar(&logDir, "logdir", "/var/log", "Specify the directory containing log files")
	flag.Parse()

	// Load initial data
	refreshData()

	// Start background goroutine to refresh data every 24 hours
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			refreshData()
		}
	}()

	http.HandleFunc("/top10", handleTop10)
	http.HandleFunc("/stats", handleStats)

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", nil)
}

func refreshData() {
	fmt.Println("Refreshing data...")

	newStats := loadStats()
	files, err := filepath.Glob(filepath.Join(logDir, "syslog*"))
	if err != nil {
		fmt.Printf("Error reading log files: %v\n", err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file, ".gz") {
			processGzipLogFile(file, newStats)
		} else {
			processLogFile(file, newStats)
		}
	}

	statsMutex.Lock()
	stats = newStats
	statsMutex.Unlock()

	fmt.Println("Data refresh complete.")
}

func handleTop10(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, "Missing period query parameter", http.StatusBadRequest)
		return
	}
	period = strings.ToLower(period)

	statsMutex.RLock()
	top10Stats := getTop10Stats(stats, period)
	statsMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(top10Stats)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		http.Error(w, "Missing period query parameter", http.StatusBadRequest)
		return
	}
	period = strings.ToLower(period)

	statsMutex.RLock()
	defer statsMutex.RUnlock()

	var data []byte
	var err error

	switch {
	case period == "month" || isMonth(period):
		data, err = json.MarshalIndent(map[string]map[string][]GameStats{"month": stats.Month}, "", "  ")
	case period == "year" || isYear(period):
		data, err = json.MarshalIndent(map[string]map[string][]GameStats{"year": stats.Year}, "", "  ")
	case period == "all":
		data, err = json.MarshalIndent(map[string]YearStats{"all": stats.All}, "", "  ")
	default:
		http.Error(w, "Invalid period query parameter", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Error marshaling JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func loadStats() *Stats {
	return &Stats{
		Month: make(map[string][]GameStats),
		Year:  make(map[string][]GameStats),
		All:   make(YearStats),
	}
}

func processLogFile(file string, stats *Stats) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", file, err)
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		matches := logPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			processLogEntry(line, matches[1], stats)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file %s: %v\n", file, err)
	}
}

func processGzipLogFile(file string, stats *Stats) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening log file %s: %v\n", file, err)
		return
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		fmt.Printf("Error creating gzip reader for file %s: %v\n", file, err)
		return
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	for scanner.Scan() {
		line := scanner.Text()
		matches := logPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			processLogEntry(line, matches[1], stats)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file %s: %v\n", file, err)
	}
}

func processLogEntry(line, gameName string, stats *Stats) {
	timeLayout := "2006-01-02T15:04:05.999999-07:00"
	timePart := strings.Split(line, " ")[0]
	t, err := time.Parse(timeLayout, timePart)
	if err != nil {
		fmt.Printf("Error parsing time: %v\n", err)
		return
	}

	month := strings.ToLower(t.Format("January"))
	year := t.Format("2006")

	updateStats(stats, year, month, gameName)
}

func updateStats(stats *Stats, year, month, gameName string) {
	updateStatsMap(stats.Month, month, gameName)
	updateStatsMap(stats.Year, year, gameName)
	updateNestedStats(stats.All, year, month, gameName)
}

func updateStatsMap(stats map[string][]GameStats, key, gameName string) {
	if _, exists := stats[key]; !exists {
		stats[key] = []GameStats{}
	}

	for i, gameStat := range stats[key] {
		if gameStat.GameName == gameName {
			stats[key][i].LaunchCount++
			return
		}
	}

	stats[key] = append(stats[key], GameStats{GameName: gameName, LaunchCount: 1})
}

func updateNestedStats(stats YearStats, year, month, gameName string) {
	if _, exists := stats[year]; !exists {
		stats[year] = make(MonthStats)
	}

	if _, exists := stats[year][month]; !exists {
		stats[year][month] = []GameStats{}
	}

	for i, gameStat := range stats[year][month] {
		if gameStat.GameName == gameName {
			stats[year][month][i].LaunchCount++
			return
		}
	}

	stats[year][month] = append(stats[year][month], GameStats{GameName: gameName, LaunchCount: 1})
}

func getTop10Stats(stats *Stats, period string) Top10Stats {
	var top10 []GameStats

	switch {
	case period == "month" || isMonth(period):
		for _, gameStats := range stats.Month {
			top10 = append(top10, gameStats...)
		}
	case period == "year" || isYear(period):
		for _, gameStats := range stats.Year {
			top10 = append(top10, gameStats...)
		}
	case period == "all":
		for _, yearStats := range stats.All {
			for _, monthStats := range yearStats {
				top10 = append(top10, monthStats...)
			}
		}
	}

	sort.Slice(top10, func(i, j int) bool {
		return top10[i].LaunchCount > top10[j].LaunchCount
	})

	if len(top10) > 10 {
		top10 = top10[:10]
	}

	return Top10Stats{
		Period: period,
		Games:  top10,
	}
}

func isMonth(period string) bool {
	months := map[string]struct{}{
		"january": {}, "february": {}, "march": {}, "april": {}, "may": {}, "june": {}, "july": {}, "august": {}, "september": {}, "october": {}, "november": {}, "december": {},
	}
	_, exists := months[period]
	return exists
}

func isYear(period string) bool {
	_, err := time.Parse("2006", period)
	return err == nil
}
