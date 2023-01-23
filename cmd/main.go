package main

import (
	"log"
	"os"

	"github.com/Groszczu/gocass/internal/backend"
	"github.com/Groszczu/gocass/internal/carts"
	"github.com/Groszczu/gocass/internal/models"
	"github.com/Groszczu/gocass/internal/users"
)

var (
	dbHost = "127.0.0.1"
	dbPort = "9042"
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "gocass: ", log.Lmicroseconds)

	if dbHostEnv, ok := os.LookupEnv("DATABASE_HOST"); ok {
		logger.Println("Loaded DATABASE_HOST environment variable")
		dbHost = dbHostEnv
	} else {
		logger.Println("Using default database host value")
	}

	if dbPortEnv, ok := os.LookupEnv("DATABASE_PORT"); ok {
		logger.Println("Loaded DATABASE_PORT environment variable")
		dbPort = dbPortEnv
	} else {
		logger.Println("Using default database port value")
	}
}

func main() {

	session, err := backend.Session(dbHost + ":" + dbPort)

	if err != nil {
		logger.Fatalln("Failed to create backend session ", err)
	}

	defer func() {
		logger.Println("Closing backend session")
		session.Close()
		logger.Println("Backend session closed")
	}()

	logger.Printf("Created backed session on %s:%s", dbHost, dbPort)

	usersService := users.NewService(&session)
	cartsService := carts.NewService(&session)

	user := models.UsersStruct{Name: "roch"}
	if err := usersService.GetUser(&user); err != nil {
		if err := usersService.RegisterUser(&user); err != nil {
			logger.Fatalln("Failed to create a new user")
		}
	}

	cart := models.CartsStruct{UserId: user.Id}
	if err := cartsService.GetCart(&cart); err != nil {
		newCart, err := cartsService.CreateCartForUser(user.Id)
		if err != nil {
			logger.Fatalln("Failed to create a cart for user:", err)
		}
		cart = *newCart
	}

	logger.Printf("Got cart: %+v\n", cart)

	products := []models.CartProductsStruct{
		{
			Name:         "Cassandra guidebook",
			Description:  "Learn how to use Cassandra",
			Quantity:     3,
			PriceInCents: 40 * 100,
		},
		{
			Name:         "Gaming keyboard",
			Description:  "Best keyboard",
			Quantity:     2,
			PriceInCents: 120 * 100,
		},
	}

	for _, product := range products {
		product.CartId = cart.CartId
		product.ProductId = models.RandomUUID()
		cartsService.AddToCart(&cart, &product)
	}

	totalPrice, err := cartsService.GetCartTotalPrice(&cart)
	if err != nil {
		logger.Fatalln("Failed to get a total price of a cart:", err)
	}

	logger.Printf("Total cart price: %d\n", totalPrice)

	if err := cartsService.AddDiscountCodeToCart(&cart, "abc"); err != nil {
		logger.Fatalln("Failed to add discount code to a cart:", err)
	}

	totalPrice, err = cartsService.GetCartTotalPrice(&cart)
	if err != nil {
		logger.Fatalln("Failed to get a total price of a cart:", err)
	}

	logger.Printf("Total cart price: %d\n", totalPrice)
}
