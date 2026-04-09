package k8s

// IngressProfile defines the default annotations for an Ingress Controller
type IngressProfile struct {
	Name               string
	TLSAnnotations     map[string]string
	DefaultAnnotations map[string]string
}

var ingressProfiles = map[string]*IngressProfile{
	"nginx": {
		Name: "nginx",
		TLSAnnotations: map[string]string{
			"nginx.ingress.kubernetes.io/ssl-redirect":       "true",
			"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
		},
		DefaultAnnotations: map[string]string{},
	},
	"traefik": {
		Name: "traefik",
		TLSAnnotations: map[string]string{
			"traefik.ingress.kubernetes.io/router.tls":         "true",
			"traefik.ingress.kubernetes.io/router.entrypoints": "websecure",
		},
		DefaultAnnotations: map[string]string{},
	},
}

// GetIngressProfile returns the profile for a given controller name.
// Returns nil if the controller is unknown (caller should use empty annotations).
func GetIngressProfile(controllerName string) *IngressProfile {
	return ingressProfiles[controllerName]
}

// GetAnnotationsForController returns merged annotations for a given controller and SSL state.
// customAnnotations are user-defined and take highest priority.
func GetAnnotationsForController(controllerName string, sslEnabled bool, customAnnotations map[string]string) map[string]string {
	result := make(map[string]string)

	profile := GetIngressProfile(controllerName)
	if profile != nil {
		for k, v := range profile.DefaultAnnotations {
			result[k] = v
		}
		if sslEnabled {
			for k, v := range profile.TLSAnnotations {
				result[k] = v
			}
		}
	}

	// User-defined annotations override profile defaults
	for k, v := range customAnnotations {
		result[k] = v
	}

	return result
}
