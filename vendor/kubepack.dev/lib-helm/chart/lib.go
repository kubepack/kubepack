package chart

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/alessio/shellescape"
)

func GetChangedValues(original, modified map[string]interface{}) ([]string, error) {
	cmds, err := getChangedValues(original, modified, "", nil)
	if err != nil {
		return nil, err
	}
	sort.Strings(cmds)
	return cmds, nil
}

func getChangedValues(original, modified map[string]interface{}, prefix string, cmds []string) ([]string, error) {
	for k, v := range modified {
		curKey := ""
		if prefix == "" {
			curKey = escapeKey(k)
		} else {
			curKey = prefix + "." + escapeKey(k)
		}

		switch val := v.(type) {
		case map[string]interface{}:
			oVal, ok := original[k].(map[string]interface{})
			if !ok {
				oVal = map[string]interface{}{}
			}
			next, err := getChangedValues(oVal, val, curKey, cmds)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, next...)
		case []interface{}:
			if !reflect.DeepEqual(v, original[k]) {
				if len(val) == 0 {
					cmds = append(cmds, fmt.Sprintf("%s=null", curKey))
					continue
				}

				if isSimpleArray(val) {
					s, err := PrintArray(val)
					if err != nil {
						return nil, fmt.Errorf("failed to print simple array %v, reason: %v", v, err)
					}
					cmds = append(cmds, fmt.Sprintf("%s=%s", curKey, s))
					continue
				}

				for i, element := range val {
					em, ok := element.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("%s[%d] element is not a map", curKey, i)
					}
					next, err := getChangedValues(map[string]interface{}{}, em, fmt.Sprintf("%s[%d]", curKey, i), cmds)
					if err != nil {
						return nil, err
					}
					cmds = append(cmds, next...)
				}
			}
		case string:
			if !reflect.DeepEqual(original[k], val) {
				cmds = append(cmds, fmt.Sprintf("%s=%v", curKey, escapeValue(val)))
			}
		case int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, json.Number:
			if !reflect.DeepEqual(original[k], val) {
				cmds = append(cmds, fmt.Sprintf("%s=%v", curKey, val))
			}
		case nil:
			if !reflect.DeepEqual(original[k], val) {
				cmds = append(cmds, fmt.Sprintf("%s=null", curKey))
			}
		default:
			return nil, fmt.Errorf("unknown type %v with value %v", reflect.TypeOf(v), v)
		}
	}
	return cmds, nil
}

// kubernetes.io/role becomes "kubernetes\.io/role"
func escapeKey(s string) string {
	if !strings.ContainsRune(s, '.') && !strings.ContainsRune(s, '\\') {
		return shellescape.Quote(s)
	}
	return shellescape.Quote(strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `.`, `\.`))
}

// "value1,value2" becomes value1\,value2
func escapeValue(s string) string {
	return shellescape.Quote(strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `,`, `\,`))
}

func isSimpleArray(a []interface{}) bool {
	for i := range a {
		switch a[i].(type) {
		case string, int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, nil, json.Number:
		default:
			return false
		}
	}
	return true
}

func PrintArray(a []interface{}) (string, error) {
	var buf bytes.Buffer
	buf.WriteRune('{')
	for i := range a {
		switch v := a[i].(type) {
		case string:
			if i > 0 {
				buf.WriteString(", ")
			}
			_, err := fmt.Fprint(&buf, escapeValue(v))
			if err != nil {
				return "", err
			}
		case int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, json.Number:
			if i > 0 {
				buf.WriteString(", ")
			}
			_, err := fmt.Fprint(&buf, v)
			if err != nil {
				return "", err
			}
		case nil:
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("null")
		default:
			return "", fmt.Errorf("[%d] holds a complex type %v", i, reflect.TypeOf(a[i]))
		}
	}
	buf.WriteRune('}')
	return buf.String(), nil
}
