package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gaokaohub/payment-service/internal/middleware"
	"github.com/gaokaohub/payment-service/internal/models"
	"github.com/gaokaohub/payment-service/internal/services"
)

// OrderHandler 订单处理器
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建会员订单
// @Tags order
// @Accept json
// @Produce json
// @Param request body models.CreateOrderRequest true "订单请求"
// @Success 200 {object} models.CreateOrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/create [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "INVALID_REQUEST",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 创建订单
	resp, err := h.orderService.CreateOrder(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "CREATE_ORDER_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrders 获取订单列表
// @Summary 获取订单列表
// @Description 获取用户的订单列表
// @Tags order
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param status query string false "订单状态"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} models.OrderListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/list [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 解析查询参数
	req := models.OrderListRequest{
		Page:     1,
		PageSize: 10,
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			req.PageSize = ps
		}
	}

	req.Status = c.Query("status")
	req.StartTime = c.Query("start_time")
	req.EndTime = c.Query("end_time")

	// 获取订单列表
	resp, err := h.orderService.GetOrders(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_ORDERS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrder 获取单个订单
// @Summary 获取订单详情
// @Description 获取指定订单的详细信息
// @Tags order
// @Accept json
// @Produce json
// @Param orderNo path string true "订单号"
// @Success 200 {object} models.PaymentOrder
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{orderNo} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing order number",
			Code:  "MISSING_ORDER_NO",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 获取订单
	order, err := h.orderService.GetOrder(c.Request.Context(), userID, orderNo)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: err.Error(),
				Code:  "ORDER_NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_ORDER_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Description 取消未支付的订单
// @Tags order
// @Accept json
// @Produce json
// @Param orderNo path string true "订单号"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{orderNo}/cancel [put]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing order number",
			Code:  "MISSING_ORDER_NO",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 取消订单
	err := h.orderService.CancelOrder(c.Request.Context(), userID, orderNo)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: err.Error(),
				Code:  "ORDER_NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "CANCEL_ORDER_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "order canceled successfully",
	})
}

// GetInvoice 获取发票
// @Summary 获取发票
// @Description 获取已支付订单的发票信息
// @Tags order
// @Accept json
// @Produce json
// @Param orderNo path string true "订单号"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{orderNo}/invoice [get]
func (h *OrderHandler) GetInvoice(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "missing order number",
			Code:  "MISSING_ORDER_NO",
		})
		return
	}

	// 获取用户ID
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "user not authenticated",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	// 获取发票
	invoice, err := h.orderService.GetInvoice(c.Request.Context(), userID, orderNo)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: err.Error(),
				Code:  "ORDER_NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "GET_INVOICE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, invoice)
}