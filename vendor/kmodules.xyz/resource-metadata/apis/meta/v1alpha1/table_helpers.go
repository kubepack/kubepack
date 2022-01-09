/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"

func (in *TableCell) DeepCopy() *TableCell {
	if in == nil {
		return nil
	}

	out := new(TableCell)
	if in.Data != nil {
		out.Data = runtime.DeepCopyJSONValue(in.Data)
	}
	if in.Sort != nil {
		out.Sort = runtime.DeepCopyJSONValue(in.Sort)
	}
	out.Link = in.Link
	out.Icon = in.Icon
	out.Color = in.Color
	return out
}

func Convert_ResourceColumnDefinition_To_ResourceColumn(def ResourceColumnDefinition) ResourceColumn {
	col := ResourceColumn{
		Name:   def.Name,
		Type:   def.Type,
		Format: def.Format,
	}
	if def.Sort != nil && def.Sort.Enable {
		col.Sort = &SortHeader{
			Enable: true,
			Type:   def.Sort.Type,
			Format: def.Sort.Format,
		}
	}
	if def.Link != nil && def.Link.Enable {
		col.Link = true
	}
	if def.Icon != nil && def.Icon.Enable {
		col.Icon = true
	}
	if def.Shape != "" {
		col.Shape = def.Shape
	}
	if def.Color != nil && def.Color.Color != "" {
		col.Color = def.Color.Color
	}
	return col
}
