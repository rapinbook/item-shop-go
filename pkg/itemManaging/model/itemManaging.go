package model

type (
	ItemCreatingReq struct {
		AdminID     string
		Name        string `json:"name" validate:"omitempty,max=64"`
		Description string `json:"description" validate:"omitempty,max=128"`
		Picture     string `json:"picture" validate:"omitempty"`
		Price       uint   `json:"price" validate:"omitempty"`
	}

	ItemEditingReq struct {
		AdminID     string
		Name        string `json:"name" validate:"omitempty,max=64"`
		Description string `json:"description" validate:"omitempty,max=128"`
		Picture     string `json:"picture" validate:"omitempty"`
		Price       uint   `json:"price" validate:"omitempty"`
	}
)
