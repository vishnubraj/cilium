// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package endpoint

import (
	"k8s.io/apimachinery/pkg/types"

	"github.com/cilium/cilium/pkg/endpoint/id"
)

// GetContainerName returns the name of the container for the endpoint.
func (e *Endpoint) GetContainerName() string {
	e.unconditionalRLock()
	defer e.runlock()
	return e.containerName
}

// GetK8sPodName returns the name of the pod if the endpoint represents a
// Kubernetes pod
func (e *Endpoint) GetK8sPodName() string {
	// const after creation
	k8sPodName := e.K8sPodName

	return k8sPodName
}

// HumanString returns the endpoint's most human readable identifier as string
func (e *Endpoint) HumanString() string {
	if cep := e.GetK8sNamespaceAndCEPName(); cep != "" {
		return cep
	}

	return e.StringID()
}

// GetK8sNamespaceAndPodName returns the corresponding namespace and pod
// name for this endpoint.
func (e *Endpoint) GetK8sNamespaceAndPodName() string {
	// both fields are const after creation
	return e.K8sNamespace + "/" + e.K8sPodName
}

// GetK8sCEPName returns the corresponding K8s CiliumEndpoint resource name
// for this endpoint (without the namespace)
// Returns an empty string if the endpoint does not belong to a pod.
func (e *Endpoint) GetK8sCEPName() string {
	// all fields are const after creation

	// Endpoints which have not opted out of legacy identifiers will continue
	// to use just the pod name as the cep name for backwards compatibility reasons.
	if e.disableLegacyIdentifiers && e.K8sPodName != "" && e.containerIfName != "" {
		return e.K8sPodName + "-" + e.containerIfName
	}
	return e.K8sPodName
}

// GetK8sNamespaceAndCEPName returns the corresponding namespace and
// K8s CiliumEndpoint resource name for this endpoint.
func (e *Endpoint) GetK8sNamespaceAndCEPName() string {
	// all fields are const after creation
	return e.K8sNamespace + "/" + e.GetK8sCEPName()
}

// getCNIAttachmentIDLocked returns the endpoint's unique CNI attachment ID
func (e *Endpoint) getCNIAttachmentIDLocked() string {
	if e.containerIfName != "" {
		return e.containerID + ":" + e.containerIfName
	}
	return e.containerID
}

// GetCNIAttachmentID returns the endpoint's unique CNI attachment ID
func (e *Endpoint) GetCNIAttachmentID() string {
	e.unconditionalRLock()
	defer e.runlock()
	return e.getCNIAttachmentIDLocked()
}

// GetContainerID returns the endpoint's container ID
func (e *Endpoint) GetContainerID() string {
	e.unconditionalRLock()
	cID := e.containerID
	e.runlock()
	return cID
}

// GetShortContainerID returns the endpoint's shortened container ID
func (e *Endpoint) GetShortContainerID() string {
	e.unconditionalRLock()
	defer e.runlock()
	return e.getShortContainerIDLocked()
}

func (e *Endpoint) getShortContainerIDLocked() string {
	if e == nil {
		return ""
	}

	cid := e.containerID

	caplen := 10
	if len(cid) <= caplen {
		return cid
	}

	return cid[:caplen]

}

func (e *Endpoint) GetDockerEndpointID() string {
	// const after creation
	return e.dockerEndpointID
}

// IdentifiersLocked fetches the set of attributes that uniquely identify the
// endpoint. The caller must hold exclusive control over the endpoint.
func (e *Endpoint) IdentifiersLocked() id.Identifiers {
	refs := make(id.Identifiers, 8)
	if cniID := e.getCNIAttachmentIDLocked(); cniID != "" {
		refs[id.CNIAttachmentIdPrefix] = cniID
	}

	if !e.disableLegacyIdentifiers && e.containerID != "" {
		refs[id.ContainerIdPrefix] = e.containerID
	}

	if e.dockerEndpointID != "" {
		refs[id.DockerEndpointPrefix] = e.dockerEndpointID
	}

	if e.IPv4.IsValid() {
		refs[id.IPv4Prefix] = e.IPv4.String()
	}

	if e.IPv6.IsValid() {
		refs[id.IPv6Prefix] = e.IPv6.String()
	}

	if !e.disableLegacyIdentifiers && e.containerName != "" {
		refs[id.ContainerNamePrefix] = e.containerName
	}

	if podName := e.GetK8sNamespaceAndPodName(); !e.disableLegacyIdentifiers && podName != "" {
		refs[id.PodNamePrefix] = podName
	}

	if cepName := e.GetK8sNamespaceAndCEPName(); cepName != "" {
		refs[id.CEPNamePrefix] = cepName
	}

	return refs
}

// Identifiers fetches the set of attributes that uniquely identify the endpoint.
func (e *Endpoint) Identifiers() (id.Identifiers, error) {
	if err := e.rlockAlive(); err != nil {
		return nil, err
	}
	defer e.runlock()

	return e.IdentifiersLocked(), nil
}

// GetCiliumEndpointUID returns the UID of the CiliumEndpoint.
func (e *Endpoint) GetCiliumEndpointUID() types.UID {
	e.unconditionalRLock()
	defer e.runlock()
	return e.ciliumEndpointUID
}

// SetCiliumEndpointUID modifies the endpoint's CiliumEndpoint UID.
func (e *Endpoint) SetCiliumEndpointUID(uid types.UID) {
	e.unconditionalLock()
	e.ciliumEndpointUID = uid
	e.unlock()
}
