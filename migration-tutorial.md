# Migration to new API Rule version guide

## As an example version migration API Rule status "OK" will be changed to "GOOD"

1. Create new APIRule version with:

    ```sh
    kubebuilder create api --group gateway  --version v1alpha2 --kind APIRule
    ```

2. Make sure that v1alpha1 api version is set as storage type

   ```go
   // In file api/v1alpha1/apiRule_types.go

   // +kubebuilder:storageversion
   type APIRule struct {
   ```

3. Implement [hub](api/v1alpha1/apirule_conversions.go) and [spoke](api/v1alpha2/apirule_conversion.go) that implements
`func (src *APIRule) ConvertTo(dstRaw conversion.Hub) error` and `func (src *APIRule) ConvertFrom(srcRaw conversion.Hub) error`

4. Update v1alpha1 definitition

   ```sh
   kubebuilder create api --group gateway  --version v1alpha1 --kind APIRule
   ```

5. Create conversion webhook

   ```sh
   kubebuilder create webhook --group gateway --version v1alpha1 --kind APIRule --conversion
   ```
