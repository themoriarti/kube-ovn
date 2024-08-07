package framework

import (
	"context"

	apiv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	v1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/client/clientset/versioned/typed/k8s.cni.cncf.io/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo/v2"
)

// NetworkAttachmentDefinitionClient is a struct for nad client.
type NetworkAttachmentDefinitionClient struct {
	f *Framework
	v1.NetworkAttachmentDefinitionInterface
}

func (f *Framework) NetworkAttachmentDefinitionClient() *NetworkAttachmentDefinitionClient {
	return f.NetworkAttachmentDefinitionClientNS(f.Namespace.Name)
}

func (f *Framework) NetworkAttachmentDefinitionClientNS(namespace string) *NetworkAttachmentDefinitionClient {
	return &NetworkAttachmentDefinitionClient{
		f:                                    f,
		NetworkAttachmentDefinitionInterface: f.AttachNetClient.K8sCniCncfIoV1().NetworkAttachmentDefinitions(namespace),
	}
}

func (c *NetworkAttachmentDefinitionClient) Get(name string) *apiv1.NetworkAttachmentDefinition {
	ginkgo.GinkgoHelper()
	nad, err := c.NetworkAttachmentDefinitionInterface.Get(context.TODO(), name, metav1.GetOptions{})
	ExpectNoError(err)
	return nad
}

// Create creates a new nad according to the framework specifications
func (c *NetworkAttachmentDefinitionClient) Create(nad *apiv1.NetworkAttachmentDefinition) *apiv1.NetworkAttachmentDefinition {
	ginkgo.GinkgoHelper()
	nad, err := c.NetworkAttachmentDefinitionInterface.Create(context.TODO(), nad, metav1.CreateOptions{})
	ExpectNoError(err, "Error creating nad")
	return c.Get(nad.Name)
}

// Delete deletes a nad if the nad exists
func (c *NetworkAttachmentDefinitionClient) Delete(name string) {
	ginkgo.GinkgoHelper()
	err := c.NetworkAttachmentDefinitionInterface.Delete(context.TODO(), name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return
	}
	ExpectNoError(err, "Error deleting nad")
}

func MakeNetworkAttachmentDefinition(name, namespace, conf string) *apiv1.NetworkAttachmentDefinition {
	nad := &apiv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiv1.NetworkAttachmentDefinitionSpec{
			Config: conf,
		},
	}
	return nad
}
