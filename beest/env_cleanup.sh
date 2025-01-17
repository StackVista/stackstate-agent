#!/bin/bash

REF_NAME_OBJECTS=$(echo "${CI_COMMIT_REF_NAME:0:18}" | tr '[:upper:]' '[:lower:]')
echo $REF_NAME_OBJECTS

echo "Start removing terraformLock"
TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))')
echo $TERRAFORM_TABLE_NAME
TERRAFORM_LOCK_ID=$(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Items[] | select(.LockID.S | contains($ref_name_objects)) | .LockID.S')
echo $TERRAFORM_LOCK_ID
aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}'
echo "END"

echo "Start clearing s3 bucket"
S3_BUCKET_NAME=beest-terraform-state
S3_BUCKET_OBJECT=$(aws s3api list-objects-v2 --bucket $S3_BUCKET_NAME | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Contents[] | select(.Key | contains($ref_name_objects)) | .Key')
echo $S3_BUCKET_OBJECT
aws s3 rm s3://$S3_BUCKET_NAME/$S3_BUCKET_OBJECT
echo "END"
