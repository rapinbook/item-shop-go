package repository

import (
	"github.com/labstack/echo/v4"
	"github.com/rapinbook/item-shop-go/databases"
	"github.com/rapinbook/item-shop-go/entities"
	_itemManagingModel "github.com/rapinbook/item-shop-go/pkg/itemManaging/model"
)

type itemManagingRepositoryImpl struct {
	db     databases.Database
	logger echo.Logger
}

func NewItemMangingRepository(db databases.Database, logger echo.Logger) ItemManagingRepository {
	return &itemManagingRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

func (r *itemManagingRepositoryImpl) Creating(itemEntity *entities.Item) (*entities.Item, error) {
	item := new(entities.Item)

	err := r.db.Connect().Create(&itemEntity).Scan(item).Error
	if err != nil {
		r.logger.Errorf("Cannot insert item to table")
		return item, err
	}

	return item, nil
}

func (r *itemManagingRepositoryImpl) Editing(itemID uint64, itemEditingReq *_itemManagingModel.ItemEditingReq) (uint64, error) {
	item := new(entities.Item)
	err := r.db.Connect().First(&item, "id = ?", itemID).Updates(
		itemEditingReq,
	).Error
	if err != nil {
		r.logger.Errorf("Cannot Update item from table")
		return 0, err
	}

	return itemID, nil
}
