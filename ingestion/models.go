package ingestion

import "time"

type SDKEvent struct {
	Contexts struct {
		Device struct {
			Arch   string `json:"arch"`
			NumCpu int    `json:"num_cpu"`
		} `json:"device"`
		Os struct {
			Name string `json:"name"`
		} `json:"os"`
		Runtime struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"runtime"`
	} `json:"contexts"`
	EventId     string   `json:"event_id"`
	Fingerprint []string `json:"fingerprint"`
	Level       string   `json:"level"`
	Platform    string   `json:"platform"`
	Sdk         struct {
		Name         string   `json:"name"`
		Version      string   `json:"version"`
		Integrations []string `json:"integrations"`
		Packages     []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
	} `json:"sdk"`
	User struct {
	} `json:"user"`
	// Name -> Version
	Modules   map[string]string `json:"modules"`
	Exception []struct {
		Type       string `json:"type"`
		Value      string `json:"value"`
		Stacktrace struct {
			Frames []struct {
				Function string `json:"function"`
				Module   string `json:"module"`
				AbsPath  string `json:"abs_path"`
				Lineno   int    `json:"lineno"`
			} `json:"frames"`
		} `json:"stacktrace"`
	} `json:"exception"`
	Timestamp time.Time `json:"timestamp"`
}

type ProjectEventMessage struct {
	ProjectID uint  `json:"project_id"`
	Event     Event `json:"event"`
}

type Event SDKEvent
