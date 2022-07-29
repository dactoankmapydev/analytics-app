package website

import (
	"analytics-api/configs"
	"analytics-api/internal/app/auth"
	dur "analytics-api/internal/pkg/duration"
	"analytics-api/internal/pkg/middleware"
	"analytics-api/internal/pkg/security"
	str "analytics-api/internal/pkg/string"
	"analytics-api/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type httpDelivery struct {
	websiteUseCase UseCase
	authUsecase    auth.UseCase
}

var validate = validator.New()

type RequestWebsite struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	URL  string `json:"url" validate:"required,min=3"`
}

// InitRoutes ...
func (instance *httpDelivery) InitRoutes(r *gin.RouterGroup) {
	websiteRoutes := r.Group("/website")
	{
		websiteRoutes.GET("/:website_id", middleware.JWTMiddleware(), instance.GetWebsite)
		websiteRoutes.GET("/list", middleware.JWTMiddleware(), instance.GetAllWebsite)
		websiteRoutes.GET("/tracking/:website_id", middleware.JWTMiddleware(), instance.Tracking)
		websiteRoutes.POST("/add", middleware.JWTMiddleware(), instance.AddWebsite)
		websiteRoutes.PUT("/update", middleware.JWTMiddleware(), instance.UpdateWebsite)
		websiteRoutes.DELETE("/delete", middleware.JWTMiddleware(), instance.DeleteWebsite)
	}
}

// Tracking guide tracking code to website
func (instance *httpDelivery) Tracking(c *gin.Context) {
	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "extract token metadata failed"})
		return
	}

	userID, err := instance.authUsecase.GetAuth(tokenAuth.AccessUUID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "get token auth failed"})
		return
	}

	websiteID := c.Param("website_id")

	c.HTML(http.StatusOK, "tracking.html", gin.H{
		"URL":       configs.AppURL,
		"WebsiteID": websiteID,
		"UserID":    userID,
	})
}

func (instance *httpDelivery) GetWebsite(c *gin.Context) {
	websiteID := c.Param("website_id")
	var website models.Website
	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "extract token metadata failed"})
		return
	}

	userID, err := instance.authUsecase.GetAuth(tokenAuth.AccessUUID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "get token auth failed"})
		return
	}

	getWebsiteErr := instance.websiteUseCase.GetWebsite(userID, websiteID, &website)
	if getWebsiteErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while get the website by id"})
		return
	}

	c.JSON(http.StatusOK, website)
}

func (instance *httpDelivery) GetAllWebsite(c *gin.Context) {
	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "extract token metadata failed"})
		return
	}

	userID, err := instance.authUsecase.GetAuth(tokenAuth.AccessUUID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "get token auth failed"})
		return
	}

	websites, err := instance.websiteUseCase.GetAllWebsite(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while get all website"})
		return
	}

	c.JSON(http.StatusOK, websites)
}

func (instance *httpDelivery) AddWebsite(c *gin.Context) {
	var request RequestWebsite
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validationErr := validate.Struct(request)
	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	hostName, err := str.ParseURL(request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "error occured while parse url the website"})
		return
	}

	tokenAuth, err := security.ExtractAccessTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "extract token metadata failed"})
		return
	}

	userID, err := instance.authUsecase.GetAuth(tokenAuth.AccessUUID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "get token auth failed"})
		return
	}

	count, err := instance.websiteUseCase.FindWebsite(userID, hostName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "error occured while check for the email"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"msg": "this website already exists"})
		return
	} else {

		websiteID := security.Hash(hostName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "error occured while create id for the website"})
			return
		}

		createdAt, err := dur.ParseTime(time.Now().Format("2006-01-02, 15:04:05"))
		if err != nil {
			logrus.Error(c, err)
			return
		}

		website := models.Website{
			ID:        websiteID,
			UserID:    userID,
			Name:      request.Name,
			HostName:  hostName,
			URL:       request.URL,
			Tracked:   false,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}

		insertErr := instance.websiteUseCase.InsertWebsite(userID, website)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": "error occured while insert website"})
			return
		}

		c.JSON(http.StatusOK, website)
	}
}

func (instance *httpDelivery) UpdateWebsite(c *gin.Context) {
}

func (instance *httpDelivery) DeleteWebsite(c *gin.Context) {
}
