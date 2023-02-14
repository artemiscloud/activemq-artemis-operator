package version

import (
	"strings"

	"github.com/blang/semver/v4"
)

var (
	Version = "1.0.10"

	// The build timestamp injected at build-time
	BuildTimestamp = ""
)

const (
	// Latest ActiveMQ Artemis version
	LatestActiveMQArtemisVersion = "2.28.0"

	// Latest ActiveMQ Artemis Init image
	LatestActiveMQArtemisInitImage = "quay.io/artemiscloud/activemq-artemis-broker-init:artemis." + LatestActiveMQArtemisVersion

	// Latest ActiveMQ Artemis Kubernetes image
	LatestActiveMQArtemisKubeImage = "quay.io/artemiscloud/activemq-artemis-broker-kubernetes:artemis." + LatestActiveMQArtemisVersion
)

func DefaultImageName(archSpecificRelatedImageEnvVarName string) string {
	if strings.Contains(archSpecificRelatedImageEnvVarName, "_Init_") {
		return LatestActiveMQArtemisInitImage
	} else {
		return LatestActiveMQArtemisKubeImage
	}
}

func CompactActiveMQArtemisVersion(version string) string {
	return strings.Replace(version, ".", "", -1)
}

var supportedActiveMQArtemisSemanticVersions []semver.Version

func SupportedActiveMQArtemisSemanticVersions() []semver.Version {
	if supportedActiveMQArtemisSemanticVersions == nil {
		supportedActiveMQArtemisSemanticVersions = make([]semver.Version, len(SupportedActiveMQArtemisVersions))
		for i := 0; i < len(SupportedActiveMQArtemisVersions); i++ {
			supportedActiveMQArtemisSemanticVersions[i] = semver.MustParse(SupportedActiveMQArtemisVersions[i])
		}
		semver.Sort(supportedActiveMQArtemisSemanticVersions)
	}

	return supportedActiveMQArtemisSemanticVersions
}

// Sorted arry of supported ActiveMQ Artemis versions
var SupportedActiveMQArtemisVersions = []string{
	"2.15.0",
	"2.16.0",
	"2.18.0",
	"2.20.0",
	"2.21.0",
	"2.22.0",
	"2.23.0",
	"2.25.0",
	"2.26.0",
	"2.27.0",
	"2.28.0",
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
	"2.25.0": "2.21.0",
	"2.26.0": "2.21.0",
	"2.27.0": "2.21.0",
	"2.28.0": "2.21.0",
}

var YacfgProfileName string = "artemis"
