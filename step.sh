#!/bin/bash

set -e

# Required parameters
if [ -z "${lane}" ] ; then
  echo "Missing required input: lane"
  exit 1
fi
export lane_name="${lane}"

if [ -z "${work_dir}" ] ; then
  echo "Missing required input: work_dir"
  exit 1
fi

echo

# Running fastlane actions
echo '$' cd "${work_dir}"
cd "${work_dir}"

# Install fastlane
if [ -f './Gemfile' ] ; then
  echo
  echo "Found 'Gemfile' - using it..."
  echo '$' bundle install
  bundle install
  echo
  echo "Fastlane version:"
  echo '$' bundle exec fastlane --version
  bundle exec fastlane --version
else
  echo " (i) No Gemfile found - using system installed fastlane ..."
  echo '$' gem install fastlane --no-document
  gem install fastlane --no-document
  echo
  echo "Fastlane version:"
  echo '$' fastlane --version
  fastlane --version
fi

echo
echo '$' fastlane "${lane_name}"
fastlane ${lane_name}
