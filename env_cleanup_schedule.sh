#!/bin/bash
set -euo pipefail
test -n "$1"
GITLAB_TOKEN=$1
GITLAB_TOKEN="${gitlab_api_scope_token}"

PROJECT_ID="7831967"
GITLAB_URL="https://gitlab.com/api/v4"
BEEST_TAG_NAME="beest-resource"
active_branches=()
days_threshold=180
threshold_date=$(date -d "$days_threshold days ago" +%s)
page=1

# Function to wait for EKS cluster deletion
function wait_for_cluster_deletion() {
  local CLUSTER_NAME=$1
  local MAX_ATTEMPTS=10  # Set a maximum number of attempts
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

while true; do
    set -x

    # Debug - check token not empty
    if [ -z "${gitlab_api_scope_token}" ]; then
        echo "token null"
        exit 1
    fi
#    echo ${gitlab_api_scope_token} > token.txt
    url="$GITLAB_URL/projects/$PROJECT_ID/repository/branches?protected=false&per_page=100&page=$page"
    branches_data=$(curl --header "PRIVATE-TOKEN: $GITLAB_TOKEN" "$url")
    if [[ "[]" = "$branches_data" ]]; then
        echo "END"
        break
    fi

    branch_names=$(echo "$branches_data" | jq -r '.[] | .name + ">" + (.commit.committed_date | gsub("T";" ") | gsub("Z";""))')
    mapfile -t branch_info_array < <(echo "$branch_names")
    for branch_info in "${branch_info_array[@]}"; do
        IFS='>' read branch commit_date <<< "$branch_info"
        if [[ -n "$commit_date" ]]; then
            commit_epoch=$(date -d "$commit_date" +%s 2>/dev/null)
            if [[ -n "$commit_epoch" ]]; then
                if [[ $commit_epoch -gt $threshold_date ]]; then
                    ref_name_objects=$(echo "${branch:0:18}" | tr '[:upper:]' '[:lower:]')
                    active_branches+=("$ref_name_objects") # Append directly to the array
                fi
            fi
        fi
    done

    page=$((page + 1))
    set +x
done

echo "Start removing old terraformLocks"
TERRAFORM_TABLE_NAME=$(aws dynamodb list-tables --output json | jq -r '.TableNames[] | select(contains("beest"))') || true
TERRAFORM_LOCK_IDS=($(aws dynamodb scan --table-name "$TERRAFORM_TABLE_NAME" --output json | jq -r '.Items[].LockID.S'))

for TERRAFORM_LOCK_ID in "${TERRAFORM_LOCK_IDS[@]}"; do
  match_found=false
  for branch in "${active_branches[@]}"; do
    if [[ "$TERRAFORM_LOCK_ID" == *"$branch"* ]]; then
      match_found=true
      break
    fi
  done

  if ! $match_found; then
       echo "Terraform lock '$TERRAFORM_LOCK_ID' exists. Deleting..."
       aws dynamodb delete-item --table-name "$TERRAFORM_TABLE_NAME" --key '{"LockID"':' {"S"':' "'"$TERRAFORM_LOCK_ID"'"}}' || true
  fi
done
echo "END"

echo "Start clearing s3 bucket"
S3_BUCKET_NAME=beest-terraform-state
S3_BUCKET_OBJECTS=($(aws s3api list-objects-v2 --bucket $S3_BUCKET_NAME | jq -r '.Contents[].Key')) || true

for S3_BUCKET_OBJECT in "${S3_BUCKET_OBJECTS[@]}"; do
  match_found=false
  for branch in "${active_branches[@]}"; do
    if [[ "$S3_BUCKET_OBJECT" == *"$branch"* ]]; then
      match_found=true
      break
    fi
  done

  if ! $match_found; then
       echo "S3 bucket '$S3_BUCKET_OBJECT' exists. Deleting..."
       aws s3 rm s3://$S3_BUCKET_NAME/$S3_BUCKET_OBJECT || true
  fi
done
echo "END"

echo "Start clearing autoscaling group"
AUTOSCALING_GROUP_NAMES=($(aws autoscaling describe-auto-scaling-groups --query 'AutoScalingGroups[*].[AutoScalingGroupName, Tags[].Value]' --output json | jq -r --arg beest_tag_name "$BEEST_TAG_NAME" '.[] | select(.[1] | any(. == $beest_tag_name) and .[0]) | .[0]')) || true
for AUTOSCALING_GROUP_NAME in "${AUTOSCALING_GROUP_NAMES[@]}"; do
  match_found=false
  for branch in "${active_branches[@]}"; do
    if [[ "$AUTOSCALING_GROUP_NAME" == *"$branch"* ]]; then
      match_found=true
      break
    fi
  done

  if ! $match_found; then
      echo "Auto Scaling group '$AUTOSCALING_GROUP_NAME' exists. Deleting..."
      aws autoscaling delete-auto-scaling-group --auto-scaling-group-name "$AUTOSCALING_GROUP_NAME" --force-delete || true
  fi
done
echo "END"

echo "Start clearing eks cluster"
CLUSTER_NAMES=($(aws eks list-clusters --query 'clusters[]' --output json | jq -r '.[]')) || true

for CLUSTER_NAME in "${CLUSTER_NAMES[@]}"; do
    CLUSTER_TAG=$(aws eks describe-cluster --name "$CLUSTER_NAME" --query 'cluster.tags.Name' --output json | jq -r '.')

    match_found=false
    for branch in "${active_branches[@]}"; do
        if [[ "$CLUSTER_NAME" == *"$branch"* ]]; then
            match_found=true
            break
        fi
    done

    if ! $match_found && [[ "$CLUSTER_TAG" != *"$BEEST_TAG_NAME"* ]]; then
        echo "Cluster '$CLUSTER_NAME' exists. Deleting..."
        aws eks delete-cluster --name "$CLUSTER_NAME" --query 'cluster.status' --output json &> /dev/null
        wait_for_cluster_deletion "$CLUSTER_NAME"
    fi
done
echo "END"

echo "Start clearing instance profiles"
INSTANCE_PROFILE_NAMES=($(aws iam list-instance-profiles --query 'InstanceProfiles[*].InstanceProfileName' --output json | jq -r '.[]')) || true

for INSTANCE_PROFILE_NAME in "${INSTANCE_PROFILE_NAMES[@]}"; do
    match_found=false

  for branch in "${active_branches[@]}"; do
    if [[ "$INSTANCE_PROFILE_NAME" == *"$branch"* ]]; then
      match_found=true
      break
    fi
  done

    if ! $match_found && [[ "$INSTANCE_PROFILE_NAME" == *"beest"*  ]]; then
        echo "Instance profile '$INSTANCE_PROFILE_NAME' exists. Deleting..."
        aws iam delete-instance-profile --instance-profile-name "$INSTANCE_PROFILE_NAME" --query 'InstanceProfile.InstanceProfileName' --output json &> /dev/null
    fi
done
echo "END"


echo "Start clearing EC2 key pair"
KEY_PAIRS=($(aws ec2 describe-key-pairs --query 'KeyPairs[*].[KeyName, Tags[*].Value]' --output json | jq -r --arg beest_tag_name "$BEEST_TAG_NAME" '.[] | select(.[1] | any(. == $beest_tag_name) and .[0]) | .[0]')) || true

for KEY_PAIR in "${KEY_PAIRS[@]}"; do
    match_found=false

  for branch in "${active_branches[@]}"; do
    if [[ "$KEY_PAIR" == *"$branch"* ]]; then
      match_found=true
      break
    fi
  done

    if ! $match_found; then
        echo "Ec2 - Key pair '$KEY_PAIR' exists. Deleting..."
        aws ec2 delete-key-pair --key-name "$KEY_PAIR" &> /dev/null
    fi
done
echo "END"

