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

cmd_prefix=""

# Install fastlane
if [ -f './Gemfile' ] ; then
  echo
  echo " (i) Found 'Gemfile' - using it..."
  echo '$' bundle install
  bundle install

  cmd_prefix="bundle exec"
else
  echo " (i) No Gemfile found - using system installed fastlane ..."
  echo " (i) Updating system installed fastlane ..."
  echo '$' gem install fastlane --no-document
  gem install fastlane --no-document
fi

echo
echo "Fastlane version:"
echo '$' $cmd_prefix fastlane --version
$cmd_prefix fastlane --version

echo
echo "Run fastlane:"
echo '$' $cmd_prefix fastlane "${lane_name}"
$cmd_prefix fastlane ${lane_name}
