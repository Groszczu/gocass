package carts

import (
	"github.com/Groszczu/gocass/internal/models"
	"github.com/Groszczu/gocass/internal/repository"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

type CartRepository = repository.Repository[models.CartsStruct]
type CartProductRepository interface {
	repository.Repository[models.CartProductsStruct]
	SumProductPrices(cartId models.UUID) (int, error)
}
type DiscountCodeRepository = repository.Repository[models.DiscountCodesStruct]
type DiscountCodeUsageRepository interface {
	repository.Repository[models.DiscountCodeUsagesStruct]
	IncreaseCodeUsageCount(codeUsage *models.DiscountCodeUsagesStruct, amount int) error
}
type OrderRepository = repository.Repository[models.OrdersStruct]

func newCartRepository(session *gocqlx.Session) CartRepository {
	return repository.New[models.CartsStruct](session, models.Carts)
}

func newDiscountCodeRepository(session *gocqlx.Session) DiscountCodeRepository {
	return repository.New[models.DiscountCodesStruct](session, models.DiscountCodes)
}

type discountCodeUsageRepositoryImpl struct {
	repository.Repository[models.DiscountCodeUsagesStruct]
}

func newDiscountCodeUsageRepository(session *gocqlx.Session) DiscountCodeUsageRepository {
	return discountCodeUsageRepositoryImpl{
		repository.New[models.DiscountCodeUsagesStruct](session, models.DiscountCodeUsages),
	}
}

func (r discountCodeUsageRepositoryImpl) IncreaseCodeUsageCount(codeUsage *models.DiscountCodeUsagesStruct, amount int) error {
	session := r.Session()
	table := r.TableDefinition()

	builder := table.UpdateBuilder().Add("usage_count")
	query := builder.Query(*session).BindMap(qb.M{
		"usage_count": amount,
		"code":        codeUsage.Code,
	})

	return query.Exec()
}

func newOrderRepository(session *gocqlx.Session) OrderRepository {
	return repository.New[models.OrdersStruct](session, models.Orders)
}

type cartProductRepositoryImpl struct {
	repository.Repository[models.CartProductsStruct]
}

func newCartProductRepository(session *gocqlx.Session) CartProductRepository {
	return cartProductRepositoryImpl{
		repository.New[models.CartProductsStruct](session, models.CartProducts),
	}
}

func (r cartProductRepositoryImpl) SumProductPrices(cartId models.UUID) (int, error) {
	session := r.Session()
	table := r.TableDefinition()

	builder := table.SelectBuilder().Sum("price_in_cents * quantity")
	query := builder.Query(*session).BindMap(qb.M{
		"cart_id": cartId,
	})

	var result int
	if err := query.Get(&result); err != nil {
		return 0, err
	}

	return result, nil
}
