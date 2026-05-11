package models

import "github.com/google/uuid"

// AssignNewUUIDIfZero 在 id 当前为 uuid.Nil 时填入新 UUID，用于 GORM BeforeCreate
// 钩子内的 ID 默认值生成 —— 跨服务 16 个 model 此前各自复制了同一段 if-nil-assign
// 模板，集中到此 helper 后调用方一行收尾。
//
// 不使用嵌入式 base struct 路径的理由：GORM v2 嵌入字段的 method 提升语义
// 受 struct 字段名遮蔽 + 多重嵌入冲突影响，行为不直观，且 16 个 model 已有
// 各自的 ID 字段定义和 gorm tag。helper 路径风险最小，仍把业务逻辑收敛到一处。
func AssignNewUUIDIfZero(id *uuid.UUID) {
	if id == nil {
		return
	}
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
