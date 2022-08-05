package webhook

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CertFile       = "server-cert.pem"
	KeyFile        = "server-key.pem"
	DefaultCertDir = "/tmp/k8s-webhook-server/serving-certs"
	APIRuleCRDName = "apirules.gateway.kyma-project.io"
)

func SetupCertificates(ctx context.Context, webhookNamespace, serviceName string) error {
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}
	if err := apiextensionsv1.AddToScheme(serverClient.Scheme()); err != nil {
		return errors.Wrap(err, "while adding apiextensions.v1 schema to k8s client")
	}

	key, _, err := CreateCABundle(ctx, webhookNamespace, DefaultCertDir, serviceName)
	if err != nil {
		return errors.Wrap(err, "failed to ensure webhook secret")
	}

	if err := AddCertToConversionWebhook(ctx, serverClient, key); err != nil {
		return errors.Wrap(err, "while adding CaBundle to Conversion Webhook for function CRD")
	}
	return nil
}

func CreateCABundle(ctx context.Context, webhookNamespace string, certDir string, serviceName string) ([]byte, []byte, error) {
	logger := ctrl.LoggerFrom(ctx)
	logger.Info("creating certificate for webhook")

	key, cert, err := createCert(ctx, certDir, webhookNamespace, serviceName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to crete cert")
	}
	return key, cert, nil
}

func AddCertToConversionWebhook(ctx context.Context, client ctrlclient.Client, caBundle []byte) error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := client.Get(ctx, types.NamespacedName{Name: APIRuleCRDName}, crd)
	if err != nil {
		return errors.Wrap(err, "failed to get function crd")
	}

	if contains, msg := containsConversionWebhookClientConfig(crd); !contains {
		return errors.Errorf("while validating CRD to be CaBundle injectable,: %s", msg)
	}

	crd.Spec.Conversion.Webhook.ClientConfig.CABundle = caBundle
	err = client.Update(ctx, crd)
	if err != nil {
		return errors.Wrap(err, "while updating CRD with Conversion webhook caBundle")
	}
	return nil
}

func containsConversionWebhookClientConfig(crd *apiextensionsv1.CustomResourceDefinition) (bool, string) {
	if crd.Spec.Conversion == nil {
		return false, "conversion not found in function CRD"
	}

	if crd.Spec.Conversion.Webhook == nil {
		return false, "conversion webhook not found in function CRD"
	}

	if crd.Spec.Conversion.Webhook.ClientConfig == nil {
		return false, "client config for conversion webhook not found in function CRD"
	}
	return true, ""
}

func createCert(ctx context.Context, certDir string, webhookNamespace string, serviceName string) ([]byte, []byte, error) {

	cert, key, err := buildCert(webhookNamespace, serviceName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build certificate")
	}
	keyPath := path.Join(certDir, "tls.key")
	os.WriteFile(keyPath, key, 0644)

	certPath := path.Join(certDir, "tls.crt")
	os.WriteFile(certPath, cert, 0644)

	return cert, key, nil
}

func isValidSecret(s *corev1.Secret) (bool, error) {
	if !hasRequiredKeys(s.Data) {
		return false, nil
	}
	if err := verifyCertificate(s.Data[CertFile]); err != nil {
		return false, err
	}
	if err := verifyKey(s.Data[KeyFile]); err != nil {
		return false, err
	}

	return true, nil
}

func verifyCertificate(c []byte) error {
	certificate, err := cert.ParseCertsPEM(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse certificate data")
	}
	// certificate is self signed. So we use it as a root cert
	root, err := cert.NewPoolFromBytes(c)
	if err != nil {
		return errors.Wrap(err, "failed to parse root certificate data")
	}
	// make sure the certificate is valid for the next 10 days. Otherwise it will be recreated.
	_, err = certificate[0].Verify(x509.VerifyOptions{CurrentTime: time.Now().Add(10 * 24 * time.Hour), Roots: root})
	if err != nil {
		return errors.Wrap(err, "certificate verification failed")
	}
	return nil
}

func verifyKey(k []byte) error {
	b, _ := pem.Decode(k)
	key, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		return errors.Wrap(err, "failed to parse key data")
	}
	if err = key.Validate(); err != nil {
		return errors.Wrap(err, "key verification failed")
	}
	return nil
}

func hasRequiredKeys(data map[string][]byte) bool {
	if data == nil {
		return false
	}
	for _, key := range []string{CertFile, KeyFile} {
		if _, ok := data[key]; !ok {
			return false
		}
	}
	return true
}

func buildCert(namespace, serviceName string) (cert []byte, key []byte, err error) {
	cert, key, err = generateWebhookCertificates(serviceName, namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate webhook certificates")
	}

	return cert, key, nil
}

func generateWebhookCertificates(serviceName, namespace string) ([]byte, []byte, error) {
	altNames := serviceAltNames(serviceName, namespace)
	return cert.GenerateSelfSignedCertKey(altNames[0], nil, altNames)
}

func serviceAltNames(serviceName, namespace string) []string {
	namespacedServiceName := strings.Join([]string{serviceName, namespace}, ".")
	commonName := strings.Join([]string{namespacedServiceName, "svc"}, ".")
	serviceHostname := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)

	return []string{
		commonName,
		serviceName,
		namespacedServiceName,
		serviceHostname,
	}
}
