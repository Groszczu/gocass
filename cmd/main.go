package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Groszczu/gocass/internal/backend"
	"github.com/Groszczu/gocass/internal/carts"
	"github.com/Groszczu/gocass/internal/models"
	"github.com/Groszczu/gocass/internal/users"
	"github.com/gocql/gocql"
)

var (
	logger  *log.Logger
	dbHosts = flag.String("hosts", "127.0.0.1:9042,127.0.0.1:9043,127.0.0.1:9044", "comma-separated hosts addresses")
	workers = flag.Int("workers", 3, "number of workers to use")
	runs    = flag.Int("runs", 100, "number of runs per worker")
)

func init() {
	flag.Parse()

	logger = log.New(os.Stdout, "gocass: ", log.Lmicroseconds)

	logger.Printf("Cassandra database hosts: %s\n", *dbHosts)
	logger.Printf("Number of workers: %d\n", *workers)
	logger.Printf("Number of runs: %d\n", *runs)
}

func main() {
	work := make(chan map[string]int)

	for workerId := 0; workerId < *workers; workerId++ {
		go runWorker(work, workerId, *runs)
	}

	codesUsage := map[string]int{}
	for _, code := range discountCodes {
		codesUsage[code] = 0
	}

	for workerId := 0; workerId < *workers; workerId++ {
		workerCodesUsage := <-work
		for code, usage := range workerCodesUsage {
			codesUsage[code] += usage
		}
	}

	logger.Printf("Code usage: %+v\n", codesUsage)
}

func parseUUID(input string) gocql.UUID {
	uuid, err := gocql.ParseUUID(input)
	if err != nil {
		panic("Invalid UUID")
	}
	return uuid
}

func randomInt(min int, max int) int {
	return rand.Intn(max-min+1) + min
}
func randomQuantity() int32 {
	return int32(randomInt(1, 5))
}

func randomSleep() {
	time.Sleep(time.Duration(randomInt(20, 100)) * time.Millisecond)
}

var products = []models.CartProductsStruct{
	{
		ProductId:    parseUUID("c1433f31-ea57-4f16-9555-dc616169a6d2"),
		Name:         "Cassandra guidebook",
		Description:  "Learn how to use Cassandra",
		Quantity:     1,
		PriceInCents: 40 * 100,
	},
	{
		ProductId:    parseUUID("64efbc98-63e1-404d-9789-5ccb307c7d7e"),
		Name:         "Gaming keyboard",
		Description:  "Best keyboard",
		Quantity:     1,
		PriceInCents: 120 * 100,
	},
	{
		ProductId:    parseUUID("d5b384d9-cf7f-4383-b59e-0558fa6b79c1"),
		Name:         "Smartphone",
		Description:  "Best smartphone",
		Quantity:     1,
		PriceInCents: 300 * 100,
	},
	{
		ProductId:    parseUUID("89c7b039-1481-4ef8-b5ad-83c227b83540"),
		Name:         "Headphones",
		Description:  "Best headphones",
		Quantity:     1,
		PriceInCents: 50 * 100,
	},
}

var discountCodes = []string{
	"abc",
	"def",
}

var user = models.UsersStruct{
	Id:   parseUUID("c372e753-2624-430a-95c7-f2e84e0415cb"),
	Name: "user",
}

func runWorker(work chan map[string]int, workerId int, runs int) {
	workerLogger := log.New(os.Stdout, fmt.Sprintf("worker-%d: ", workerId), log.Lmicroseconds)

	session, err := backend.Session(strings.Split(*dbHosts, ",")...)

	if err != nil {
		workerLogger.Fatalln("Failed to create backend session ", err)
	}

	defer func() {
		workerLogger.Println("Closing backend session")
		session.Close()
		workerLogger.Println("Backend session closed")
	}()

	workerLogger.Printf("Created backend session on %s", *dbHosts)

	codesUsage := map[string]int{}
	for _, code := range discountCodes {
		codesUsage[code] = 0
	}

	usersService := users.NewService(&session)
	cartsService := carts.NewService(&session)

	for runId := 0; runId < runs; runId++ {
		if err := usersService.GetUser(&user); err != nil {
			workerLogger.Printf("No user with name '%s' found, registering new user.", user.Name)
			if err := usersService.RegisterUser(&user); err != nil {
				workerLogger.Println("Failed to create a new user:", err)
				continue
			}
		}

		cart := models.CartsStruct{UserId: user.Id}
		if err := cartsService.GetCart(&cart); err != nil {
			workerLogger.Println("No single cart found, creating new cart")
			newCart, err := cartsService.CreateCartForUser(user.Id)
			if err != nil {
				workerLogger.Println("Failed to create a cart for user:", err)
				continue
			}
			workerLogger.Println("New cart created")
			cart = *newCart
		}

		numberOfProductsToAdd := randomInt(1, len(products))
		addedProducts := []models.CartProductsStruct{}
		addedProductsPrice := 0
		for i := 0; i < numberOfProductsToAdd; i++ {
			product := products[randomInt(0, len(products)-1)]
			product.Quantity = randomQuantity()
			addedProducts = append(addedProducts, product)
			addedProductsPrice += int(product.PriceInCents * product.Quantity)
			cartsService.AddToCart(&cart, &product)
			randomSleep()
		}

		workerLogger.Printf("Added products worth: %d\n", addedProductsPrice)

		discountCodeIndex := randomInt(0, len(discountCodes))
		if discountCodeIndex != len(discountCodes) {
			discountCode := discountCodes[discountCodeIndex]
			if err := cartsService.AddDiscountCodeToCart(&cart, discountCode); err != nil {
				workerLogger.Printf("Failed to add discount code '%s' to a cart: %s\n", discountCode, err)
			} else {
				workerLogger.Printf("Added discount code '%s' to cart\n", discountCode)
				codesUsage[discountCode] += 1
			}
		}

		order, err := cartsService.PlaceOrder(&cart)
		if err != nil {
			workerLogger.Println("Failed to place order:", err)
			if err := cartsService.RemoveDiscountCodeFromCart(&cart); err != nil {
				workerLogger.Println("Failed to remove disount code from cart:", err)
			} else {
				order, err := cartsService.PlaceOrder(&cart)
				if err != nil {
					workerLogger.Println("Failed to place order:", err)
				} else {
					workerLogger.Printf("Placed order for: %d\n", order.TotalPriceInCents)
				}
			}
		} else {
			workerLogger.Printf("Placed order for: %d\n", order.TotalPriceInCents)
		}
	}

	workerLogger.Printf("Codes used: %+v\n", codesUsage)
	work <- codesUsage
}
