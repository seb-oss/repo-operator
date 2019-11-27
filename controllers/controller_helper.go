package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// It returns a dockerconfig secret for service account.
func (r *RepositoryReconciler) generateRepoSecret(namespace string, secretName string, u string, p string) *v1.Secret {
	labels := map[string]string{
		"origin": "repo-operator",
		"type":   "dockercfg",
	}
	s := &v1.Secret{
		TypeMeta: v12.TypeMeta{
			Kind:       "Secret",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
	s.Name = secretName
	s.Namespace = namespace
	s.Labels = labels
	s.Type = v1.SecretTypeDockerConfigJson
	s.Data = map[string][]byte{}

	dockercfgJSONContent, _ := handleDockerCfgJSONContent(u, p, "unused", secretName)
	s.Data[v1.DockerConfigJsonKey] = dockercfgJSONContent
	return s
}

//Generate Docker config secret content
func handleDockerCfgJSONContent(username, password, email, server string) ([]byte, error) {
	dockercfgAuth := DockerConfigEntry{
		Email: email,
		Auth:  encodeDockerConfigFieldAuth(username, password),
	}

	dockerCfgJSON := DockerConfigJSON{
		Auths: map[string]DockerConfigEntry{server: dockercfgAuth},
	}

	return json.Marshal(dockerCfgJSON)
}

// Encode auth to base64
func encodeDockerConfigFieldAuth(username, password string) string {
	fieldValue := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(fieldValue))
}

//Link secret to service account as pull secret
func linkImagePullSecret(sa *v1.ServiceAccount, secret string) bool {
	exist := false
	for _, s := range sa.ImagePullSecrets {
		if s.Name == secret {
			exist = true
			break
		}
	}

	if !exist {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{
			Name: secret,
		})
		return true
	}

	return false
}

// Link secret
func linkSecret(sa *v1.ServiceAccount, secret string) bool {
	exist := false
	for _, s := range sa.Secrets {
		if s.Name == secret {
			exist = true
			break
		}
	}

	if !exist {
		sa.Secrets = append(sa.Secrets, v1.ObjectReference{Namespace: sa.Namespace, Name: secret})
		return true
	}

	return false
}

// UnLink the secret from builder SA
func unLinkBuilderSASecret(builderSA *v1.ServiceAccount, reqName string) *v1.ServiceAccount {
	secrets := []v1.ObjectReference{}
	for _, s := range builderSA.Secrets {
		if s.Name != reqName+suffixSecretName {
			secrets = append(secrets, s)
		}
	}
	builderSA.Secrets = secrets
	return builderSA
}

// Unlink the pull secret from default SA
func unLinkDefaultSAPullSecret(defaultSA *v1.ServiceAccount, reqName string) *v1.ServiceAccount {
	secrets := []v1.LocalObjectReference{}
	for _, s := range defaultSA.ImagePullSecrets {
		if s.Name != reqName+suffixSecretName {
			secrets = append(secrets, v1.LocalObjectReference{
				Name: s.Name,
			})
		}
	}
	defaultSA.ImagePullSecrets = secrets
	return defaultSA
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// newOwnerRef creates an OwnerReference pointing to the given owner.
func newOwnerRef(owner metav1.Object, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	blockOwnerDeletion := false
	isController := false
	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: &blockOwnerDeletion,
		Controller:         &isController,
	}
}

func setOwnerReference(owner, object metav1.Object, scheme *runtime.Scheme) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("is not a %T a runtime.Object, cannot call setOwnerReference", owner)
	}

	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}

	// Create a new ref
	ref := *newOwnerRef(owner, schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})

	existingRefs := object.GetOwnerReferences()
	fi := -1
	for i, r := range existingRefs {
		if referSameObject(ref, r) {
			fi = i
		}
	}
	if fi == -1 {
		existingRefs = append(existingRefs, ref)
	} else {
		existingRefs[fi] = ref
	}
	fmt.Printf("%+v", existingRefs)
	// Update owner references
	object.SetOwnerReferences(existingRefs)
	return nil
}

// Returns true if a and b point to the same object
func referSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}
