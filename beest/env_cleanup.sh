#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Missing BRANCH NAME parameter"
    exit 1
fi

CI_COMMIT_REF_NAME=$1
REF_NAME_OBJECTS=$(echo "${CI_COMMIT_REF_NAME:0:18}" | tr '[:upper:]' '[:lower:]')
echo $REF_NAME_OBJECTS

echo "Start removing terraformLock"
TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))')
echo $TERRAFORM_TABLE_NAME
TERRAFORM_LOCK_ID=$(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Items[] | select(.LockID.S | contains($ref_name_objects)) | .LockID.S')
if [[ "$TERRAFORM_LOCK_ID" -gt 0 ]]; then
  echo "Terraform lock '$TERRAFORM_LOCK_ID' exists. Deleting..."
  aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}'
fi
echo "END"

echo "Start clearing s3 bucket"
S3_BUCKET_NAME=beest-terraform-state
S3_BUCKET_OBJECT=$(aws s3api list-objects-v2 --bucket $S3_BUCKET_NAME | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Contents[] | select(.Key | contains($ref_name_objects)) | .Key')
if [[ "$S3_BUCKET_OBJECT" -gt 0 ]]; then
  echo "S3 bucket '$S3_BUCKET_OBJECT' exists. Deleting..."
 aws s3 rm s3://$S3_BUCKET_NAME/$S3_BUCKET_OBJECT
fi
echo "END"

echo "Start clearing autoscaling group"
AUTOSCALING_GROUP_NAME=$(aws autoscaling describe-auto-scaling-groups --query 'AutoScalingGroups[*].AutoScalingGroupName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))')
if [[ "$AUTOSCALING_GROUP_NAME" -gt 0 ]]; then
  echo "Auto Scaling group '$AUTOSCALING_GROUP_NAME' exists. Deleting..."
  aws autoscaling delete-auto-scaling-group --auto-scaling-group-name "$AUTOSCALING_GROUP_NAME" --force-delete
fi
echo "END"
