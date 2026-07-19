package server

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
)

type PageData struct {
	Title              string
	Version            string
	Refresh            int
	Saved              bool
	GridColumns        string
	EnabledColumns     []string
	PortsWrapThreshold int
	ColumnWidths       string
	ColumnOrder        string
	Switches           []SwitchFormData
	CurrentLogLevel    string
}

type SwitchFormData struct {
	Index      int
	Name       string
	IP         string
	Username   string
	Password   string
	Model      string
	PortCount  int
	ParentIP   string
	ParentPort string
	UplinkPort string
	Enabled    bool
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"tojson": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "null"
			}
			return string(b)
		},
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
		"add": func(a, b int) int {
			return a + b
		},
	}
}

func loadTemplates() *template.Template {
	base := filepath.Dir("templates")
	if _, err := os.Stat(base); os.IsNotExist(err) {
		slog.Warn("templates directory not found, using embedded")
		return nil
	}
	tmpl := template.New("").Funcs(templateFuncs())
	return template.Must(tmpl.ParseGlob("templates/*.html"))
}
