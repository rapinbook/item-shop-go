package service

import (
	_itemManagingRepository "github.com/rapinbook/item-shop-go/pkg/itemManaging/repository"
)

type itemManagingServiceImpl struct {
	itemManagingRepository _itemManagingRepository.ItemManagingRepository
}
