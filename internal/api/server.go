package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"fmt"

	"fingerprint-service/internal/zkfp"
	"fingerprint-service/pkg/fingerprint"
)

// Server provides HTTP API for capture, verify, identify.
type Server struct {
	engine *zkfp.Engine
	svc    fingerprint.Service
}

// NewServer creates an API server. If engine is nil, a new engine is created and initialized.
func NewServer(engine *zkfp.Engine) (*Server, error) {
	if engine == nil {
		var err error
		engine, err = zkfp.NewEngine("")
		if err != nil {
			return nil, err
		}
		if _, err = engine.Init(); err != nil {
			engine.Close()
			return nil, err
		}
		_ = engine.SetFPEngineVersion("9")
	}
	svc := &serviceImpl{engine: engine}
	return &Server{engine: engine, svc: svc}, nil
}

// Close releases the engine.
func (s *Server) Close() error {
	if s.engine != nil {
		return s.engine.Close()
	}
	return nil
}

// Handler returns the HTTP handler (mux).
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/capture", s.handleCapture)
	mux.HandleFunc("/enroll", s.handleEnroll)
	mux.HandleFunc("/verify", s.handleVerify)
	mux.HandleFunc("/identify", s.handleIdentify)
	mux.HandleFunc("/health", s.handleHealth)
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCapture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	timeout := 30 * time.Second
	if t := r.URL.Query().Get("timeout_sec"); t != "" {
		if sec, err := strconv.Atoi(t); err == nil && sec > 0 {
			timeout = time.Duration(sec) * time.Second
		}
	}
	result, err := s.svc.Capture(timeout)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	presses := 3
	timeout := 30 * time.Second
	if p := r.URL.Query().Get("presses"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n >= 1 && n <= 5 {
			presses = n
		}
	}
	if t := r.URL.Query().Get("timeout_sec"); t != "" {
		if sec, err := strconv.Atoi(t); err == nil && sec > 0 {
			timeout = time.Duration(sec) * time.Second
		}
	}
	result, err := s.svc.Enroll(presses, timeout)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req fingerprint.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	fmt.Printf("DEBUG HASIL DECODE VERIFY: %+v\n", req)
	if req.RegisteredTemplate == "" || req.VerificationTemplate == "" {
		writeJSONError(w, http.StatusBadRequest, "registered_template and verification_template required")
		return
	}
	match, score, err := s.svc.Verify(req.RegisteredTemplate, req.VerificationTemplate, req.DoLearning)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(fingerprint.VerifyResult{Match: match, Score: score})
}

func (s *Server) handleIdentify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req fingerprint.IdentifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(req.Templates) == 0 || req.VerificationTemplate == "" {
		writeJSONError(w, http.StatusBadRequest, "templates and verification_template required")
		return
	}
	matchedID, score, processed, err := s.svc.Identify(req.Templates, req.VerificationTemplate)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(fingerprint.IdentifyResult{MatchedID: matchedID, Score: score, Processed: processed})
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}