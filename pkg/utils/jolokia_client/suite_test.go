package jolokia_client_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJolokia(t *testing.T) {
	RegisterFailHandler(Fail)
	prepareTmpEnv(t)
	RunSpecs(t, "Jolokia Suite")
}

func prepareTmpEnv(t *testing.T) {
	file, err := os.CreateTemp(os.TempDir(), "kubeconfig")
	Expect(err).NotTo(HaveOccurred())

	defer file.Close()

	file.WriteString("kind: Config\n")
	file.WriteString("apiVersion: v1\n")
	file.WriteString("clusters:\n")
	file.WriteString("- cluster:\n")
	file.WriteString("    server: https://example.com:6443\n")
	file.WriteString("  name: default\n")
	file.WriteString("contexts:\n")
	file.WriteString("- context:\n")
	file.WriteString("    cluster: default\n")
	file.WriteString("    namespace: default\n")
	file.WriteString("    user: admin\n")
	file.WriteString("users:\n")
	file.WriteString("- name: admin\n")
	file.WriteString("  user:\n")
	file.WriteString("    token: example-token\n")

	t.Setenv("KUBECONFIG", file.Name())
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
}
