package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/alessio/shellescape"
	kj "gomodules.xyz/encoding/json"
	"sigs.k8s.io/yaml"
)

func GetValuesMapDiff(original, modified map[string]any) (map[string]any, error) {
	return getValuesDiff(original, modified, "", nil)
}

func getValuesDiff(original, modified map[string]any, prefix string, diff map[string]any) (map[string]any, error) {
	if diff == nil {
		diff = map[string]any{}
	}

	for k, v := range modified {
		curKey := ""
		if prefix == "" {
			curKey = escapeKey(k)
		} else {
			curKey = prefix + "." + escapeKey(k)
		}

		switch val := v.(type) {
		case map[string]any:
			oVal, ok := original[k].(map[string]any)
			if !ok {
				oVal = map[string]any{}
			}

			d2, err := getValuesDiff(oVal, val, curKey, nil)
			if err != nil {
				return nil, err
			}
			if len(d2) > 0 {
				diff[k] = d2
			}
		case []any, string, int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, json.Number, nil:
			if origVal, ok := original[k]; !ok || !reflect.DeepEqual(origVal, val) {
				diff[k] = val
			}
		default:
			return nil, fmt.Errorf("unknown type %v with value %v", reflect.TypeOf(v), v)
		}
	}

	// https://github.com/kubepack/lib-helm/blob/32de2acacbfb84f57d4a66c6d896360eb664399c/pkg/values/options.go#L133
	for k, v := range original {
		if _, found := modified[k]; !found {
			curKey := ""
			if prefix == "" {
				curKey = escapeKey(k)
			} else {
				curKey = prefix + "." + escapeKey(k)
			}

			// TODO: how does Helm merge --values remove keys?
			// diff[k] = nil
			return nil, fmt.Errorf("key %s is missing in the modified values, original values %v", curKey, v)
		}
	}
	return diff, nil
}

func GetValuesDiff(orig, od any) (map[string]any, error) {
	origMap, err := kj.ToJsonMap(orig)
	if err != nil {
		return nil, err
	}
	modMap, err := kj.ToJsonMap(od)
	if err != nil {
		return nil, err
	}

	return GetValuesMapDiff(origMap, modMap)
}

func GetValuesDiffYAML(orig, od any) ([]byte, error) {
	origMap, err := kj.ToJsonMap(orig)
	if err != nil {
		return nil, err
	}
	modMap, err := kj.ToJsonMap(od)
	if err != nil {
		return nil, err
	}

	diff, err := GetValuesMapDiff(origMap, modMap)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(diff)
}

func GetValuesDiffJson(orig, od any) ([]byte, error) {
	origMap, err := kj.ToJsonMap(orig)
	if err != nil {
		return nil, err
	}
	modMap, err := kj.ToJsonMap(od)
	if err != nil {
		return nil, err
	}

	diff, err := GetValuesMapDiff(origMap, modMap)
	if err != nil {
		return nil, err
	}
	return kj.Marshal(diff)
}

func GetChangedValues(original, modified map[string]any) ([]string, error) {
	cmds, err := getChangedValues(original, modified, "", nil)
	if err != nil {
		return nil, err
	}
	sort.Strings(cmds)
	return cmds, nil
}

func getChangedValues(original, modified map[string]any, prefix string, cmds []string) ([]string, error) {
	for k, v := range modified {
		curKey := ""
		if prefix == "" {
			curKey = escapeKey(k)
		} else {
			curKey = prefix + "." + escapeKey(k)
		}

		switch val := v.(type) {
		case map[string]any:
			oVal, ok := original[k].(map[string]any)
			if !ok {
				oVal = map[string]any{}
			}
			next, err := getChangedValues(oVal, val, curKey, nil)
			if err != nil {
				return nil, err
			}
			cmds = append(cmds, next...)
		case []any:
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
					em, ok := element.(map[string]any)
					if !ok {
						return nil, fmt.Errorf("%s[%d] element is not a map", curKey, i)
					}
					next, err := getChangedValues(map[string]any{}, em, fmt.Sprintf("%s[%d]", curKey, i), nil)
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
			if origVal, ok := original[k]; !ok || !reflect.DeepEqual(origVal, val) {
				cmds = append(cmds, fmt.Sprintf("%s=null", curKey))
			}
		default:
			return nil, fmt.Errorf("unknown type %v with value %v", reflect.TypeOf(v), v)
		}
	}

	for k := range original {
		if _, found := modified[k]; !found {
			curKey := ""
			if prefix == "" {
				curKey = escapeKey(k)
			} else {
				curKey = prefix + "." + escapeKey(k)
			}

			cmds = append(cmds, fmt.Sprintf("%s=null", curKey))
		}
	}
	return cmds, nil
}

// kubernetes.io/role becomes "kubernetes\.io/role"
func escapeKey(s string) string {
	return shellescape.Quote(strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `.`, `\.`))
}

// "value1,value2" becomes value1\,value2
func escapeValue(s string) string {
	return shellescape.Quote(strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `,`, `\,`))
}

func isSimpleArray(a []any) bool {
	for i := range a {
		switch a[i].(type) {
		case string, int8, uint8, int16, uint16, int32, uint32, int64, uint64, int, uint, float32, float64, bool, nil, json.Number:
		default:
			return false
		}
	}
	return true
}

func PrintArray(a []any) (string, error) {
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
