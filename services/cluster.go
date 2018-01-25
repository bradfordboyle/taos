package services

import (
	"github.com/kmacoskey/taos/app"
	"github.com/kmacoskey/taos/daos"
	"github.com/kmacoskey/taos/models"
)

func GetCluster(rc app.RequestContext, id int) (*models.Cluster, error) {
	cluster, err := daos.GetCluster(rc, id)
	return &cluster, err
}
