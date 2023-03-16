package e2e_test

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestFixture struct {
	Cli    client.Client
	K8SCli *kubernetes.Clientset
	NS     *corev1.Namespace
}

var fixture TestFixture

func TestE2e(t *testing.T) {
	BeforeSuite(func() {
		Expect(initClient()).ToNot(HaveOccurred())
		Expect(initK8SClient()).ToNot(HaveOccurred())
		Expect(createNamespace("mixedcpus-testing-")).ToNot(HaveOccurred())
	})
	AfterSuite(func() {
		Expect(fixture.Cli.Delete(context.TODO(), fixture.NS)).ToNot(HaveOccurred())
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")

}

func initClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	fixture.Cli, err = client.New(cfg, client.Options{})
	return err
}

func initK8SClient() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	fixture.K8SCli, err = kubernetes.NewForConfig(cfg)
	return err
}

func createNamespace(prefix string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: prefix,
			Labels: map[string]string{
				"security.openshift.io/scc.podSecurityLabelSync": "false",
				"pod-security.kubernetes.io/audit":               "privileged",
				"pod-security.kubernetes.io/enforce":             "privileged",
				"pod-security.kubernetes.io/warn":                "privileged",
			},
		},
	}
	if err := fixture.Cli.Create(context.TODO(), ns); err != nil {
		return err
	}
	fixture.NS = ns
	return nil
}

// TODO make it possible to read directly from DaemonSet
func GetSharedCPUs() string {
	cpus, ok := os.LookupEnv("E2E_SHARED_CPUS")
	if !ok {
		return ""
	}
	return cpus
}
