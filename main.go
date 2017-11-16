// Copyright 2014 qiufeng-sun. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"util/logs"
	"util/run"

	"core/server"
)

var _ = logs.Debug

// main entrance for gate
//  -- tranfer msg to special server which assigned by protobuf target default value
//  -- cache server info to conn. server info is specified by to client msg
func main() {
	defer run.Recover(true)

	server.Run(NewGate())
}
