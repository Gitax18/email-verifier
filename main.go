package main

import (
	"encoding/json"
	"fmt"
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
	log.Println("server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {

	enableCORS(&w)

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
			p := parsePolicy(rec) // retrieving Dmarc type
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

// parsePolicy parses a raw DMARC TXT record and returns the DMARC policy value.
//
// The function expects the full TXT record string for a DNS name like
// `_dmarc.example.com` (for example: "v=DMARC1; p=reject; rua=mailto:admin@example.com").
// It tokenizes the record on ';', trims whitespace, and searches for a tag
// named "p" (case-insensitive). When found, the policy value is lowercased
// and normalized to the known values "none", "quarantine", or "reject" when
// applicable. If the policy value is not one of the three common options the
// raw lowercased value is returned to preserve the tag's content.
//
// If no "p" tag is present, the function returns "none" as a safe default.
// Note: this is a lightweight parser intended for simple extraction of the
// policy tag only — it does not perform full DMARC syntax validation or parse
// other tags such as "sp", "pct" or reporting URIs.
//
// Examples:
//   parsePolicy("v=DMARC1; p=quarantine; rua=mailto:x@example.com") => "quarantine"
//   parsePolicy("v=dmarc1; P=REJECT") => "reject"
//   parsePolicy("v=dmarc1") => "none"
func parsePolicy(rec string) string {
	parts := strings.Split(rec, ";")
	for _, p := range parts {
		kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
		fmt.Println(kv)
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
