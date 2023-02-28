# Loading environment variables

imgproxy can load environment variables from various sources such as:

* [Local file](#local-file)
* [AWS Secrets Manager](#aws-secrets-manager)
* [AWS Systems Manager Parameter Store](#aws-systems-manager-parameter-store)
* [Google Cloud Secret Manager](#google-cloud-secret-manager)

## Local file

You can create an [environment file](#environment-file-syntax) and configure imgproxy to read environment variables from it.

* `IMGPROXY_ENV_LOCAL_FILE_PATH`: the path of the environmebt file to load

## AWS Secrets Manager

You can store the content of an [environment file](#environment-file-syntax) as an AWS Secrets Manager secret and configure imgproxy to read environment variables from it.

* `IMGPROXY_ENV_AWS_SECRET_ID`: the ARN or name of the secret to load
* `IMGPROXY_ENV_AWS_SECRET_VERSION_ID`: _(optional)_ the unique identifier of the version of the secret to load
* `IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE`: _(optional)_ the staging label of the version of the secret to load
* `IMGPROXY_ENV_AWS_SECRET_REGION`: _(optional)_ the region of the secret to load

**📝 Note:** If both `IMGPROXY_ENV_AWS_SECRET_VERSION_ID` and `IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE` are set, `IMGPROXY_ENV_AWS_SECRET_VERSION_STAGE` will be ignored

### Set up AWS Secrets Manager credentials

There are three ways to specify your AWS credentials. The credentials policy should allow performing the `secretsmanager:GetSecretValue` and `secretsmanager:ListSecretVersionIds` actions with the specified secret:

#### IAM Roles

If you're running imgproxy on an Amazon Web Services platform, you can use IAM roles to to get the security credentials to retrieve the secret.

* **Elastic Container Service (ECS):** Assign an [IAM role to a task](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html).
* **Elastic Kubernetes Service (EKS):** Assign a [service account to a pod](https://docs.aws.amazon.com/eks/latest/userguide/pod-configuration.html).
* **Elastic Beanstalk:** Assign an [IAM role to an instance](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/iam-instanceprofile.html).

#### Environment variables

You can specify an AWS Access Key ID and a Secret Access Key by setting the standard `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

``` bash
AWS_ACCESS_KEY_ID=my_access_key AWS_SECRET_ACCESS_KEY=my_secret_key imgproxy

# same for Docker
docker run -e AWS_ACCESS_KEY_ID=my_access_key -e AWS_SECRET_ACCESS_KEY=my_secret_key -it darthsim/imgproxy
```

#### Shared credentials file

Alternatively, you can create the `.aws/credentials` file in your home directory with the following content:

```ini
[default]
aws_access_key_id = %access_key_id
aws_secret_access_key = %secret_access_key
```

## AWS Systems Manager Parameter Store

You can store multiple AWS Systems Manager Parameter Store entries and configure imgproxy to load their values to separate environment variables.

* `IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH`: the [path](#aws-systems-manager-path) of the parameters to load
* `IMGPROXY_ENV_AWS_SSM_PARAMETERS_REGION`: _(optional)_ the region of the parameters to load

### AWS Systems Manager path

Let's assume that you created the following AWS Systems Manager parameters:

* `/imgproxy/prod/IMGPROXY_KEY`
* `/imgproxy/prod/IMGPROXY_SALT`
* `/imgproxy/prod/IMGPROXY_CLOUD_WATCH/SERVICE_NAME`
* `/imgproxy/prod/IMGPROXY_CLOUD_WATCH/NAMESPACE`
* `/imgproxy/staging/IMGPROXY_KEY`

If you set `IMGPROXY_ENV_AWS_SSM_PARAMETERS_PATH` to `/imgproxy/prod`, imgproxy will load these parameters the following way:

* `/imgproxy/prod/IMGPROXY_KEY` value will be loaded to `IMGPROXY_KEY`
* `/imgproxy/prod/IMGPROXY_SALT` value will be loaded to `IMGPROXY_SALT`
* `/imgproxy/prod/IMGPROXY_CLOUD_WATCH/SERVICE_NAME` value will be loaded to `IMGPROXY_CLOUD_WATCH_SERVICE_NAME`
* `/imgproxy/prod/IMGPROXY_CLOUD_WATCH/NAMESPACE` value will be loaded to `IMGPROXY_CLOUD_WATCH_NAMESPACE`
* `/imgproxy/staging/IMGPROXY_KEY` will be ignored since its path is not `/imgproxy/prod`

### Set up AWS Systems Manager credentials

There are three ways to specify your AWS credentials. The credentials policy should allow performing the `ssm:GetParametersByPath` action with the specified parameters:

#### IAM Roles

If you're running imgproxy on an Amazon Web Services platform, you can use IAM roles to to get the security credentials to retrieve the secret.

* **Elastic Container Service (ECS):** Assign an [IAM role to a task](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html).
* **Elastic Kubernetes Service (EKS):** Assign a [service account to a pod](https://docs.aws.amazon.com/eks/latest/userguide/pod-configuration.html).
* **Elastic Beanstalk:** Assign an [IAM role to an instance](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/iam-instanceprofile.html).

#### Environment variables

You can specify an AWS Access Key ID and a Secret Access Key by setting the standard `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

``` bash
AWS_ACCESS_KEY_ID=my_access_key AWS_SECRET_ACCESS_KEY=my_secret_key imgproxy

# same for Docker
docker run -e AWS_ACCESS_KEY_ID=my_access_key -e AWS_SECRET_ACCESS_KEY=my_secret_key -it darthsim/imgproxy
```

#### Shared credentials file

Alternatively, you can create the `.aws/credentials` file in your home directory with the following content:

```ini
[default]
aws_access_key_id = %access_key_id
aws_secret_access_key = %secret_access_key
```

## Google Cloud Secret Manager

You can store the content of an [environment file](#environment-file-syntax) in Google Cloud Secret Manager secret and configure imgproxy to read environment variables from it.

* `IMGPROXY_ENV_GCP_SECRET_ID`: the name of the secret to load
* `IMGPROXY_ENV_GCP_SECRET_VERSION_ID`: _(optional)_ the unique identifier of the version of the secret to load
* `IMGPROXY_ENV_GCP_SECRET_PROJECT_ID`: the name or ID of the Google Cloud project that contains the secret

### Setup credentials

If you run imgproxy inside Google Cloud infrastructure (Compute Engine, Kubernetes Engine, App Engine, Cloud Functions, etc), and you have granted access to the specified secret to the service account, you probably don't need to do anything here. imgproxy will try to use the credentials provided by Google.

Otherwise, set `IMGPROXY_ENV_GCP_KEY` environment variable to the content of Google Cloud JSON key. Get more info about JSON keys: [https://cloud.google.com/iam/docs/creating-managing-service-account-keys](https://cloud.google.com/iam/docs/creating-managing-service-account-keys).

## Environment file syntax

The following syntax rules apply to environment files:

* Blank lines are ignored
* Lines beginning with `#` are processed as comments and ignored
* Each line represents a key-value pair. Values can optionally be quoted:
  * `VAR=VAL` -> `VAL`
  * `VAR="VAL"` -> `VAL`
  * `VAR='VAL'` -> `VAL`
* Unquoted and double-quoted (`"`) values have variable substitution applied:
  * `VAR=${OTHER_VAR}` -> value of `OTHER_VAR`
  * `VAR=$OTHER_VAR` -> value of `OTHER_VAR`
  * `VAR="$OTHER_VAR"` -> value of `OTHER_VAR`
  * `VAR="${OTHER_VAR}"` -> value of `OTHER_VAR`
* Single-quoted (`'`) values are used literally:
  * `VAR='$OTHER_VAR'` -> `$OTHER_VAR`
  * `VAR='${OTHER_VAR}'` -> `${OTHER_VAR}`
* Double quotes in double-quoted (`"`) values can be escaped with `\`:
  * `VAR="{\"hello\": \"json\"}"` -> `{"hello": "json"}`
* Slash (`\`) in double-quoted values can be escaped with another slash:
  * `VAR="some\\value"` -> `some\value`
* A new line can be added to double-quoted values using `\n`:
  * `VAR="some\nvalue"` ->
    ```
    some
    value
    ```

