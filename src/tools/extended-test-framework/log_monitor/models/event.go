package models

type Event struct {
	SourceClass    string   `json:"sourceClass"`
	SourceInstance string   `json:"sourceInstance"`
	Resource       string   `json:"resource"`
	Class          string   `json:"class"`
	Message        string   `json:"message"`
	Data           []string `json:"data"`
	Items          string   `json:"items"`
}
