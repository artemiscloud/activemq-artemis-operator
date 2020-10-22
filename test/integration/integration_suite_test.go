package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCrAndCrds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CR and CRD Suite")
}
