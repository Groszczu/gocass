// Code generated by "gocqlx/cmd/schemagen"; DO NOT EDIT.

package models

import (
	"github.com/scylladb/gocqlx/v2/table"
)

// Table models.
var (
	CartProducts = table.New(table.Metadata{
		Name: "cart_products",
		Columns: []string{
			"cart_id",
			"description",
			"name",
			"price_in_cents",
			"product_id",
			"quantity",
		},
		PartKey: []string{
			"cart_id",
		},
		SortKey: []string{
			"product_id",
		},
	})

	Carts = table.New(table.Metadata{
		Name: "carts",
		Columns: []string{
			"cart_id",
			"discount_code",
			"discount_percent",
			"user_id",
		},
		PartKey: []string{
			"user_id",
		},
		SortKey: []string{
			"cart_id",
		},
	})

	DiscountCodeUsages = table.New(table.Metadata{
		Name: "discount_code_usages",
		Columns: []string{
			"code",
			"usage_count",
		},
		PartKey: []string{
			"code",
		},
		SortKey: []string{},
	})

	DiscountCodes = table.New(table.Metadata{
		Name: "discount_codes",
		Columns: []string{
			"code",
			"discount_percent",
			"usage_limit",
			"used_by",
		},
		PartKey: []string{
			"code",
		},
		SortKey: []string{},
	})

	Users = table.New(table.Metadata{
		Name: "users",
		Columns: []string{
			"id",
			"name",
		},
		PartKey: []string{
			"id",
		},
		SortKey: []string{},
	})

	Orders = table.New(table.Metadata{
		Name: "orders",
		Columns: []string{
			"cart_id",
			"total_price_in_cents",
			"products",
			"status",
		},
		PartKey: []string{
			"cart_id",
		},
		SortKey: []string{},
	})
)

type OrdersStruct struct {
	CartId [16]byte
	TotalPriceInCents int32
	Products map[[16]byte]OrderProductUserType
	Status string
}

type OrderProductUserType struct {
	ProductId    [16]byte
	Name         string
	Description  string
	PriceInCents int32
	Quantity     int32
}

type CartProductsStruct struct {
	CartId       [16]byte
	Description  string
	Name         string
	PriceInCents int32
	ProductId    [16]byte
	Quantity     int32
}
type CartsStruct struct {
	CartId          [16]byte
	DiscountCode    string
	DiscountPercent int32
	UserId          [16]byte
}
type DiscountCodeUsagesStruct struct {
	Code       string
	UsageCount int
}
type DiscountCodesStruct struct {
	Code            string
	DiscountPercent int32
	UsageLimit      int32
	UsedBy          [][16]byte
}
type UsersStruct struct {
	Id   [16]byte
	Name string
}