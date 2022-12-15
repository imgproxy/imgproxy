# Amazon CloudWatch

imgproxy can send its metrics to AmazonCloudFront. To use this feature, do the following:

1. Set the `IMGPROXY_CLOUD_WATCH_SERVICE_NAME` environment variable. imgproxy will use the value of this variable as a value for the `ServiceName` dimension.
2. [Set up the necessary credentials](#set-up-credentials) to grant access to CloudWatch.
3. _(optional)_ Specify the AWS region with `IMGPROXY_CLOUD_WATCH_REGION` or `AWS_REGION`. Default: `us-west-1`
4. _(optional)_ Set the `IMGPROXY_CLOUD_WATCH_NAMESPACE` environment variable to be the desired CloudWatch namespace. Default: `imgproxy`

imgproxy sends the following metrics to CloudWatch:

* `RequestsInProgress`: the number of requests currently in progress
* `ImagesInProgress`: the number of images currently in progress
* `ConcurrencyUtilization`: the percentage of imgproxy's concurrency utilization. Calculated as `RequestsInProgress / IMGPROXY_CONCURRENCY * 100`
* `BufferSize`: a summary of the download buffers sizes (in bytes)
* `BufferDefaultSize`: calibrated default buffer size (in bytes)
* `BufferMaxSize`: calibrated maximum buffer size (in bytes)
* `VipsMemory`: libvips memory usage (in bytes)
* `VipsMaxMemory`: libvips maximum memory usage (in bytes)
* `VipsAllocs`: the number of active vips allocations

### Set up credentials

There are three ways to specify your AWS credentials. The credentials need to have rights to write metrics to CloudWatch:

#### IAM Roles

If you're running imgproxy on an Amazon Web Services platform, you can use IAM roles to to get the security credentials to make calls to AWS CloudWatch.

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

This is the recommended method when using dockerized imgproxy.

#### Shared credentials file

Alternatively, you can create the `.aws/credentials` file in your home directory with the following content:

```ini
[default]
aws_access_key_id = %access_key_id
aws_secret_access_key = %secret_access_key
```
