/*
 * Copyright 2017 Huawei Technologies Co., Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
* Created by on 2017/6/22.
 */

// Package configmanager provides deserializer
package configmanager

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/ServiceComb/go-archaius/core/cast"
)

const (
	configClientTag  = `yaml`
	ignoreField      = `ignoredField` // when used -
	doNotConsiderTag = ``
)

/*
   unmarshal configurations on supplied object.
   multi level configuration key structure > source.module.type.config: value
   simple key structure > config: value
*/
func (cMgr *ConfigurationManager) unmarshal(rValue reflect.Value, tagName string) (err error) {
	// handle panic
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("unmarshalling failed")
		}
	}()

	switch rValue.Kind() {
	case reflect.Ptr:
		err := cMgr.handlePtr(rValue, getTagKey(tagName, doNotConsiderTag))
		if err != nil {
			return err
		}

	case reflect.Struct:
		err := cMgr.handleStruct(rValue, getTagKey(tagName, doNotConsiderTag))
		if err != nil {
			return err
		}
	case reflect.Map:
		err := cMgr.handleMap(rValue, getTagKey(tagName, doNotConsiderTag))
		if err != nil {
			return err
		}
	case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Bool, reflect.Interface, reflect.Array, reflect.Slice:
		if rValue.CanSet() {
			err := cMgr.setValue(rValue, tagName)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// handle pointer type objects
func (cMgr *ConfigurationManager) handlePtr(rValue reflect.Value, tagName string) error {
	if rValue.IsNil() {
		ptrValue := reflect.New(rValue.Type().Elem())
		err := cMgr.unmarshal(ptrValue, getTagKey(tagName, doNotConsiderTag))
		if err != nil {
			return err
		}

		if rValue.CanSet() {
			rValue.Set(ptrValue)
		}
		return nil
	} else if rValue.Elem().Kind() == reflect.Ptr {
		ptrValue := rValue.Elem()
		err := cMgr.handlePtr(ptrValue, getTagKey(tagName, doNotConsiderTag))
		if err != nil {
			return err
		}
	}

	ptrValue := rValue.Elem()
	err := cMgr.unmarshal(ptrValue, getTagKey(tagName, doNotConsiderTag))
	if err != nil {
		return err
	}

	return nil
}

// get multi level configuration key
func getTagKey(currentTag, addTag string) string {
	if currentTag == doNotConsiderTag && addTag == doNotConsiderTag {
		return doNotConsiderTag
	} else if currentTag == doNotConsiderTag && addTag != doNotConsiderTag {
		return addTag
	} else if currentTag != doNotConsiderTag && addTag == doNotConsiderTag {
		return currentTag
	}

	return currentTag + `.` + addTag
}

// handle struct type object
func (cMgr *ConfigurationManager) handleStruct(rValue reflect.Value, tagName string) error {
	structType := rValue.Type()
	numOfField := structType.NumField()

	for i := 0; i < numOfField; i++ {
		structField := structType.Field(i)
		fieldValue := rValue.Field(i)
		keyName := cMgr.getKeyName(structField.Name, structField.Tag)
		if keyName == ignoreField {
			return nil
		}

		switch structField.Type.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Interface, reflect.Array,
			reflect.Slice:
			if fieldValue.CanSet() {
				err := cMgr.setValue(fieldValue, getTagKey(tagName, keyName))
				if err != nil {
					return err
				}
			}
		case reflect.Ptr:
			err := cMgr.handlePtr(fieldValue, getTagKey(tagName, keyName))
			if err != nil {
				return err
			}
		case reflect.Struct:
			err := cMgr.handleStruct(fieldValue, getTagKey(tagName, keyName))
			if err != nil {
				return err
			}
		case reflect.Map:
			err := cMgr.handleMap(fieldValue, getTagKey(tagName, keyName))
			if err != nil {
				return err
			}
		case reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func,
			reflect.UnsafePointer:
			// ignore
		}
	}

	return nil
}

// handle map
func (cMgr *ConfigurationManager) handleMap(rValue reflect.Value, tagName string) error {
	if tagName == doNotConsiderTag {
		if rValue.CanSet() {
			configValue := cMgr.GetConfigurations()
			if configValue == nil {
				return nil
			}
			configRValue := reflect.ValueOf(configValue)
			rValue.Set(configRValue)
		}

		return nil
	}

	mapType := rValue.Type()
	// check if key is not string return error
	if mapType.Key().Kind() != reflect.String {
		return errors.New("map key should be string")
	}

	mapValue, err := cMgr.populateMap(tagName, mapType)
	if err != nil {
		return err
	}

	// if assignable then only assign
	if mapValue.Type() != mapType {
		return fmt.Errorf("value types of %s not matched. expect type : %s, config client type : %s",
			tagName, rValue.Kind(), mapValue.Kind())
	}

	if rValue.CanSet() {
		rValue.Set(mapValue)
	}

	return nil
}

// generate map from config map
func (cMgr *ConfigurationManager) populateMap(prefix string, mapType reflect.Type) (reflect.Value, error) {
	rValuePtr := reflect.New(mapType)
	rValue := rValuePtr.Elem()
	rValue.Set(reflect.MakeMap(mapType))
	//rValue := reflect.MakeMap(mapType)
	var mapKeys []string
	mapValueType := rValue.Type().Elem()

	configValue := cMgr.GetConfigurations()
	for key := range configValue {
		isPrifix, index := checkPrefix(key, prefix)
		if !isPrifix {
			continue
		}

		mapKeys = append(mapKeys, key[index:])
	}

	for _, key := range mapKeys {
		// if key itself has map value stored
		if key == "" {
			val := cMgr.GetConfigurationsByKey(prefix)
			setVal := reflect.ValueOf(val)
			if mapType != setVal.Type() {
				return rValue, fmt.Errorf("invalid value for map %s", mapType.String())
			}
			if rValue.CanSet() {
				rValue.Set(setVal)
			}
			return rValue, nil
		}

		switch mapValueType.Kind() {
		// for '.' separated configurations
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Float32, reflect.Float64, reflect.Uint, reflect.Uint8, reflect.Uint16,
			reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Interface:
			val := cMgr.GetConfigurationsByKey(prefix + key)
			setVal := reflect.ValueOf(val)

			if mapValueType != setVal.Type() {
				returnCongValue, err := ToRvalueType(setVal.Interface(), mapValueType)
				if err != nil {
					return rValue, fmt.Errorf("value types of %s not matched. expect type : %s, config client type : %s",
						prefix+key, mapValueType, setVal.String())
				}

				setVal = returnCongValue
			}

			if rValue.CanSet() {
				rValue.SetMapIndex(reflect.ValueOf(key[1:]), setVal)
			}
		default:
			splitKey := strings.Split(key, `.`)
			mapKey := splitKey[1]
			mapValue := reflect.New(mapValueType)
			err := cMgr.unmarshal(mapValue, getTagKey(prefix, mapKey))
			if err != nil {
				return rValue, err
			}

			if rValue.CanSet() {
				rValue.SetMapIndex(reflect.ValueOf(mapKey), mapValue.Elem())
			}
		}
	}

	return rValue, nil
}

func checkPrefix(heap, prefix string) (bool, int) {
	if len(heap) < len(prefix) {
		return false, 0
	}

	var index int
	for i := range prefix {
		if heap[i] != prefix[i] {
			break
		}
		index++
	}

	if len(prefix) != index {
		return false, 0
	}

	return true, index
}

// set values in object
func (cMgr *ConfigurationManager) setValue(rValue reflect.Value, keyName string) error {
	configValue := cMgr.GetConfigurationsByKey(keyName)
	if configValue == nil {
		return nil
	}

	// assign value if assignable
	configRValue := reflect.ValueOf(configValue)
	if configRValue.Kind() != rValue.Kind() {
		returnCongValue, err := ToRvalueType(configRValue.Interface(), rValue.Type())
		if err != nil {
			return fmt.Errorf("value types of %s not matched. expect type : %s, config client type : %s",
				keyName, rValue.Kind(), configRValue.Kind())
		}

		configRValue = returnCongValue
	}

	if rValue.CanSet() {
		rValue.Set(configRValue)
	}

	return nil
}

// get key from tag
func (*ConfigurationManager) getKeyName(fieldName string, fieldTagName reflect.StructTag) string {
	tagName := fieldTagName.Get(configClientTag)
	if tagName == "-" {
		return ignoreField
	} else if tagName == "" {
		return toSnake(fieldName)
	}

	return tagName
}

//convert camel case to snake case
func toSnake(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) ||
			unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

// ToRvalueType Deserializes the object to a particular type
func ToRvalueType(confValue interface{}, convertType reflect.Type) (returnValue reflect.Value, err error) {
	castValue := cast.NewValue(confValue, nil)
	returnValue = reflect.New(convertType).Elem()

	switch convertType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		returnInt, rErr := castValue.ToInt64()
		if err != nil {
			err = rErr
		}
		returnValue.SetInt(returnInt)

	case reflect.String:
		returnString, rErr := castValue.ToString()
		if err != nil {
			err = rErr
		}

		returnValue.SetString(returnString)

	case reflect.Float32, reflect.Float64:
		returnFloat, rErr := castValue.ToFloat64()
		if err != nil {
			err = rErr
		}
		returnValue.SetFloat(returnFloat)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		returnUInt, rErr := castValue.ToUint64()
		if err != nil {
			err = rErr
		}
		returnValue.SetUint(returnUInt)
	case reflect.Bool:
		returnBool, rErr := castValue.ToBool()
		if err != nil {
			err = rErr
		}
		returnValue.SetBool(returnBool)
	default:
		err = errors.New("canot convert type")
	}

	return returnValue, err
}
