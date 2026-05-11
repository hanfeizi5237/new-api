package model

import "gorm.io/gorm"

type MarketplaceOperationAudit struct {
	Id          int    `json:"id"`
	ActorUserId int    `json:"actor_user_id" gorm:"not null;default:0;index:idx_marketplace_operation_audits_actor_user_id"`
	ActorType   string `json:"actor_type" gorm:"type:varchar(32);not null;default:'system';index:idx_marketplace_operation_audits_actor_type"`
	Action      string `json:"action" gorm:"type:varchar(64);not null;index:idx_marketplace_operation_audits_action"`
	TargetType  string `json:"target_type" gorm:"type:varchar(32);not null;index:idx_marketplace_operation_audits_target,priority:1"`
	TargetId    int    `json:"target_id" gorm:"not null;default:0;index:idx_marketplace_operation_audits_target,priority:2"`
	RequestId   string `json:"request_id" gorm:"type:varchar(64);index:idx_marketplace_operation_audits_request_id"`
	Ip          string `json:"ip" gorm:"type:varchar(64)"`
	Reason      string `json:"reason" gorm:"type:text"`
	Result      string `json:"result" gorm:"type:varchar(32);not null;default:'success';index:idx_marketplace_operation_audits_result"`
	BeforeState string `json:"before_state" gorm:"type:text"`
	AfterState  string `json:"after_state" gorm:"type:text"`
	Meta        string `json:"meta" gorm:"type:text"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index:idx_marketplace_operation_audits_created_at"`
}

func (a *MarketplaceOperationAudit) BeforeCreate(tx *gorm.DB) error {
	setMarketplaceTimestampsOnCreate(&a.CreatedAt, nil)
	return nil
}

func CreateMarketplaceOperationAuditTx(tx *gorm.DB, audit *MarketplaceOperationAudit) error {
	return tx.Create(audit).Error
}
