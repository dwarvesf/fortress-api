package view

import "github.com/dwarvesf/fortress-api/pkg/model"

type CheckWithdrawConditionResponse struct {
	Data CheckWithdrawCondition `json:"data"`
} // @name CheckWithdrawConditionResponse

type CheckWithdrawCondition struct {
	ICYAmount  float64 `json:"icyAmount"`
	ICYVNDRate float64 `json:"icyVNDRate"`
	VNDAmount  float64 `json:"vndAmount"`
} // @name CheckWithdrawCondition

func ToCheckWithdrawCondition(in *model.WithdrawalCondition) CheckWithdrawCondition {
	return CheckWithdrawCondition{
		ICYAmount:  in.ICYAmount,
		ICYVNDRate: in.ICYVNDRate,
		VNDAmount:  in.VNDAmount,
	}
}
