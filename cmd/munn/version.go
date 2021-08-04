package main

var Version string
var BuildTime string

func init() {
	if Version == "" {
		Version = "(devel)"
	}

	rootCmd.Version = Version
	if rootCmd.Annotations == nil {
		rootCmd.Annotations = make(map[string]string)
	}
	rootCmd.Annotations["BuildTime"] = BuildTime
	rootCmd.SetVersionTemplate(`
		{{- with .Name}}{{printf "%s " .}}{{end}}
		{{- printf "version %s" .Version}}
		{{- with .Annotations.BuildTime}} (built {{.}}){{end}}
`)
}
