package carts

import (
	"math"

	"github.com/Groszczu/gocass/internal/models"
	"github.com/scylladb/gocqlx/v2"
)

type Service struct {
	cartRepo         CartRepository
	cartProductRepo  CartProductRepository
	discountCodeRepo DiscountCodeRepository
	orderRepo        OrderRepository
}

func NewService(session *gocqlx.Session) Service {
	return Service{
		newCartRepository(session),
		newCartProductRepository(session),
		newDiscountCodeRepository(session),
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
	// TODO: Check if user already used the code

	cart.DiscountCode = discountCode.Code
	cart.DiscountPercent = discountCode.DiscountPercent

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

	priceAfterDiscount := float64(totalPrice) * ((100 - float64(cart.DiscountPercent)) / 100)
	priceRoundedUp := int(math.Ceil(priceAfterDiscount))

	return priceRoundedUp, nil
}
