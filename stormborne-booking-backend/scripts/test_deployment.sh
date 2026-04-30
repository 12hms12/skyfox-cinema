#!/bin/bash
set -e
set -v

echo "Running this as `whoami`"
echo "Currently running this in $PWD"

IFS='-' read -ra identifiers <<< $DEPLOYMENT_GROUP_NAME  # DEPLOYMENT_GROUP_NAME is of form neev-xx-xx-[backend|frontend]-[deployment|seed]-[integration|staging|production]
export BATCH_ID=${identifiers[1]}
export TEAM_ID=${identifiers[2]}
export TASK=${identifiers[4]}
export ENVIRONMENT=${identifiers[5]}
export PREFIX="/neev-$BATCH_ID/team-$TEAM_ID/$ENVIRONMENT"

# horrible hack because of the horrible code deploy programming model!
# so codedeploy will place artifacts in a directory specified in appspec.yml, but it won't let you parameterize anything in the appspec.yml file
# this works as long as you have one environment per server, but if you have multiple - it won't work as artifacts will be overriden
# so one creates this environment specific directories themselves and move the artifacts
# this probably means that at anytime only one environment could be deployed too (if this works at all)
# that is, one cannot simultaneously deploy to staging and to integration.
cd /home/ec2-user/deployment/
mkdir -p $ENVIRONMENT
mv docker-compose.yml $ENVIRONMENT/
mv outputs.sh $ENVIRONMENT/
mv seedShowData.sh $ENVIRONMENT/
cd $ENVIRONMENT
echo "Existing contents of the directory are"
ls

export POSTGRES_IMAGE=`aws ssm get-parameters --name "$PREFIX/POSTGRES_IMAGE" | jq ".Parameters[0].Value" | tr -d \"`
export MIGRATION_IMAGE=`aws ssm get-parameters --name "$PREFIX/MIGRATION_IMAGE" | jq ".Parameters[0].Value" | tr -d \"`
export HOST_DB_PORT=`aws ssm get-parameters --name "$PREFIX/HOST_DB_PORT" | jq ".Parameters[0].Value"  | tr -d \"`
export DB_HOST=`aws ssm get-parameters --name "$PREFIX/DB_HOST" | jq ".Parameters[0].Value" | tr -d \"`
export DB_PORT=`aws ssm get-parameters --name "$PREFIX/DB_PORT" | jq ".Parameters[0].Value" | tr -d \"`
# export DB_HOST=db # unique within the network
# export DB_PORT=5432 # for reasons I don't understand, its always 5432 (even if db container is mapped to a different port)
export POSTGRES_USERNAME=`aws ssm get-parameters --name "$PREFIX/POSTGRES_USERNAME" | jq ".Parameters[0].Value" | tr -d \"`
export DB_NAME=`aws ssm get-parameters --name "$PREFIX/DB_NAME" | jq ".Parameters[0].Value" | tr -d \"`
export UI_HOST=`aws ssm get-parameters --name "$PREFIX/UI_HOST" | jq ".Parameters[0].Value" | tr -d \"`
export MOVIE_SERVICE_HOST=`aws ssm get-parameters --name "$PREFIX/MOVIE_SERVICE_HOST" | jq ".Parameters[0].Value" | tr -d \"`
export BACKEND_PORT=`aws ssm get-parameters --name "$PREFIX/BACKEND_PORT" | jq ".Parameters[0].Value" | tr -d \"`
export POSTGRES_PASSWORD=`aws ssm get-parameters --name "$PREFIX/POSTGRES_PASSWORD" | jq ".Parameters[0].Value" | tr -d \"`
export REGISTRY_ID=`aws ssm get-parameters --name "$PREFIX/REGISTRY_ID" | jq ".Parameters[0].Value" | tr -d \"`

. ./outputs.sh # exports VERSION and BOOKING_IMAGE

env > /home/ec2-user/envs_available_at_deploytime_$ENVIRONMENT
echo "Logging into ECR"
$(aws ecr get-login --no-include-email --registry-ids $REGISTRY_ID)
/home/ec2-user/bin/docker-compose down || true
echo docker network list
/home/ec2-user/bin/docker-compose up -d

if [ "$TASK" == "seed" ]; then
  docker cp seedShowData.sh db_$ENVIRONMENT:/usr/local/bin
  docker exec db_$ENVIRONMENT /bin/bash -c "seedShowData.sh $(date +%Y-%m-%d) 5"
fi
