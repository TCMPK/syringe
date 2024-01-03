package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/santosh/gingo/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type DomainDefinition struct {
	Domain string `json:"domain" example:"google.com"`
	Type   string `json:"type" example:"A"`
}
type DomainListDefinition struct {
	Domains []DomainDefinition `json:"domains"`
}

type Response struct {
	Message string `json:"message" example:"success"`
}

type ResponseError struct {
	Message string `json:"message" example:"error <error msg here>"`
}

type ResponseWithDomains struct {
	Message string             `json:"message" example:"success"`
	Domains []DomainDefinition `json:"domains"`
}

type ResponseWithSize struct {
	Message string `json:"message" example:"success"`
	Size    int    `json:"size" example:"1"`
}

// HandleLoadRandomDomains godoc
// @Summary     Load random domains from the configured domains file
// @Description Responds with the new queue size
// @Param 		count  		query 		int 	false 	"int valid"		minimum(1) example(10)
// @Tags        syringe
// @Produce     json
// @Success     200  {object}  main.ResponseWithSize
// @Failure     412  {object}  main.ResponseError
// @Router      /domains/random [post]
func HandleLoadRandomDomains(c *gin.Context, dh *DomainHeap) {
	qps_aim := 50
	query_param_count := c.Query("count")
	count, err := strconv.Atoi(query_param_count)
	if query_param_count == "" {
		count = 1
	} else if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"message": "query argument 'count' must be a positive integer",
		})
		return
	} else {
		if count < 1 {
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"message": "query argument 'count' must be a positive integer > 0",
			})
			return
		}
		bulk_slot_size := count / qps_aim
		if bulk_slot_size == 0 {
			bulk_slot_size = 1
		}
		for i := 0; i < count; i++ {
			dh.AppendRandom(uint(i % bulk_slot_size))
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"size":    dh.Len(),
		})
	}
}

// HandleDumpDomains godoc
// @Summary      Return a list of domains currently in the queue
// @Description  Responds with the queue
// @Tags         syringe
// @Produce      json
// @Success      200  {object}  main.ResponseWithDomains
// @Router       /domains [get]
func HandleDumpDomains(c *gin.Context, dh *DomainHeap) {
	var domainList []DomainDefinition
	for _, d := range *dh {
		domainList = append(domainList, DomainDefinition{Domain: d.Record_name, Type: d.Record_type})
	}
	c.JSON(http.StatusOK, ResponseWithDomains{Domains: domainList})
}

// HandleCountDomains godoc
// @Summary      Return the number of domains in the queue
// @Description  Responds with the queue size
// @Tags         syringe
// @Produce      json
// @Success      200  {object}  main.ResponseWithSize
// @Router       /domains/count [get]
func HandleCountDomains(c *gin.Context, dh *DomainHeap) {
	c.JSON(http.StatusOK, ResponseWithSize{Message: "success", Size: dh.Len()})
}

// HandleAddDomains godoc
// @Summary     Load domains into the queue
// @Description Responds with the new queue size
// @Param 		body body main.DomainListDefinition true "domain list"
// @Tags        syringe
// @Produce     json
// @Success     200  {object}  main.ResponseWithSize
// @Failure     400  {object}  main.ResponseError
// @Router      /domains [post]
func HandleAddDomains(c *gin.Context, dh *DomainHeap) {
	requestBody := &DomainListDefinition{}

	if err := c.ShouldBindJSON(requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": `invalid domain list received. example: {"domains":[{"domain":"google.de","type":"A"}]}`,
		})
		return
	}
	var domainList []Domain
	for i := 0; i < len(requestBody.Domains); i++ {
		// we need to validate the input before we push it onto the heap
		domain := Domain{Record_name: requestBody.Domains[i].Domain, Record_type: requestBody.Domains[i].Type, Refresh_at: 0}
		if !domain.Validate() {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": fmt.Sprintf("invalid domain in domain list. Unknown type %s for domain %s", requestBody.Domains[i].Type, requestBody.Domains[i].Domain),
			})
			return
		}
		domainList = append(domainList, domain)
	}
	// append after validating
	for _, d := range domainList {
		HeapPush(dh, d)
	}
	c.JSON(http.StatusOK, ResponseWithSize{Message: "success", Size: dh.Len()})
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/domains", func(c *gin.Context) {
			HandleDumpDomains(c, dh)
		})
		v1.GET("/domains/count", func(c *gin.Context) {
			HandleCountDomains(c, dh)
		})
		v1.POST("/domains/random", func(c *gin.Context) {
			HandleLoadRandomDomains(c, dh)
		})
		v1.POST("/domains", func(c *gin.Context) {
			HandleAddDomains(c, dh)
		})
	}

	return router
}

func DocOverrideHandler(c *gin.Context) {
	if strings.HasSuffix(c.Request.URL.Path, "/doc.json") {
		location := url.URL{Path: "/swagger-static/doc.json"}
		c.Redirect(http.StatusFound, location.RequestURI())

	} else {
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	}
}
