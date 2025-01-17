#!/bin/bash

TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))')
echo $TERRAFORM_TABLE_NAME
TERRAFORM_LOCK_ID=$(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r --arg ci_commit_ref_name "${CI_COMMIT_REF_NAME:0:18}" '.Items[] | select(.LockID.S | contains($ci_commit_ref_name | ascii_downcase)) | .LockID.S')
echo $TERRAFORM_LOCK_ID
aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}'
