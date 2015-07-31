#!/bin/bash

set -e

#
# Required parameters
if [ -z "${fastlane_action}" ] ; then
  echo "Missing required input: fastlane_action"
  exit 1
fi

#
# Install fastlane
echo "Installing fastlane"
gem install fastlane

#
# Running fastlane actions
fastlane "${fastlane_action}"