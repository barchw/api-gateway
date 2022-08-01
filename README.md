# Migration to new API Rule version guide

## Implementing new version and conversion webhook

1. Create new APIRule **v1beta1** version with:

    ```sh
    kubebuilder create api --group gateway  --version v1beta1 --kind APIRule
    ```

2. Introduce api changes

3. Make sure that **v1beta1** api version is set as storage type

   ```go
   // In file api/v1beta1/apiRule_types.go

   // +kubebuilder:storageversion
   type APIRule struct {
   ```

4. Implement [hub](./api/v1beta1/apirule_conversion.go) and [spoke](./api/v1alpha1/apirule_conversion.go) that implements
`func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error` and `func (dst *APIRule) ConvertFrom(srcRaw conversion.Hub) error`

   > Sidenote: If breaking changes are introduced in new api version, the `ConvertFrom` should throw an error if the conversion is not possible. However the conversion MUST be done if possible. If the `ConvertFrom` would just throw an error, creation of the resource in older version will fail (even though we are implementing NEW -> OLD conversion)

   > Other than that, ConvertFrom will not block any feature creation

1. Create conversion webhook

   ```sh
   kubebuilder create webhook --group gateway --version v1beta1 --kind APIRule --conversion
   ```

2. Set up conversion versions in [config/crd/patches/webhook_in_apirules.yaml](https://github.tools.sap/xf-goat/api-rule-controller-conversion/blob/d52bfe6f35bca8114ffd595447b4316dff12f90d/config/crd/patches/webhook_in_apirules.yaml#L15)

3. Enable "patches/webhook_in_apirules.yaml" and "patches/cainjection_in_apirules.yaml" in [config/crd/kustomization.yaml](./config/crd/kustomization.yaml)

4. Uncomment lines under [WEBHOOK], "../certmanager" and vars at the end of file in [config/default/kustomization.yaml](./config/default/kustomization.yaml)

5. Comment out "- manifests.yaml" in [config/webhook/kustomization.yaml](config/webhook/kustomization.yaml)

6.  ```make generate manifests install```

7.  Build new controller version and deploy (make sure that certs are injected, I used [cert-manager](https://github.com/cert-manager/cert-manager/releases/download/v1.9.0/cert-manager.yaml))

8.  Conversion should now be in place, the change should be visible with

    ```sh
    kubectl describe apirules.v1beta1.gateway.kyma-project.io
    ```

    with `Host` changed to `Host2`
