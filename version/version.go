package version

import "strings"

var (
	Version = "1.0.6"
	// PriorVersion - prior version
	PriorVersion = "1.0.5"
)

const (
	// LatestVersion product version supported
	LatestVersion        = "2.23.0"
	CompactLatestVersion = "2230"
	// LastMinorVersion product version supported
	LastMinorVersion = "2.20.0"

	LatestKubeImage = "quay.io/artemiscloud/activemq-artemis-broker-kubernetes:" + LatestVersion
	LatestInitImage = "quay.io/artemiscloud/activemq-artemis-broker-init:" + LatestVersion
)

func DefaultImageName(archSpecificRelatedImageEnvVarName string) string {
	if strings.Contains(archSpecificRelatedImageEnvVarName, "_Init_") {
		return LatestInitImage
	} else {
		return LatestKubeImage
	}
}

var CompactVersionFromVersion map[string]string = map[string]string{
	"2.15.0": "2150",
	"2.16.0": "2160",
	"2.18.0": "2180",
	"2.20.0": "2200",
	"2.21.0": "2210",
	"2.22.0": "2220",
	"2.23.0": "2230",
}

var FullVersionFromCompactVersion map[string]string = map[string]string{
	"2150": "2.15.0",
	"2160": "2.16.0",
	"2180": "2.18.0",
	"2200": "2.20.0",
	"2210": "2.21.0",
	"2220": "2.22.0",
	"2230": "2.23.0",
}

//The yacfg profile to use for a given full version of broker
var YacfgProfileVersionFromFullVersion map[string]string = map[string]string{
	"2.15.0": "2.15.0",
	"2.16.0": "2.16.0",
	"2.18.0": "2.18.0",
	"2.20.0": "2.18.0",
	"2.21.0": "2.21.0",
	"2.22.0": "2.21.0",
	"2.23.0": "2.21.0",
}

var YacfgProfileName string = "artemis"
