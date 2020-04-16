package lib

type OS string

const (
	Linux   OS = "linux"
	Windows OS = "windows"
	MacOS   OS = "darwin"
)

type ScriptRef struct {
	OS      OS     `json:"os"`
	URL     string `json:"url"`
	Command string `json:"command"`
}
