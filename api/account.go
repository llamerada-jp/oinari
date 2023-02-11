package api

type Account struct {
	Meta   *ObjectMeta    `json:"meta"`
	Status *AccountStatus `json:"status"`
}

type AccountStatus struct {
	Pods map[string]string `json:"pods"`
}
