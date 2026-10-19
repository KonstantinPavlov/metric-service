package config

import "fmt"

func PrintBuildInfo(buildVersion string, buildDate string, buildCommit string) {
	fmt.Printf("Build version: %s\n", getBuildValue(buildVersion))
	fmt.Printf("Build date: %s\n", getBuildValue(buildDate))
	fmt.Printf("Build commit: %s\n", getBuildValue(buildCommit))
}

func getBuildValue(val string) string {
	if val == "" {
		return "N/A"
	}
	return val
}
