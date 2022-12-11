# Serving files from Azure Blob Storage

imgproxy can process images from Azure Blob Storage containers. To use this feature, do the following:

1. Set `IMGPROXY_USE_ABS` environment variable to `true`
2. Set `IMGPROXY_ABS_NAME` to your Azure account name and `IMGPROXY_ABS_KEY` to your Azure account key
4. _(optional)_ Specify the Azure Blob Storage endpoint with `IMGPROXY_ABS_ENDPOINT`
4. Use `abs://%bucket_name/%file_key` as the source image URL

### Leverage Azure Managed Identity or Service Principal

Microsoft encourages the use of a Managed Identity or Service Principal when accessing resources on an Azure Storage Account.
Both of these authentication pathways are supported out-of-the box.

#### Managed Identity

There is no additional configuration required so long as the resource that imgproxy is running on has a Managed Identity assigned to it.

#### Service Principal

Please refer to the [following documentation](https://learn.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal)
on creation of a service principal before proceeding.

Once that step is completed, the following environment variables must be configured depending on which option was chosen.

##### Secret Authentication

* `AZURE_CLIENT_ID`: Specifies the client ID for your application registration 
* `AZURE_TENANT_ID`: Specifies the tenant ID for your application registration 
* `AZURE_CLIENT_SECRET`: Specifies the client secret for your application registration.

##### Certificate Authentication

* `AZURE_CLIENT_ID`: Specifies the client ID for your application registration 
* `AZURE_TENANT_ID`: Specifies the tenant ID for your application registration 
* `AZURE_CLIENT_CERTIFICATE_PATH`: Specifies the path to a PFX or PEM-encoded certificate including private key 
* `AZURE_CLIENT_CERTIFICATE_PASSWORD`: (optional) the password protecting the certificate file (PFX (PKCS12))
* `AZURE_CLIENT_CERTIFICATE_CHAIN`: (optional) send certificate chain in x5c header to support subject name / issuer-based authentication
