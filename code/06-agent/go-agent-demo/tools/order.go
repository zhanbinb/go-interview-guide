package tools

import (
	"encoding/json"
	"fmt"
)

// QueryOrder 查询订单信息
func QueryOrder(orderID string) string {
	return fmt.Sprintf(
		"订单 %s 已发货，金额3999元，预计明天送达",
		orderID,
	)
}

// GetUser 查询用户信息
func GetUser(userID string) string {
	return fmt.Sprintf(
		"用户 %s 张三,VIP用户，账户状态正常",
		userID,
	)
}

// QueryPayment 查询订单支付信息
func QueryPayment(orderID string) string {
	return fmt.Sprintf(
		"订单 %s 已支付，支付金额3999元，支付方式为微信支付",
		orderID,
	)
}

// QueryLogistics 查询订单物流信息
func QueryLogistics(orderID string) string {
	return fmt.Sprintf(
		"订单 %s 已发货，物流单号为123456",
		orderID,
	)
}

// QueryOrderTool 创建查询订单 Tool
func QueryOrderTool() Tool {
	return Tool{
		Name:        "query_order",
		Description: "查询订单信息",
		Handler: func(args string) (string, error) {
			var params struct {
				OrderID string `json:"order_id"`
			}

			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			return QueryOrder(params.OrderID), nil
		},
	}
}

// GetUserTool 创建查询用户 Tool
func GetUserTool() Tool {
	return Tool{
		Name:        "get_user",
		Description: "查询用户信息",
		Handler: func(args string) (string, error) {
			var params struct {
				UserID string `json:"user_id"`
			}

			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			return GetUser(params.UserID), nil
		},
	}
}

// QueryPaymentTool 创建查询支付 Tool
func QueryPaymentTool() Tool {
	return Tool{
		Name:        "query_payment",
		Description: "查询订单支付信息",
		Handler: func(args string) (string, error) {
			var params struct {
				OrderID string `json:"order_id"`
			}

			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			return QueryPayment(params.OrderID), nil
		},
	}
}

// QueryLogisticsTool 创建查询物流 Tool
func QueryLogisticsTool() Tool {
	return Tool{
		Name:        "query_logistics",
		Description: "查询订单物流信息",
		Handler: func(args string) (string, error) {
			var params struct {
				OrderID string `json:"order_id"`
			}

			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}

			return QueryLogistics(params.OrderID), nil
		},
	}
}
