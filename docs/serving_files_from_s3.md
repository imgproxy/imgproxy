# Serving files from S3

imgproxy can process images from S3 buckets. To use this feature, do the following:

1. Set the `IMGPROXY_USE_S3` environment variable to be `true`.
2. [Set up the necessary credentials](#set-up-credentials) to grant access to your bucket.
3. _(optional)_ Specify the AWS region with `IMGPROXY_S3_REGION` or `AWS_REGION`. Default: `us-west-1`
4. _(optional)_ Specify the S3 endpoint with `IMGPROXY_S3_ENDPOINT`.
5. _(optional)_ Set the `IMGPROXY_S3_MULTI_REGION` environment variable to be `true`.
6. _(optional)_ Specify the AWS IAM Role to Assume with `IMGPROXY_S3_ASSUME_ROLE_ARN`
7. Use `s3://%bucket_name/%file_key` as the source image URL.

If you need to specify the version of the source object, you can use the query string of the source URL:

```
s3://%bucket_name/%file_key?%version_id
```

### Set up credentials

There are three ways to specify your AWS credentials. The credentials need to have read rights for all of the buckets given in the source URLs:

#### IAM Roles

If you're running imgproxy on an Amazon Web Services platform, you can use IAM roles to to get the security credentials to make calls to AWS S3.

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

#### Cross-Account Access

S3 access credentials may be acquired by assuming a role using STS. To do so specify the IAM Role arn with the `IMGPROXY_S3_ASSUME_ROLE_ARN` environment variable. This approach still requires you to provide initial AWS credentials by using one of the ways described above. The provided credentials role should allow assuming the role with provided ARN.

## Multi-Region mode

By default, imgproxy allows using S3 buckets located in a single region specified with `IMGPROXY_S3_REGION` or `AWS_REGION`. If your buckets are located in different regions, set `IMGPROXY_S3_MULTI_REGION` environment variable to be `true` to enable multi-region mode. In this mode, imgproxy will make an additional request to determine the bucket's region when the bucket is accessed for the first time.

In this mode, imgroxy uses a region specified with  `IMGPROXY_S3_REGION` or `AWS_REGION` to determine the endpoint to which it should send the bucket's region determination request. Thus, it's a good idea to use one of these variables to specify a region closest to the imgproxy instance.

## MinIO

[MinIO](https://github.com/minio/minio) is an object storage server released under Apache License v2.0. It is compatible with Amazon S3, so it can be used with imgproxy.

To use MinIO as source images provider, do the following:

* Set up Amazon S3 support as usual using environment variables or a shared config file.
* Specify an endpoint with `IMGPROXY_S3_ENDPOINT`. Use the `http://...` endpoint to disable SSL.
