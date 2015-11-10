#!/bin/bash

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

# Install fastlane
if [ -f './Gemfile' ] ; then
  echo
  echo "Found 'Gemfile' - using it..."
  bundle install --verbose
  echo
  echo "Fastlane version:"
  bundle exec fastlane --version
else
  echo "gem install fastlane --no-document"
  gem install fastlane --no-document
fi
echo

# Running fastlane actions
echo "cd \"${work_dir}\""
cd "${work_dir}"

echo "fastlane ${lane_name}"
fastlane ${lane_name}
