#!/bin/bash

# Required parameters
if [ -z "${fastlane_action}" ] ; then
  echo "Missing required input: fastlane_action"
  exit 1
fi
export lane_name="${fastlane_action}"

if [ -z "${work_dir}" ] ; then
  echo "Missing required input: work_dir"
  exit 1
fi

# Install fastlane
echo "gem install fastlane --no-document"
gem install fastlane --no-document
echo

# Running fastlane actions
echo "cd \"${work_dir}\""
cd "${work_dir}"

echo "set -eu -o pipefail && fastlane ${lane_name}"
fastlane ${lane_name}