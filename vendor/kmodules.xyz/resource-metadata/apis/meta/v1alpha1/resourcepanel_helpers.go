/*
Copyright The Kmodules Authors.

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

func (section *PanelSection) Contains(rd *ResourceDescriptor) bool {
	for _, entry := range section.Entries {
		if entry.Type != nil &&
			entry.Type.Group == rd.Spec.Resource.Group &&
			entry.Type.Version == rd.Spec.Resource.Version &&
			entry.Type.Resource == rd.Spec.Resource.Name {
			return true
		}
	}
	return false
}

func (e PanelEntry) Equals(other PanelEntry) bool {
	if e.Type != nil && other.Type != nil {
		return *e.Type == *other.Type
	} else if e.Type == nil && other.Type == nil {
		return e.Path == other.Path
	}
	return false
}

func (a *ResourcePanel) Minus(b *ResourcePanel) {
	for _, bs := range b.Sections {
	NEXT_ENTRY:
		for _, be := range bs.Entries {
			for _, as := range a.Sections {
				for idx, ae := range as.Entries {
					if ae.Equals(be) {
						as.Entries = append(as.Entries[:idx], as.Entries[idx+1:]...)
						continue NEXT_ENTRY
					}
				}
			}
		}
	}
}
