package servicebinding

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/converter"
	"github.com/redhat-developer/service-binding-operator/pkg/log"
)

// secret represents the data collected by this operator, and later handled as a secret.
type secret struct {
	logger *log.Log          // logger instance
	client dynamic.Interface // Kubernetes API client
	ns     string
	name   string
}

// buildResourceClient creates a resource client to handle corev1/secret resource.
func (s *secret) buildResourceClient() dynamic.ResourceInterface {
	gvr := corev1.SchemeGroupVersion.WithResource(secretResource)
	return s.client.Resource(gvr).Namespace(s.ns)
}

// compare existing secret and new secret
func (s *secret) isSame(secretName string, payload map[string][]byte) bool {
	logger := s.logger.WithValues("Namespace", s.ns, "Name", s.name)

	resourceClient := s.buildResourceClient()
	existingSecret, err := resourceClient.Get(secretName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "Error fetching secret")
		return false
	}
	existingSecretData, _, _ := unstructured.NestedMap(existingSecret.Object, "data")

	payloadInterim := make(map[string]interface{})
	for k, v := range payload {
		payloadInterim[k] = base64.StdEncoding.EncodeToString(v)
	}

	comparisonResult := nestedMapComparison(existingSecretData, payloadInterim)

	if comparisonResult.Success {
		logger.Debug("Secret data is same.")
		return true
	}

	return false
}

func hash(text string) string {
	algorithm := sha1.New()
	algorithm.Write([]byte(text))
	return hex.EncodeToString(algorithm.Sum(nil))
}

// create will take informed payload and create a new secret
// one. It can return error when Kubernetes client does.
func (s *secret) create(payload map[string][]byte, ownerReference metav1.OwnerReference) (*unstructured.Unstructured, error) {
	logger := s.logger.WithValues("Namespace", s.ns, "Name", s.name)

	secretObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       s.ns,
			Name:            s.name + "-" + hash(payload),
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: payload,
	}

	gvk := corev1.SchemeGroupVersion.WithKind(secretKind)
	u, err := converter.ToUnstructuredAsGVK(secretObj, gvk)
	if err != nil {
		return nil, err
	}

	resourceClient := s.buildResourceClient()

	logger.Debug("Attempt to create secret...")
	_, err = s.get()
	if err != nil {
		if errors.IsNotFound(err) {
			_, err := resourceClient.Create(u, metav1.CreateOptions{})
			if err != nil {
				logger.Error(err, "Error creating secret")
				return nil, err
			}
			logger.Info("Secret created")
			return u, nil
		}
		return nil, err
	}
	return u, nil
}

func (s *secret) delete(secretName string) error {
	resourceClient := s.buildResourceClient()
	err := resourceClient.Delete(secretName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil

}

// get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *secret) get() (*unstructured.Unstructured, error) {
	resourceClient := s.buildResourceClient()
	u, err := resourceClient.Get(s.name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// get an unstructured object from the secret handled by this component. It can return errors in case
// the API server does.
func (s *secret) get2(secretName string) (*unstructured.Unstructured, error) {
	resourceClient := s.buildResourceClient()
	u, err := resourceClient.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return u, nil
}

// newSecret instantiate a new Secret.
func newSecret(
	client dynamic.Interface,
	ns string,
	name string,
) *secret {
	return &secret{
		logger: log.NewLog("secret"),
		client: client,
		name:   name,
		ns:     ns,
	}
}
