package api

type ResourceType string

const (
	ResourceTypeAccount = ResourceType("account")
	ResourceTypeNode    = ResourceType("node")
	ResourceTypePod     = ResourceType("pod")
)

type ObjectMeta struct {
	Type              ResourceType `json:"type"`
	Name              string       `json:"name"`
	Owner             string       `json:"owner"`
	CreatorNode       string       `json:"creatorNode"`
	Uuid              string       `json:"uuid"`
	DeletionTimestamp string       `json:"deletionTimestamp"`
}
