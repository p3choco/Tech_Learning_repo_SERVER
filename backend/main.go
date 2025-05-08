package main

import (
	"log"
    "net/http"
    "strconv"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)


const ProductByIDPath = "/products/:id"
const ProductNotFoundStr = "Product not found"

var DB *gorm.DB

type Category struct {
    gorm.Model
    Name     string    `json:"name"`
    Products []Product `json:"products"`
}

type Product struct {
    gorm.Model
    Name       string   `json:"name"`
    Price      float64  `json:"price"`
    CategoryID uint     `json:"category_id"`
    Category   Category `json:"category"`
}

type Cart struct {
    gorm.Model
    UserID    uint    `json:"user_id"`
    CartValue float64 `json:"cart_value"`
}


type Payment struct {
    gorm.Model
    CustomerName  string        `json:"name"`
    CustomerEmail string        `json:"email"`
    Total         float64       `json:"total"`
    Items         []PaymentItem `json:"items" gorm:"foreignKey:PaymentID"`
}

type PaymentItem struct {
    gorm.Model
    PaymentID uint    `json:"-"`
    ProductID uint    `json:"product_id"`
    Name      string  `json:"name"`
    Price     float64 `json:"price"`
    Qty       uint    `json:"qty"`
}

func initDB() {
    var err error
    DB, err = gorm.Open(sqlite.Open("example.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }
    if err := DB.AutoMigrate(&Category{}, &Product{}, &Cart{}, &Payment{}, &PaymentItem{}); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}	
}

func main() {
    initDB()

    e := echo.New()
    e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
        AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
    }))

    e.GET("/products", getProducts)
    e.GET(ProductByIDPath, getProductByID)
    e.POST("/products", createProduct)
    e.PUT(ProductByIDPath, updateProduct)
    e.DELETE(ProductByIDPath, deleteProduct)
    e.GET("/products/filter", filterProducts)

    e.GET("/carts", getCarts)
    e.POST("/carts", createCart)

    e.GET("/categories", getCategories)
    e.GET("/categories/:id", getCategoryByID)
    e.POST("/categories", createCategory)

    e.POST("/payments", createPayment)

    e.Logger.Fatal(e.Start(":8080"))
}


func getProducts(c echo.Context) error {
    var products []Product
    if err := DB.Preload("Category").Find(&products).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, products)
}

func getProductByID(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var product Product
    if err := DB.Preload("Category").First(&product, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, echo.Map{"message": ProductNotFoundStr})
    }
    return c.JSON(http.StatusOK, product)
}

func createProduct(c echo.Context) error {
    var newProduct Product
    if err := c.Bind(&newProduct); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }
    if err := DB.Create(&newProduct).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, newProduct)
}

func updateProduct(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var product Product
    if err := DB.First(&product, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, echo.Map{"message": ProductNotFoundStr})
    }
    if err := c.Bind(&product); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }
    if err := DB.Save(&product).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, product)
}

func deleteProduct(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var product Product
    if err := DB.First(&product, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, echo.Map{"message": ProductNotFoundStr})
    }
    if err := DB.Delete(&product).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.NoContent(http.StatusNoContent)
}

func filterProducts(c echo.Context) error {
    var products []Product
    minPriceStr := c.QueryParam("minPrice")
    categoryIDStr := c.QueryParam("categoryID")

    dbQuery := DB.Preload("Category").Model(&Product{})
    if minPriceStr != "" {
        if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
            dbQuery = dbQuery.Scopes(ScopeMinPrice(minPrice))
        }
    }
    if categoryIDStr != "" {
        if catID, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
            dbQuery = dbQuery.Scopes(ScopeCategoryID(uint(catID)))
        }
    }
    if err := dbQuery.Find(&products).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, products)
}


func getCarts(c echo.Context) error {
    var carts []Cart
    if err := DB.Find(&carts).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, carts)
}

func createCart(c echo.Context) error {
    var cart Cart
    if err := c.Bind(&cart); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }
    if err := DB.Create(&cart).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, cart)
}


func getCategories(c echo.Context) error {
    var categories []Category
    if err := DB.Preload("Products").Find(&categories).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, categories)
}

func getCategoryByID(c echo.Context) error {
    id, _ := strconv.Atoi(c.Param("id"))
    var category Category
    if err := DB.Preload("Products").First(&category, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, echo.Map{"message": "Category not found"})
    }
    return c.JSON(http.StatusOK, category)
}

func createCategory(c echo.Context) error {
    var category Category
    if err := c.Bind(&category); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }
    if err := DB.Create(&category).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, category)
}


func createPayment(c echo.Context) error {
    var payload struct {
        Customer struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        } `json:"customer"`
        Items []struct {
            ID    uint    `json:"product_id"`
            Name  string  `json:"name"`
            Price float64 `json:"price"`
            Qty   uint    `json:"qty"`
        } `json:"items"`
    }

    if err := c.Bind(&payload); err != nil {
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Niepoprawny format danych"})
    }

    var total float64
    for _, it := range payload.Items {
        total += it.Price * float64(it.Qty)
    }

    payment := Payment{
        CustomerName:  payload.Customer.Name,
        CustomerEmail: payload.Customer.Email,
        Total:         total,
    }
    for _, it := range payload.Items {
        payment.Items = append(payment.Items, PaymentItem{
            ProductID: it.ID,
            Name:      it.Name,
            Price:     it.Price,
            Qty:       it.Qty,
        })
    }

    if err := DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&payment).Error; err != nil {
            return err
        }
        return nil
    }); err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    return c.JSON(http.StatusCreated, echo.Map{
        "message": "Płatność przyjęta",
        "payment": payment,
    })
}


func ScopeMinPrice(min float64) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("price >= ?", min)
    }
}

func ScopeCategoryID(catID uint) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("category_id = ?", catID)
    }
}
