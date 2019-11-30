#!/usr/bin/env bash
# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

source common.sh

export TRACE=1
export GO111MODULE=on

# This test is used by prow and if the dep not be installed by git then it will face the GOBIN issue.
install_dep_by_git

function test_init_project {
  header_text "performing init project"
  kubebuilder init --project-version 1 --domain example.com <<< "n"
}

function test_make_project {
  header_text "running make in project"
  make
}

function test_create_api_controller {
  header_text "performing creating api and controller"
  kubebuilder create api --group insect --version v1beta1 --kind Bee --namespaced false <<EOF
y
y
EOF
}

function test_create_namespaced_api_controller {
  header_text "performing creating namespaced api and controller"
  kubebuilder create api --group insect --version v1beta1 --kind Bee --namespaced true <<EOF
y
y
EOF
}

function test_create_api_only {
  header_text "performing creating api only"
  kubebuilder create api --group insect --version v1beta1 --kind Bee --namespaced false <<EOF
y
n
EOF
}

function test_create_namespaced_api_only {
  header_text "performing creating api only"
  kubebuilder create api --group insect --version v1beta1 --kind Bee --namespaced true <<EOF
y
n
EOF
}

function test_create_skip {
  header_text "performing creating but skipping everything"
  kubebuilder create api --group insect --version v1beta1 --kind Bee <<EOF
n
n
EOF
}

function test_create_coretype_controller {
  header_text "performing creating coretype controller"
  kubebuilder create api --group apps --version v1 --kind Deployment --namespaced false <<EOF
n
y
EOF
}

function test_create_namespaced_coretype_controller {
  header_text "performing creating coretype controller"
  kubebuilder create api --group apps --version v1 --kind Deployment --namespaced true <<EOF
n
y
EOF
}


function test_project {
  project_dir=$1
  version=$2
  header_text "performing tests in dir $project_dir for project version v$version"
  vendor_tarball=$tmp_root/vendor.v$version.tgz
  old_gopath=$GOPATH
  if [[ $version == "1" ]]; then
      export GOPATH=$(pwd)/testdata/gopath
      download_vendor_archive
  fi
  cd testdata/$project_dir
  # v2 uses modules, and thus doesn't have a vendor directory
  [[ -e ${vendor_tarball} ]] && tar -zxf $vendor_tarball
  make all test # v2 doesn't test on all by default
  [[ -e ${vendor_tarball} ]] && rm -rf ./vendor && rm -f Gopkg.lock
  cd -
  export GOPATH=$old_gopath
}

prepare_staging_dir
fetch_tools
build_kb

setup_envs

export GO111MODULE=off
prepare_testdir_under_gopath
test_init_project
cache_project

prepare_testdir_under_gopath
dump_project
test_make_project

prepare_testdir_under_gopath
dump_project
test_create_api_controller

prepare_testdir_under_gopath
dump_project
test_create_namespaced_api_controller

prepare_testdir_under_gopath
dump_project
test_create_api_only

prepare_testdir_under_gopath
dump_project
test_create_namespaced_api_only

prepare_testdir_under_gopath
dump_project
test_create_coretype_controller

prepare_testdir_under_gopath
dump_project
test_create_namespaced_coretype_controller

header_text "running kubebuilder unit tests"
cd ${go_workspace}/src/github.com/eggsbenjamin/kubebuilder

export GO111MODULE=on
go test ./cmd/... ./pkg/...

# test project v1
# auto is roughly equivalent to off in our case,
# since we'll be in a gopath (basically, reset to default)
GO111MODULE=off test_project gopath/src/project 1

# test project v2
GO111MODULE=on test_project project-v2 2

exit $rc

