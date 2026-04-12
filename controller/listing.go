package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type CreateListingAdminRequest struct {
	Listing model.Listing      `json:"listing"`
	SKUs    []model.ListingSKU `json:"skus"`
}

type UpdateListingStatusRequest struct {
	Status      string `json:"status"`
	AuditStatus string `json:"audit_status"`
	AuditRemark string `json:"audit_remark"`
}

type ListingAdminView struct {
	Listing model.Listing      `json:"listing"`
	SKUs    []model.ListingSKU `json:"skus"`
}

func GetListingAdmin(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	sellerId, _ := strconv.Atoi(c.Query("seller_id"))
	listings, total, err := service.ListListings(
		c.Query("keyword"),
		c.Query("status"),
		c.Query("audit_status"),
		sellerId,
		pageInfo.GetStartIdx(),
		pageInfo.GetPageSize(),
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	items := make([]ListingAdminView, 0, len(listings))
	for _, listing := range listings {
		skus, skuErr := model.GetListingSKUsByListingID(listing.Id)
		if skuErr != nil {
			common.ApiError(c, skuErr)
			return
		}
		items = append(items, ListingAdminView{
			Listing: *listing,
			SKUs:    skus,
		})
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(items)
	common.ApiSuccess(c, pageInfo)
}

func CreateListingAdmin(c *gin.Context) {
	var req CreateListingAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	listing, skus, err := service.CreateListingWithSKUs(service.CreateListingInput{
		Listing: req.Listing,
		SKUs:    req.SKUs,
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, ListingAdminView{
		Listing: *listing,
		SKUs:    skus,
	})
}

func UpdateListingAdminStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	var req UpdateListingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := service.UpdateListingStatus(id, req.Status, req.AuditStatus, req.AuditRemark); err != nil {
		common.ApiError(c, err)
		return
	}
	listing, err := model.GetListingByID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	skus, err := model.GetListingSKUsByListingID(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, ListingAdminView{
		Listing: *listing,
		SKUs:    skus,
	})
}
