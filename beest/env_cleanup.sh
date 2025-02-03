#!/bin/bash
set -euo pipefail

test -n "$1"
INITIAL_REF_NAME=$1
REF_NAME_OBJECTS=$(echo "${INITIAL_REF_NAME:0:18}" | tr '[:upper:]' '[:lower:]')

# Function to wait for EKS cluster deletion
function wait_for_cluster_deletion() {
  local CLUSTER_NAME=$1
  local MAX_ATTEMPTS=5  # Set a maximum number of attempts
  local ATTEMPTS=0

  while true; do
    ATTEMPTS=$((ATTEMPTS + 1))

    if [[ "$ATTEMPTS" -gt "$MAX_ATTEMPTS" ]]; then
      echo "Cluster '$CLUSTER_NAME' deletion timed out after $MAX_ATTEMPTS attempts."
      return 1
    fi

    CLUSTER_STATUS=$(aws eks describe-cluster --name "$CLUSTER_NAME" --query 'cluster.status' --output json | jq -r 'select(.) | .') || true

    if [[ "$CLUSTER_STATUS" == "DELETING" ]]; then
      echo "Cluster '$CLUSTER_NAME' is still deleting... (Attempt $ATTEMPTS/$MAX_ATTEMPTS)"
      sleep 30
    elif [[ -z "$CLUSTER_STATUS" ]]; then
      echo "Cluster '$CLUSTER_NAME' has been deleted."
      return 0
    else
      echo "Error getting cluster status: $CLUSTER_STATUS (Attempt $ATTEMPTS/$MAX_ATTEMPTS)"
      return 1
    fi
  done
}

echo "Start removing terraformLock"
TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))') || true
TERRAFORM_LOCK_ID=$(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Items[] | select(.LockID.S | contains($ref_name_objects)) | .LockID.S') || true
if [[ -n "$TERRAFORM_LOCK_ID" ]]; then
  echo "Terraform lock '$TERRAFORM_LOCK_ID' exists. Deleting..."
  aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}' || true
fi
echo "END"

echo "Start clearing s3 bucket"
S3_BUCKET_NAME=beest-terraform-state
S3_BUCKET_OBJECT=$(aws s3api list-objects-v2 --bucket $S3_BUCKET_NAME | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Contents[] | select(.Key | contains($ref_name_objects)) | .Key') || true
if [[ -n "$S3_BUCKET_OBJECT" ]]; then
  echo "S3 bucket '$S3_BUCKET_OBJECT' exists. Deleting..."
  aws s3 rm s3://$S3_BUCKET_NAME/$S3_BUCKET_OBJECT || true
fi
echo "END"

echo "Start clearing autoscaling group"
AUTOSCALING_GROUP_NAME=$(aws autoscaling describe-auto-scaling-groups --query 'AutoScalingGroups[*].AutoScalingGroupName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))') || true
if [[ -n "$AUTOSCALING_GROUP_NAME" ]]; then
  echo "Auto Scaling group '$AUTOSCALING_GROUP_NAME' exists. Deleting..."
  aws autoscaling delete-auto-scaling-group --auto-scaling-group-name "$AUTOSCALING_GROUP_NAME" --force-delete || true
fi
echo "END"

echo "Start clearing eks cluster"
CLUSTER_NAME=$(aws eks list-clusters --query 'clusters[]' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))') || true
if [[ -n "$CLUSTER_NAME" ]]; then
  echo "Cluster '$CLUSTER_NAME' exists. Deleting..."
  aws eks delete-cluster --name "$CLUSTER_NAME" --query 'cluster.status' --output json &> /dev/null
  wait_for_cluster_deletion "$CLUSTER_NAME"
fi
echo "END"

echo "Start clearing Role - IAM"
ROLE_NAMES=$(aws iam list-roles --query 'Roles[*].RoleName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))') || true
for ROLE_NAME in $ROLE_NAMES; do
    echo "Role - IA '$ROLE_NAME' exists. Checking the instance profiles..."
    INSTANCE_PROFILE_NAMES=$(aws iam list-instance-profiles-for-role --role-name "$ROLE_NAME" --query 'InstanceProfiles[*].InstanceProfileName' --output json | jq -r '.[]') || true

    for INSTANCE_PROFILE_NAME in $INSTANCE_PROFILE_NAMES; do
        echo "Removing role '$ROLE_NAME' from instance profile '$INSTANCE_PROFILE_NAME'"
        aws iam remove-role-from-instance-profile --instance-profile-name "$INSTANCE_PROFILE_NAME" --role-name "$ROLE_NAME" &> /dev/null
    done

    INLINE_POLICIES=$(aws iam list-role-policies --role-name "$ROLE_NAME" --query 'PolicyNames[]'  --output json | jq -r '.[]') || true
    for INLINE_POLICY in $INLINE_POLICIES; do
      echo "Deleting inline policy '$INLINE_POLICY' from role '$ROLE_NAME'"
      aws iam delete-role-policy --role-name "$ROLE_NAME" --policy-name "$INLINE_POLICY" &> /dev/null
    done

    MANAGED_POLICIES=$(aws iam list-attached-role-policies --role-name "$ROLE_NAME" --query 'AttachedPolicies[*].PolicyArn'  --output json | jq -r '.[]') || true
    for MANAGED_POLICY in $MANAGED_POLICIES; do
      echo "Detaching managed policy '$MANAGED_POLICY' from role '$ROLE_NAME'"
      aws iam detach-role-policy --role-name "$ROLE_NAME" --policy-arn "$MANAGED_POLICY" &> /dev/null
    done

    echo "Role - IA '$ROLE_NAME' exists. Deleting..."
    aws iam delete-role --role-name "$ROLE_NAME" --query 'Role.RoleName' --output json &> /dev/null

    for INSTANCE_PROFILE_NAME in $INSTANCE_PROFILE_NAMES; do
        echo "Removing instance profile '$INSTANCE_PROFILE_NAME'"
        aws iam delete-instance-profile --instance-profile-name "$INSTANCE_PROFILE_NAME" &> /dev/null
    done
done
echo "END"

echo "Start clearing instance profiles"
INSTANCE_PROFILE_NAME=$(aws iam list-instance-profiles --query 'InstanceProfiles[*].InstanceProfileName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))') || true
if [[ -n "$INSTANCE_PROFILE_NAME" ]]; then
  echo "Instance profile '$INSTANCE_PROFILE_NAME' exists. Deleting..."
  aws iam delete-instance-profile --instance-profile-name "$INSTANCE_PROFILE_NAME" --query 'InstanceProfile.InstanceProfileName' --output json &> /dev/null
fi
echo "END"

echo "Start clearing EC2 key pair"
KEY_PAIR=$(aws ec2 describe-key-pairs --query 'KeyPairs[*].KeyName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))') || true
if [[ -n "$KEY_PAIR" ]]; then
  echo "Ec2 - Key pair '$KEY_PAIR' exists. Deleting..."
  aws ec2 delete-key-pair --key-name "$KEY_PAIR" &> /dev/null
fi
echo "END"
