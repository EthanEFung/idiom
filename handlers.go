package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Response is the expected struct that will be sent to slack
type Response struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

func writeError(w http.ResponseWriter, err error) bool {
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	return err != nil
}

func handleIdiom(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	// parse x-www-urlencoded
	if err := r.ParseForm(); writeError(w, err) {
		return
	}
	// prep url parameters for xml response
	query := strings.TrimSpace(r.Form.Get("text"))
	id := os.Getenv("PHRASE_UID")
	token := os.Getenv("PHRASE_TOKEN")
	params := url.Values{}
	params.Add("uid", id)
	params.Add("tokenid", token)
	params.Add("format", "xml")
	params.Add("phrase", query)
	url := "https://www.stands4.com/services/v2/phrases.php?" + params.Encode()

	// parse get and parse phrases response
	res, err := http.Get(url)
	if writeError(w, err) {
		return
	}
	type result struct {
		XMLName     xml.Name `xml:"result"`
		Term        string   `xml:"term,string"`
		Explanation string   `xml:"explanation,string"`
		Example     string   `xml:"example,string"`
	}
	type results struct {
		XMLName xml.Name `xml:"results"`
		Results []result `xml:"result"`
	}
	var m results

	d := xml.NewDecoder(res.Body)
	if err := d.Decode(&m); writeError(w, err) {
		return
	}
	if len(m.Results) == 0 {
		w.Write([]byte("No results"))
		return
	}

	var t string
	for _, result := range m.Results {
		t += ">_" + strings.TrimSpace(result.Term) + "_\n"
		t += "*Explanation:* " + strings.TrimSpace(result.Explanation) + "\n"
		t += "*Example:* \"" + strings.TrimSpace(result.Example) + "\"\n\n"
	}

	// write response
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(Response{ResponseType: "in_channel", Text: t})
	if writeError(w, err) {
		return
	}
	w.Write(b)
}
