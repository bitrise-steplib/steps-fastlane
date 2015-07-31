#!/bin/bash

set -e

#
# Required parameters
if [ -z "${fastlane_action}" ] ; then
  echo "Missing required input: fastlane_action"
  exit 1
fi

if [ -z "${work_dir}" ] ; then
  echo "Missing required input: work_dir"
  exit 1
fi

#
# Install fastlane
echo "Installing fastlane"
gem install fastlane
echo

#
# Running fastlane actions
cd "${work_dir}"
echo "Running action"
fastlane "${fastlane_action}"