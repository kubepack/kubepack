package types

const ManifestFile = "manifest.yaml"

type ManifestDefinition struct {
	Package string `json:"package"`
	Owners []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"owners"`
	Dependencies []struct {
		Package string `json:"package"`
		Version string `json:"version,omitempty"`
		Branch  string `json:"branch,omitempty"`
		Folder  string `json:"folder,omitempty"`
		Repo    string `json:"repo,omitempty"`
	} `json:"dependencies"`
}
