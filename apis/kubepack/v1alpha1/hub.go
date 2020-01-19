package v1alpha1

type Hub struct {
	Repositories []Repository `json:"repositories" protobuf:"bytes,1,rep,name=repositories"`
}

type Repository struct {
	Name        string        `json:"name" protobuf:"bytes,1,opt,name=name"`
	URL         string        `json:"url" protobuf:"bytes,2,opt,name=url"`
	Maintainers []ContactData `json:"maintainers,omitempty" protobuf:"bytes,3,rep,name=maintainers"`
}
