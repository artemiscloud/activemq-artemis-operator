package v2alpha2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestV2alpha2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2alpha2 Suite")
}
