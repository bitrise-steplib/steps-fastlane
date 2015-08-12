#!/bin/bash

set -e

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

# Print configs
echo "Params:"
echo "* work_dir: ${work_dir}"
echo "* lane_name: ${lane_name}"

set -v

# Install fastlane
gem install fastlane --no-document

# Running fastlane actions
cd "${work_dir}"
fastlane ${lane_name}