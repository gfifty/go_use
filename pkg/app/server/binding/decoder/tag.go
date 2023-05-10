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
	"reflect"
	"strings"
)

const (
	pathTag     = "path"
	formTag     = "form"
	queryTag    = "query"
	cookieTag   = "cookie"
	headerTag   = "header"
	jsonTag     = "json"
	rawBodyTag  = "raw_body"
	fileNameTag = "file_name"
)

const (
	defaultTag = "default"
)

const (
	requiredTagOpt = "required"
)

type TagInfo struct {
	Key      string
	Value    string
	Required bool
	Default  string
	Options  []string
	Getter   getter
}

func head(str, sep string) (head, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}

func lookupFieldTags(field reflect.StructField) []TagInfo {
	var ret []string
	tags := []string{pathTag, formTag, queryTag, cookieTag, headerTag, jsonTag, rawBodyTag, fileNameTag}
	for _, tag := range tags {
		if _, ok := field.Tag.Lookup(tag); ok {
			ret = append(ret, tag)
		}
	}

	defaultVal := ""
	if val, ok := field.Tag.Lookup(defaultTag); ok {
		defaultVal = val
	}

	var tagInfos []TagInfo
	for _, tag := range ret {
		tagContent := field.Tag.Get(tag)
		tagValue, opts := head(tagContent, ",")
		var options []string
		var opt string
		var required bool
		for len(opts) > 0 {
			opt, opts = head(opts, ",")
			options = append(options, opt)
			if opt == requiredTagOpt {
				required = true
			}
		}
		tagInfos = append(tagInfos, TagInfo{Key: tag, Value: tagValue, Options: options, Required: required, Default: defaultVal})
	}

	return tagInfos
}

func getDefaultFieldTags(field reflect.StructField) (tagInfos []TagInfo) {
	defaultVal := ""
	if val, ok := field.Tag.Lookup(defaultTag); ok {
		defaultVal = val
	}

	tags := []string{pathTag, formTag, queryTag, cookieTag, headerTag, jsonTag, fileNameTag}
	for _, tag := range tags {
		tagInfos = append(tagInfos, TagInfo{Key: tag, Value: field.Name, Default: defaultVal})
	}

	return
}