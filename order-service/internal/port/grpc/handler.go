package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/Abdurahmanit/GroupProject/order-service/internal/platform/logger"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/repository"
	"github.com/Abdurahmanit/GroupProject/order-service/internal/service"
	cartpb "github.com/Abdurahmanit/GroupProject/order-service/proto/cart"
	commonpb "github.com/Abdurahmanit/GroupProject/order-service/proto/common"
	orderpb "github.com/Abdurahmanit/GroupProject/order-service/proto/order"
	orderservicepb "github.com/Abdurahmanit/GroupProject/order-service/proto/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderGRPCHandler struct {
	orderservicepb.UnimplementedOrderServiceServer
	cartService    service.CartService
	orderService   service.OrderService
	receiptService service.ReceiptService
	log            logger.Logger
}

func NewOrderGRPCHandler(
	cartService service.CartService,
	orderService service.OrderService,
	receiptService service.ReceiptService,
	log logger.Logger,
) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		cartService:    cartService,
		orderService:   orderService,
		receiptService: receiptService,
		log:            log,
	}
}

func (h *OrderGRPCHandler) AddItemToCart(ctx context.Context, req *orderservicepb.AddItemToCartRequest) (*cartpb.CartProto, error) {
	cartProto, err := h.cartService.AddItem(ctx, req.GetUserId(), req.GetProductId(), int(req.GetQuantity()))
	if err != nil {
		h.log.Errorf("AddItemToCart failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to add item to cart: %v", err)
	}
	return cartProto, nil
}

func (h *OrderGRPCHandler) UpdateCartItemQuantity(ctx context.Context, req *orderservicepb.UpdateCartItemQuantityRequest) (*cartpb.CartProto, error) {
	cartProto, err := h.cartService.UpdateItemQuantity(ctx, req.GetUserId(), req.GetProductId(), int(req.GetNewQuantity()))
	if err != nil {
		h.log.Errorf("UpdateCartItemQuantity failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update item quantity: %v", err)
	}
	return cartProto, nil
}

func (h *OrderGRPCHandler) RemoveItemFromCart(ctx context.Context, req *orderservicepb.RemoveItemFromCartRequest) (*cartpb.CartProto, error) {
	cartProto, err := h.cartService.RemoveItem(ctx, req.GetUserId(), req.GetProductId())
	if err != nil {
		h.log.Errorf("RemoveItemFromCart failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to remove item from cart: %v", err)
	}
	return cartProto, nil
}

func (h *OrderGRPCHandler) GetCart(ctx context.Context, req *orderservicepb.GetCartRequest) (*cartpb.CartProto, error) {
	cartProto, err := h.cartService.GetCart(ctx, req.GetUserId())
	if err != nil {
		h.log.Errorf("GetCart failed: %v", err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "cart not found for user %s", req.GetUserId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}
	return cartProto, nil
}

func (h *OrderGRPCHandler) ClearCart(ctx context.Context, req *orderservicepb.ClearCartRequest) (*emptypb.Empty, error) {
	err := h.cartService.ClearCart(ctx, req.GetUserId())
	if err != nil {
		h.log.Errorf("ClearCart failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to clear cart: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (h *OrderGRPCHandler) PlaceOrder(ctx context.Context, req *orderservicepb.PlaceOrderRequest) (*orderpb.OrderProto, error) {
	orderProto, err := h.orderService.PlaceOrder(ctx, req.GetUserId(), req.GetShippingAddress(), req.GetBillingAddress())
	if err != nil {
		h.log.Errorf("PlaceOrder failed: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to place order: %v", err)
	}
	return orderProto, nil
}

func (h *OrderGRPCHandler) GetOrder(ctx context.Context, req *orderservicepb.GetOrderRequest) (*orderpb.OrderProto, error) {
	userIDFromAuth := ""
	isAdminFromAuth := false

	orderProto, err := h.orderService.GetOrderByID(ctx, req.GetOrderId(), userIDFromAuth, isAdminFromAuth)
	if err != nil {
		h.log.Errorf("GetOrder failed for orderID %s: %v", req.GetOrderId(), err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "order %s not found", req.GetOrderId())
		}
		if err.Error() == fmt.Sprintf("access denied to order %s", req.GetOrderId()) {
			return nil, status.Errorf(codes.PermissionDenied, "access denied to order %s", req.GetOrderId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}
	return orderProto, nil
}

func (h *OrderGRPCHandler) ListUserOrders(ctx context.Context, req *orderservicepb.ListUserOrdersRequest) (*orderservicepb.ListUserOrdersResponse, error) {
	orders, total, err := h.orderService.ListUserOrders(ctx, req.GetUserId(), req.GetPagination())
	if err != nil {
		h.log.Errorf("ListUserOrders failed for userID %s: %v", req.GetUserId(), err)
		return nil, status.Errorf(codes.Internal, "failed to list user orders: %v", err)
	}

	var totalPages int32
	if req.GetPagination().GetPageSize() > 0 {
		totalPages = int32((total + int64(req.GetPagination().GetPageSize()) - 1) / int64(req.GetPagination().GetPageSize()))
	} else if total > 0 {
		totalPages = 1
	}

	return &orderservicepb.ListUserOrdersResponse{
		Orders: orders,
		Pagination: &commonpb.PaginationResponse{
			TotalItems:  total,
			CurrentPage: req.GetPagination().GetPage(),
			PageSize:    req.GetPagination().GetPageSize(),
			TotalPages:  totalPages,
		},
	}, nil
}

func (h *OrderGRPCHandler) CancelOrder(ctx context.Context, req *orderservicepb.CancelOrderRequest) (*orderpb.OrderProto, error) {
	orderProto, err := h.orderService.CancelUserOrder(ctx, req.GetOrderId(), req.GetUserId())
	if err != nil {
		h.log.Errorf("CancelOrder failed for orderID %s by userID %s: %v", req.GetOrderId(), req.GetUserId(), err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "order %s not found", req.GetOrderId())
		}
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}
	return orderProto, nil
}

func (h *OrderGRPCHandler) UpdateOrderStatus(ctx context.Context, req *orderservicepb.UpdateOrderStatusRequest) (*orderpb.OrderProto, error) {
	orderProto, err := h.orderService.UpdateOrderStatusByAdmin(ctx, req.GetOrderId(), req.GetNewStatus(), req.GetUpdatedById())
	if err != nil {
		h.log.Errorf("UpdateOrderStatus failed for orderID %s by adminID %s: %v", req.GetOrderId(), req.GetUpdatedById(), err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "order %s not found", req.GetOrderId())
		}
		return nil, status.Errorf(codes.Internal, "failed to update order status: %v", err)
	}
	return orderProto, nil
}

func (h *OrderGRPCHandler) ListAllOrders(ctx context.Context, req *orderservicepb.ListAllOrdersAdminRequest) (*orderservicepb.ListAllOrdersAdminResponse, error) {
	filters := make(map[string]string)

	orders, total, err := h.orderService.ListAllOrdersAdmin(ctx, req.GetAdminId(), req.GetPagination(), filters)
	if err != nil {
		h.log.Errorf("ListAllOrders failed for adminID %s: %v", req.GetAdminId(), err)
		return nil, status.Errorf(codes.Internal, "failed to list all orders: %v", err)
	}

	var totalPages int32
	if req.GetPagination().GetPageSize() > 0 {
		totalPages = int32((total + int64(req.GetPagination().GetPageSize()) - 1) / int64(req.GetPagination().GetPageSize()))
	} else if total > 0 {
		totalPages = 1
	}

	return &orderservicepb.ListAllOrdersAdminResponse{
		Orders: orders,
		Pagination: &commonpb.PaginationResponse{
			TotalItems:  total,
			CurrentPage: req.GetPagination().GetPage(),
			PageSize:    req.GetPagination().GetPageSize(),
			TotalPages:  totalPages,
		},
	}, nil
}

func (h *OrderGRPCHandler) GenerateOrderReceipt(ctx context.Context, req *orderservicepb.GenerateOrderReceiptRequest) (*orderservicepb.GenerateOrderReceiptResponse, error) {
	pdfBytes, fileName, err := h.receiptService.GenerateOrderReceiptPDF(ctx, req.GetOrderId(), req.GetUserId())
	if err != nil {
		h.log.Errorf("GenerateOrderReceipt failed for orderID %s by userID %s: %v", req.GetOrderId(), req.GetUserId(), err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "order %s not found", req.GetOrderId())
		}
		if err.Error() == "PDF generation is not implemented yet" {
			return nil, status.Errorf(codes.Unimplemented, "PDF generation is not yet available")
		}
		return nil, status.Errorf(codes.Internal, "failed to generate order receipt: %v", err)
	}
	return &orderservicepb.GenerateOrderReceiptResponse{
		PdfContent: pdfBytes,
		FileName:   fileName,
	}, nil
}
