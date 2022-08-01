package v1alpha1

import (
	"encoding/json"
	"log"

	"github.com/kyma-incubator/api-gateway/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this CronJob to the Hub version (v1beta1).
func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	log.Default().Printf("dst host: %s", src.Name)
	dst := dstRaw.(*v1beta1.APIRule)

	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	json.Unmarshal(specData, &dst.Spec)

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	json.Unmarshal(statusData, &dst.Status)

	dst.ObjectMeta = src.ObjectMeta

	host := *src.Spec.Service.Host
	dst.Spec.Host = &host

	return nil
}

// ConvertFrom converts this CronJob from the Hub version (v1beta1).
func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.APIRule)
	specData, err := json.Marshal(src.Spec)
	if err != nil {
		return err
	}

	json.Unmarshal(specData, &dst.Spec)

	statusData, err := json.Marshal(src.Status)
	if err != nil {
		return err
	}

	json.Unmarshal(statusData, &dst.Status)

	dst.ObjectMeta = src.ObjectMeta

	if src.Spec.Service == nil {
		log.Default().Print("conversion from v1beta1 to v1alpha1 wasn't possible as service isn't set on spec level")
		return nil
	}

	for _, rule := range src.Spec.Rules {
		if rule.Service != nil {
			log.Default().Print("conversion from v1beta1 to v1alpha1 isn't possible with rule level service definition")
			return nil
		}
	}

	host := *src.Spec.Host
	/*service := Service{}
	service.Host = &host
	if service.IsExternal != nil {
		isExternal := *src.Spec.Service.IsExternal
		service.IsExternal = &isExternal
	}
	port := *src.Spec.Service.Port
	service.Port = &port
	name := *src.Spec.Service.Name
	service.Name = &name
	dst.Spec.Service = &service*/
	dst.Spec.Service.Host = &host

	return nil
}
