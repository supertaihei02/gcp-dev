#!/bin/bash

if [ "$1" = "" ]; then
  echo "application id not set"
  exit 1
fi

if [ "$2" = "" ]; then
  echo "version not set"
  exit 1
fi

/google-cloud-sdk/platform/google_appengine/appcfg.py --application $1 --version $2 update --oauth2_access_token=$(gcloud auth print-access-token 2> /dev/null) ./src/login/app.yaml
