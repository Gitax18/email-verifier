package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
)

type request struct {
	Email string `json:"email"`
}

type response struct {
	DmarcRecord string `json:"dmarcRecord"`
	IsDmarc     bool   `json:"isDmarc"`
	DmarcType   string `json:"dmarcType"`
}

func main() {
	http.HandleFunc("/verify", verifyHandler)
	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	domain := extractDomain(req.Email)
	if domain == "" {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	records, err := net.LookupTXT("_dmarc." + domain)
	if err != nil {
		// no TXT or lookup error => return empty dmarc
		writeJSON(w, response{DmarcRecord: "", IsDmarc: false, DmarcType: "none"})
		return
	}
	for _, rec := range records {
		if strings.HasPrefix(strings.ToLower(rec), "v=dmarc1") {
			p := parsePolicy(rec)
			writeJSON(w, response{DmarcRecord: rec, IsDmarc: true, DmarcType: p})
			return
		}
	}
	// no DMARC found in TXT records
	writeJSON(w, response{DmarcRecord: "", IsDmarc: false, DmarcType: "none"})
}

func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func parsePolicy(rec string) string {
	// rec like: "v=DMARC1; p=quarantine; rua=..."
	parts := strings.Split(rec, ";")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		if len(kv) == 2 && strings.ToLower(kv[0]) == "p" {
			val := strings.ToLower(strings.TrimSpace(kv[1]))
			switch val {
			case "none":
				return "none"
			case "quarantine":
				return "quarantine"
			case "reject":
				return "reject"
			default:
				return val
			}
		}
	}
	return "none"
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
