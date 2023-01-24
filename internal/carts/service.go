package carts

import (
	"errors"
	"math"

	"github.com/Groszczu/gocass/internal/models"
	"github.com/scylladb/gocqlx/v2"
)

type Service struct {
	cartRepo              CartRepository
	cartProductRepo       CartProductRepository
	discountCodeRepo      DiscountCodeRepository
	discountCodeUsageRepo DiscountCodeUsageRepository
	orderRepo             OrderRepository
}

func NewService(session *gocqlx.Session) Service {
	return Service{
		newCartRepository(session),
		newCartProductRepository(session),
		newDiscountCodeRepository(session),
		newDiscountCodeUsageRepository(session),
		newOrderRepository(session),
	}
}

func (s Service) GetCart(cart *models.CartsStruct) error {
	return s.cartRepo.GetOne(cart)
}

func (s Service) CreateCartForUser(userId models.UUID) (*models.CartsStruct, error) {
	cart := models.CartsStruct{}
	cart.UserId = userId
	cart.CartId = models.RandomUUID()
	cart.DiscountCode = ""
	cart.DiscountPercent = 0

	if err := s.cartRepo.Insert(&cart); err != nil {
		return nil, err
	}

	return &cart, nil
}

func (s Service) AddToCart(cart *models.CartsStruct, product *models.CartProductsStruct) error {
	product.CartId = cart.CartId
	return s.cartProductRepo.Insert(product)
}

func (s Service) RemoveFromCart(product *models.CartProductsStruct) error {
	return s.cartProductRepo.Delete(product)
}

func (s Service) AddDiscountCodeToCart(cart *models.CartsStruct, code string) error {
	discountCode := models.DiscountCodesStruct{Code: code}
	if err := s.discountCodeRepo.GetOne(&discountCode); err != nil {
		return err
	}

	discountCodeUsage := models.DiscountCodeUsagesStruct{Code: discountCode.Code, UsageCount: 0}
	s.discountCodeUsageRepo.GetOne(&discountCodeUsage)
	if discountCodeUsage.UsageCount >= int(discountCode.UsageLimit) {
		return errors.New("discount code usage exceeded")
	}

	cart.DiscountCode = discountCode.Code
	cart.DiscountPercent = discountCode.DiscountPercent

	return s.cartRepo.Insert(cart)
}

func (s Service) RemoveDiscountCodeFromCart(cart *models.CartsStruct) error {
	cart.DiscountCode = ""
	cart.DiscountPercent = 0
	return s.cartRepo.Insert(cart)
}

func (s Service) GetCartProducts(cart *models.CartsStruct) (*[]models.CartProductsStruct, error) {
	productMatch := models.CartProductsStruct{CartId: cart.CartId}
	products, err := s.cartProductRepo.GetAll(&productMatch)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s Service) GetCartTotalPrice(cart *models.CartsStruct) (int, error) {
	totalPrice, err := s.cartProductRepo.SumProductPrices(cart.CartId)
	if err != nil {
		return 0, err
	}

	priceAfterDiscount := applyDiscount(totalPrice, int(cart.DiscountPercent))

	return priceAfterDiscount, nil
}

func (s Service) PlaceOrder(cart *models.CartsStruct) (*models.OrdersStruct, error) {
	products, err := s.GetCartProducts(cart)
	if err != nil {
		return nil, err
	}
	totalPrice := 0
	productsMap := map[[16]byte]models.OrderProductUserType{}
	for _, product := range *products {
		totalPrice += int(product.PriceInCents * product.Quantity)
		productsMap[product.ProductId] = models.OrderProductUserType{
			ProductId:    product.ProductId,
			Name:         product.Name,
			Description:  product.Description,
			PriceInCents: product.PriceInCents,
			Quantity:     product.Quantity,
		}
	}
	discountCode := models.DiscountCodesStruct{Code: cart.DiscountCode, DiscountPercent: 0}
	if cart.DiscountCode != "" {
		if err := s.discountCodeRepo.GetOne(&discountCode); err != nil {
			return nil, err
		}
		discountCodeUsage := models.DiscountCodeUsagesStruct{Code: discountCode.Code, UsageCount: 0}
		s.discountCodeUsageRepo.GetOne(&discountCodeUsage)
		if discountCodeUsage.UsageCount >= int(discountCode.UsageLimit) {
			return nil, errors.New("discount code usage exceeded")
		} else {
			s.discountCodeUsageRepo.IncreaseCodeUsageCount(&discountCodeUsage, 1)
		}
	}

	totalPrice = applyDiscount(totalPrice, int(discountCode.DiscountPercent))
	order := models.OrdersStruct{
		CartId:            cart.CartId,
		TotalPriceInCents: int32(totalPrice),
		Status:            "pending",
		Products:          productsMap,
	}

	err = s.orderRepo.Insert(&order)
	return &order, err
}

func applyDiscount(totalPriceInCents int, discountPercent int) int {
	priceAfterDiscount := float64(totalPriceInCents) * ((100 - float64(discountPercent)) / 100)
	return int(math.Ceil(priceAfterDiscount))
}
