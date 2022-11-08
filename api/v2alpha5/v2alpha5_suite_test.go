package v2alpha5_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV2alpha5(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2alpha5 Suite")
}
