package repository

import (
	"github.com/rapinbook/item-shop-go/entities"
)

type ItemManagingRepository interface {
	Creating(itemEntity *entities.Item) (*entities.Item, error)
	Editing(itemID uint64, itemEditingReq *_itemManagingModel.ItemEditingReq) (uint64, error)
	// Archiving(itemID uint64) error // Soft delete
}
