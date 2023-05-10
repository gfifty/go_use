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
 * MIT License
 *
 * Copyright (c) 2019-present Fenny and Contributors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2022 CloudWeGo Authors
 */

package decoder

import (
	"fmt"
	"mime/multipart"
	"reflect"

	"github.com/cloudwego/hertz/internal/bytesconv"
	path1 "github.com/cloudwego/hertz/pkg/app/server/binding/path"
	hjson "github.com/cloudwego/hertz/pkg/common/json"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type sliceTypeFieldTextDecoder struct {
	fieldInfo
	isArray bool
}

func (d *sliceTypeFieldTextDecoder) Decode(req *bindRequest, params path1.PathParam, reqValue reflect.Value) error {
	var texts []string
	var defaultValue string
	for _, tagInfo := range d.tagInfos {
		if tagInfo.Key == jsonTag || tagInfo.Key == fileNameTag {
			continue
		}
		if tagInfo.Key == headerTag {
			tagInfo.Value = utils.GetNormalizeHeaderKey(tagInfo.Value, req.Req.Header.IsDisableNormalizing())
		}
		texts = tagInfo.Getter(req, params, tagInfo.Value)
		// todo: array/slice default value
		defaultValue = tagInfo.Default
		if len(texts) != 0 {
			break
		}
	}
	if len(texts) == 0 && len(defaultValue) != 0 {
		texts = append(texts, defaultValue)
	}
	if len(texts) == 0 {
		return nil
	}

	reqValue = GetFieldValue(reqValue, d.parentIndex)
	field := reqValue.Field(d.index)
	// **[]**int
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			nonNilVal, ptrDepth := GetNonNilReferenceValue(field)
			field.Set(ReferenceValue(nonNilVal, ptrDepth))
		}
	}
	var parentPtrDepth int
	for field.Kind() == reflect.Ptr {
		field = field.Elem()
		parentPtrDepth++
	}

	if d.isArray {
		if len(texts) != field.Len() {
			return fmt.Errorf("%q is not valid value for %s", texts, field.Type().String())
		}
	} else {
		// slice need creating enough capacity
		field = reflect.MakeSlice(field.Type(), len(texts), len(texts))
	}

	// handle internal multiple pointer, []**int
	var ptrDepth int
	t := d.fieldType.Elem()
	elemKind := t.Kind()
	for elemKind == reflect.Ptr {
		t = t.Elem()
		elemKind = t.Kind()
		ptrDepth++
	}

	for idx, text := range texts {
		var vv reflect.Value
		vv, err := stringToValue(t, text)
		if err != nil {
			return err
		}
		field.Index(idx).Set(ReferenceValue(vv, ptrDepth))
	}
	reqValue.Field(d.index).Set(ReferenceValue(field, parentPtrDepth))

	return nil
}

func getSliceFieldDecoder(field reflect.StructField, index int, tagInfos []TagInfo, parentIdx []int) ([]fieldDecoder, error) {
	if !(field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array) {
		return nil, fmt.Errorf("unexpected type %s, expected slice or array", field.Type.String())
	}
	isArray := false
	if field.Type.Kind() == reflect.Array {
		isArray = true
	}
	for idx, tagInfo := range tagInfos {
		switch tagInfo.Key {
		case pathTag:
			tagInfos[idx].Getter = path
		case formTag:
			tagInfos[idx].Getter = form
		case queryTag:
			tagInfos[idx].Getter = query
		case cookieTag:
			tagInfos[idx].Getter = cookie
		case headerTag:
			tagInfos[idx].Getter = header
		case jsonTag:
			// do nothing
		case rawBodyTag:
			tagInfo.Getter = rawBody
		case fileNameTag:
			// do nothing
		default:
		}
	}

	fieldType := field.Type
	for field.Type.Kind() == reflect.Ptr {
		fieldType = field.Type.Elem()
	}
	t := getElemType(fieldType.Elem())
	if t == reflect.TypeOf(multipart.FileHeader{}) {
		return getMultipartFileDecoder(field, index, tagInfos, parentIdx)
	}

	return []fieldDecoder{&sliceTypeFieldTextDecoder{
		fieldInfo: fieldInfo{
			index:       index,
			parentIndex: parentIdx,
			fieldName:   field.Name,
			tagInfos:    tagInfos,
			fieldType:   fieldType,
		},
		isArray: isArray,
	}}, nil
}

func stringToValue(elemType reflect.Type, text string) (v reflect.Value, err error) {
	v = reflect.New(elemType).Elem()
	switch elemType.Kind() {
	case reflect.Struct:
		err = hjson.Unmarshal(bytesconv.S2b(text), v.Addr().Interface())
	case reflect.Map:
		err = hjson.Unmarshal(bytesconv.S2b(text), v.Addr().Interface())
	case reflect.Array, reflect.Slice:
		// do nothing
	default:
		decoder, err := SelectTextDecoder(elemType)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("unsupported type %s for slice/array", elemType.String())
		}
		err = decoder.UnmarshalString(text, v)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("unable to decode '%s' as %s: %w", text, elemType.String(), err)
		}
	}

	return v, err
}
