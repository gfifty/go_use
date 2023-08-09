//go:build (linux || windows || darwin) && amd64 && !stdjson
// +build linux windows darwin
// +build amd64
// +build !stdjson

/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package decoder

import (
	"github.com/cloudwego/hertz/internal/bytesconv"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"strings"

	"github.com/bytedance/sonic"
)

func checkRequireJSON(req *bindRequest, tagInfo TagInfo) bool {
	if !tagInfo.Required {
		return true
	}
	ct := bytesconv.B2s(req.Req.Header.ContentType())
	if utils.FilterContentType(ct) != "application/json" {
		return false
	}
	node, _ := sonic.Get(req.Req.Body(), stringSliceForInterface(tagInfo.JSONName)...)
	if !node.Exists() {
		idx := strings.LastIndex(tagInfo.JSONName, ".")
		if idx > 0 {
			// There should be a superior if it is empty, it will report 'true' for required
			node, _ := sonic.Get(req.Req.Body(), stringSliceForInterface(tagInfo.JSONName[:idx])...)
			if !node.Exists() {
				return true
			}
		}
		return false
	}
	return true
}

func stringSliceForInterface(s string) (ret []interface{}) {
	x := strings.Split(s, ".")
	for _, val := range x {
		ret = append(ret, val)
	}
	return
}
