package v1alpha2

import (
	"fmt"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.APIRule)

	*dst.Spec.Service.Host = fmt.Sprintf("%s%s", *dst.Spec.Service.Host, "a")
	return nil
}

func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.APIRule)
	*dst.Spec.Service.Host = (*src.Spec.Service.Host)[:len(*src.Spec.Service.Host)-1]
	return nil
}
