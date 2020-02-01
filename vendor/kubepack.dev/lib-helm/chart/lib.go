package chart

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
)

func GetChangedValues(original map[string]interface{}, modified map[string]interface{}) []string {
	var cmd []string
	getChangedValues(original, modified, "", &cmd)
	return cmd
}

func getChangedValues(original map[string]interface{}, modified map[string]interface{}, currentKey string, setCmds *[]string) {
	for key, val := range original {
		tempKey := ""
		if currentKey == "" {
			tempKey = key
		} else {
			tempKey = currentKey + "." + key
		}

		switch val := val.(type) {
		case map[string]interface{}:
			getChangedValues(val, modified[key].(map[string]interface{}), tempKey, setCmds)
		case []interface{}:
			if !cmp.Equal(val, modified[key]) {
				tempCmd := tempKey + "=["
				arrayLen := len(modified[key].([]interface{}))
				for i, element := range modified[key].([]interface{}) {
					tempCmd = tempCmd + fmt.Sprintf("%v", element)
					if i != arrayLen-1 {
						tempCmd = tempCmd + fmt.Sprintf(",")
					}
				}
				tempCmd = tempCmd + "]"
				if strings.Contains(tempCmd, "=[]") {
					tempCmd = strings.ReplaceAll(tempCmd, "[]", "null")
				}
				*setCmds = append(*setCmds, tempCmd)
			}
		case interface{}:
			if val != modified[key] {
				if isZeroOrNil(modified[key]) {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=null ", tempKey))
				} else {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=%v ", tempKey, modified[key]))
				}
			}
		default:
			if val != modified[key] {
				if isZeroOrNil(modified[key]) {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=null ", tempKey))
				} else {
					*setCmds = append(*setCmds, fmt.Sprintf("%s=%v ", tempKey, modified[key]))
				}
			}
		}
	}
}

func isZeroOrNil(x interface{}) bool {
	return x == nil || reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
