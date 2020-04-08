package annotations

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/redhat-developer/service-binding-operator/pkg/controller/servicebindingrequest/bindinginfo"
)

const ConfigMapValue = "binding:env:object:configmap"

// IsConfigMap returns true if the annotation value should trigger config map handler.
func IsConfigMap(s string) bool {
	return ConfigMapValue == s
}

// NewConfigMapHandler constructs an annotation handler that can extract related data from config
// maps.
func NewConfigMapHandler(
	client dynamic.Interface,
	bindingInfo *bindinginfo.BindingInfo,
	resource unstructured.Unstructured,
) (Handler, error) {
	return NewResourceHandler(
		client,
		bindingInfo,
		resource,
		schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		},
	)
}
