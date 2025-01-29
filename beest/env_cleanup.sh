#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Missing BRANCH NAME parameter"
    exit 1
fi

CI_COMMIT_REF_NAME=$1
REF_NAME_OBJECTS=$(echo "${CI_COMMIT_REF_NAME:0:18}" | tr '[:upper:]' '[:lower:]')
echo $REF_NAME_OBJECTS

# Function to wait for EKS cluster deletion
function wait_for_cluster_deletion() {
  local CLUSTER_NAME=$1

  while true; do
    CLUSTER_STATUS=$(aws eks describe-cluster --name "$CLUSTER_NAME" --query 'cluster.status' --output json | jq -r 'select(.) | .')
    if [[ "$CLUSTER_STATUS" == "DELETING" ]]; then
      echo "Cluster '$CLUSTER_NAME' is still deleting..."
      sleep 30
    elif [[ -z "$CLUSTER_STATUS" ]]; then
      echo "Cluster '$CLUSTER_NAME' has been deleted."
      break
    else
      echo "Error getting cluster status: $CLUSTER_STATUS"
      break
    fi
  done
}

echo "Start removing terraformLock"
TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))')
echo $TERRAFORM_TABLE_NAME
TERRAFORM_LOCK_ID=$(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Items[] | select(.LockID.S | contains($ref_name_objects)) | .LockID.S')
if [[ -n "$TERRAFORM_LOCK_ID" ]]; then
  echo "Terraform lock '$TERRAFORM_LOCK_ID' exists. Deleting..."
  aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}'
fi
echo "END"

echo "Start clearing s3 bucket"
S3_BUCKET_NAME=beest-terraform-state
S3_BUCKET_OBJECT=$(aws s3api list-objects-v2 --bucket $S3_BUCKET_NAME | jq -r --arg ref_name_objects "$REF_NAME_OBJECTS" '.Contents[] | select(.Key | contains($ref_name_objects)) | .Key')
if [[ -n "$S3_BUCKET_OBJECT" ]]; then
  echo "S3 bucket '$S3_BUCKET_OBJECT' exists. Deleting..."
 aws s3 rm s3://$S3_BUCKET_NAME/$S3_BUCKET_OBJECT
fi
echo "END"

echo "Start clearing autoscaling group"
AUTOSCALING_GROUP_NAME=$(aws autoscaling describe-auto-scaling-groups --query 'AutoScalingGroups[*].AutoScalingGroupName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))')
if [[ -n "$AUTOSCALING_GROUP_NAME" ]]; then
  echo "Auto Scaling group '$AUTOSCALING_GROUP_NAME' exists. Deleting..."
  aws autoscaling delete-auto-scaling-group --auto-scaling-group-name "$AUTOSCALING_GROUP_NAME" --force-delete
fi
echo "END"

echo "Start clearing eks cluster"
CLUSTER_NAME=$(aws eks list-clusters --query 'clusters[]' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))')
echo $CLUSTER_NAME
if [[ -n "$CLUSTER_NAME" ]]; then
  echo "Cluster '$CLUSTER_NAME' exists. Deleting..."
  aws eks delete-cluster --name "$CLUSTER_NAME" --query 'cluster.status' --output json &> /dev/null
  wait_for_cluster_deletion "$CLUSTER_NAME"
fi
echo "END"

echo "Start clearing Role - IAM"
ROLE_NAMES=$(aws iam list-roles --query 'Roles[*].RoleName' --output json | jq -r '.[] | select(contains("'"$REF_NAME_OBJECTS"'"))')
for ROLE_NAME in $ROLE_NAMES; do
    echo "Role - IA '$ROLE_NAME' exists. Checking the instance profiles..."
    INSTANCE_PROFILE_NAMES=$(aws iam list-instance-profiles-for-role --role-name "$ROLE_NAME" --query 'InstanceProfiles[*].InstanceProfileName' --output json | jq -r '.[]')

    for INSTANCE_PROFILE_NAME in $INSTANCE_PROFILE_NAMES; do
        echo "Removing role '$ROLE_NAME' from instance profile '$INSTANCE_PROFILE_NAME'"
        aws iam remove-role-from-instance-profile --instance-profile-name "$INSTANCE_PROFILE_NAME" --role-name "$ROLE_NAME" &> /dev/null
    done

    INLINE_POLICES=$(aws iam list-role-policies --role-name "$ROLE_NAME" --query 'PolicyNames[]'  --output json | jq -r '.[]')
    for INLINE_POLICE in $INLINE_POLICES; do
      echo "Deleting inline policy '$INLINE_POLICE' from role '$ROLE_NAME'"
      aws iam delete-role-policy --role-name "$ROLE_NAME" --policy-name "$INLINE_POLICE" &> /dev/null
    done

    MANAGED_POLICES=$(aws iam list-attached-role-policies --role-name "$ROLE_NAME" --query 'AttachedPolicies[*].PolicyArn'  --output json | jq -r '.[]')
    for MANAGED_POLICE in $MANAGED_POLICES; do
      echo "Detaching managed policy '$MANAGED_POLICE' from role '$ROLE_NAME'"
      aws iam detach-role-policy --role-name "$ROLE_NAME" --policy-arn "$MANAGED_POLICE" &> /dev/null
    done

    echo "Role - IA '$ROLE_NAME' exists. Deleting..."
    aws aws iam delete-role --role-name "$ROLE_NAME" --query 'Role.RoleName' --output json &> /dev/null
done
echo "END"
