package model

type ClusterRole string

const (
	ClusterRoleOperator ClusterRole = "operator"
	ClusterRoleViewer   ClusterRole = "viewer"
)
