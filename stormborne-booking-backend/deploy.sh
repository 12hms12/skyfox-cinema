#!/bin/bash
set -e
set -v
export TAG=$3

# used  in compose file
export COMMIT_SHA=$3
export ENVIRONMENT=$4
export TEAM_ID=$TEAM_ID

export BOOKING_IMAGE=$1:$COMMIT_SHA-$ENVIRONMENT
export POSTGRES_IMAGE=$2:$COMMIT_SHA-$ENVIRONMENT
export MIGRATION_IMAGE=$3:$COMMIT_SHA-$ENVIRONMENT
export BACKEND_PORT=$ENV_BACKEND_PORT
export DB_PORT=$ENV_DB_PORT_CONNECT
export HOST_DB_PORT=$ENV_DB_PORT
export DB_HOST=$ENV_DB_HOST
export UI_HOST=$ENV_UI_HOST
export MOVIE_SERVICE_HOST=$ENV_MOVIE_SERVICE_HOST
export "$(cat Makefile | grep appVersion)" # bash syntax
export VERSION=$appVersion-$COMMIT_SHA

apt-get install jq -y
curl -o /usr/local/bin/ecs-cli https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest
echo "$(curl -s https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest.md5) /usr/local/bin/ecs-cli" | md5sum -c -
chmod +x /usr/local/bin/ecs-cli

ecs-cli configure --cluster "$CLUSTER_NAME" --default-launch-type EC2 --config-name "$CLUSTER_CONFIG_NAME" --region ap-south-1
ecs-cli configure profile --access-key "$AWS_ACCESS_KEY" --secret-key "$AWS_SECRET_KEY" --profile-name "$CLUSTER_PROFILE_NAME"

./aws_check_cluster_status.sh

echo  "using image.. $BOOKING_IMAGE"
timestamp="$(date +"%s")"
creds_file_name="ecs-registry-creds_$timestamp.yml"
cp ecs-registry-creds.yml "$creds_file_name"
sed -i -e "s/ENV_GITLAB_REGISTRY_SECRET_ARN/$GITLAB_REGISTRY_SECRET_ARN/g" "ecs-registry-creds_$timestamp.yml"

ecs-cli compose --verbose --registry-creds "$creds_file_name" --project-name "booking$TEAM_ID-$ENVIRONMENT" down --cluster-config "$CLUSTER_CONFIG_NAME" --ecs-profile "$CLUSTER_PROFILE_NAME"
ecs-cli ps --cluster-config "$CLUSTER_CONFIG_NAME" --ecs-profile "$CLUSTER_PROFILE_NAME"

ecs-cli compose --verbose --registry-creds "$creds_file_name" --project-name "booking$TEAM_ID-$ENVIRONMENT" --cluster-config "$CLUSTER_CONFIG_NAME" --ecs-profile "$CLUSTER_PROFILE_NAME" up --create-log-groups --force-update

ecs-cli compose --verbose --registry-creds "$creds_file_name" --project-name "booking$TEAM_ID-$ENVIRONMENT" down --cluster-config "$CLUSTER_CONFIG_NAME" --ecs-profile "$CLUSTER_PROFILE_NAME"
ecs-cli compose --verbose --registry-creds "$creds_file_name" --project-name "booking$TEAM_ID-$ENVIRONMENT" service up --cluster-config "$CLUSTER_CONFIG_NAME" --ecs-profile "$CLUSTER_PROFILE_NAME" --deployment-min-healthy-percent 0

./verify.sh
rm -rf ecs-registry-creds_*.yml

set +v