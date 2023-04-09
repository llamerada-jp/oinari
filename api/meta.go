package api

import "fmt"

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

func (meta *ObjectMeta) Validate(t ResourceType, uuid string) error {
	if meta.Type != t {
		return fmt.Errorf("type field should be %s", t)
	}

	if len(meta.Name) == 0 {
		return fmt.Errorf("name of the resource should be specify")
	}

	if len(meta.Owner) == 0 {
		return fmt.Errorf("owner of the resource should be specify")
	}

	if len(meta.CreatorNode) == 0 {
		return fmt.Errorf("creator node of the resource should be specify")
	}

	if err := ValidateNodeId(meta.CreatorNode); err != nil {
		return fmt.Errorf("invalid creator node was specified (%s): %w",
			meta.CreatorNode, err)
	}

	if len(meta.Uuid) == 0 {
		return fmt.Errorf("uuid of the resource should be specify")
	}

	if meta.Uuid != uuid {
		return fmt.Errorf("invalid uuid was specified (%s)", meta.Uuid)
	}

	if len(meta.DeletionTimestamp) != 0 {
		if err := ValidateTimestamp(meta.DeletionTimestamp); err != nil {
			return fmt.Errorf("invalid deletion timestamp was specified (%s): %w",
				meta.DeletionTimestamp, err)
		}
	}

	return nil
}
