#!/usr/bin/env bash

test_user="00000000-0000-0000-0000-000000000000"

echo "run docker with\ndocker run -e PUSH_S3_IMAGES_BUCKET=pushd-nonprod-images -e PUSH_S3_RENDER_BUCKET=pushd-nonprod-image-renders -e IMGPROXY_LOG_FORMAT=json -e IMGPROXY_USE_S3=true -e AWS_REGION=us-east-1 -e AWS_ACCESS_KEY_ID=<access key id> -e AWS_SECRET_ACCESS_KEY=<secret access key id>  -p 8080:8080 -it  <docker image name>"

mkdir -p downloads/known_good
mkdir -p downloads/new_version
mkdir -p downloads/s3_cached

test_files=("rotate_0__width_1536__height_2048__rt_fill__quality_95__crop_3024:4032:nowe:0:0__04693066-6F52-4E07-ADA7-A292313F1F6B.jpeg"
            "rotate_0__width_1280__height_1600__rt_fill__crop_1736:2420:nowe:0:0__padding_0:200:0:200__background_blur__04693066-6F52-4E07-ADA7-A292313F1F6B.jpeg"
            "rotate_0__width_960__height_1200__rt_fill__flower-portrait.jpg"
            "rotate_0__width_2048__height_1536__rt_fill__quality_95__crop_3066:2300:nowe:0:0__squirell-landscape.jpg"
            "rotate_0__width_1280__height_800__rt_fit__padding_0:319:0:319__background_blur__crop_750:937:nowe:0:106__house-landscape.heic"
            "rotate_0__width_1536__height_2048__rt_fill__crop_1868:2490:nowe:0:0__padding_0:0:0:0__background_blur__flower-portrait.jpg"
            "rotate_0__width_640__height_800__rt_fill__crop_3024:3780:nowe:0:252__grapes-portrait.heic"
            "quality_95__rt_fill__width_366__height_489__crop_1716:2289:nowe:79:0__lady-and-dog-portrait.jpg"
            "quality_95__rt_fill__width_372__height_372__crop_1023:1365:nowe:425:0__lady-and-dog-portrait.jpg"
            "rotate_0__width_1920__height_1200__rt_fill__crop_3034:1896:nowe:0:0__padding_0:0:0:0__background_0:0:0__lightning-landscape.jpg"
            "rotate_0__width_1920__height_1200__rt_fit__padding_0:480:0:480__background_0:0:0__quality_95__crop_3024:3780:nowe:0:0__beach-landscape.heic")

echo "cleaning up past runs"

# clear existing files
for test_file in "${test_files[@]}"; do
    rm ./downloads/known_good/"${test_file}" &> /dev/null
    rm ./downloads/new_version/"${test_file}" &> /dev/null
    rm ./downloads/s3_cached/"${test_file}" &> /dev/null
    AWS_PROFILE=pushd-nonprod aws s3 ls s3://pushd-nonprod-image-renders/00000000-0000-0000-0000-000000000000/"${test_file}" &> /dev/null
done

echo "Start up the known good imgproxy version with docker and then press enter"
read


download test assets from known good imgproxy version
for test_file in "${test_files[@]}"; do
    wget "http://localhost:8080/pushd/${test_user}/${test_file}" --directory-prefix=downloads/known_good/
done

echo "Finished downloading test files from known good imgproxy version"
echo "Start up the new imgproxy version with docker and then press enter"
read

download test assets from new imgproxy version
for test_file in "${test_files[@]}"; do
    wget "http://localhost:8080/pushd/${test_user}/${test_file}" --directory-prefix=downloads/new_version/
done

echo "Now we will phash the known_good and new_version files to see if they match"

failed="false"
# compare files to see if they match phashes
for test_file in "${test_files[@]}"; do
  diff_val=$(compare -metric phash downloads/known_good/"${test_file}" downloads/new_version/"${test_file}" /dev/null 2>&1)
  if (( $(echo "$diff_val > 10" | bc -l) )); then
    failed="true"
    echo "phash is bad for ${test_file} phash diff: ${diff_val}"
  else
    echo "phash passed. phash diff: ${diff_val}"
  fi
done

echo "Now we will phash the known_good and s3_cached files to see if they match"
# confirm files uploaded to S3 exist and also pass pHash
for test_file in "${test_files[@]}"; do
    AWS_PROFILE=pushd-nonprod aws s3 cp s3://pushd-nonprod-image-renders/00000000-0000-0000-0000-000000000000/"${test_file}" ./downloads/s3_cached/"${test_file}"
    diff_val=$(compare -metric phash downloads/known_good/"${test_file}" downloads/s3_cached/"${test_file}" /dev/null 2>&1)
    if (( $(echo "$diff_val > 10" | bc -l) )); then
      failed="true"
      echo "phash is bad for ${test_file} phash diff: ${diff_val}"
    else
      echo "phash passed. phash diff: ${diff_val}"
    fi
done


if [ "$failed" = "true" ]; then
    echo "At least one test file phash was too different."
    exit 1
else
    echo "Integration test has passed."
fi