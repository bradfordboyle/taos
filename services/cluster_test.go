package services_test

import (
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"github.com/satori/go.uuid"

	"github.com/kmacoskey/taos/app"
	"github.com/kmacoskey/taos/models"
	. "github.com/kmacoskey/taos/services"
)

var _ = Describe("Cluster", func() {

	var (
		cluster                *models.Cluster
		cluster1UUID           string
		cluster1               *models.Cluster
		cluster2UUID           string
		cluster2               *models.Cluster
		clusters               []models.Cluster
		cs                     *ClusterService
		rc                     app.RequestContext
		err                    error
		validTerraformConfig   []byte
		invalidTerraformConfig []byte
	)

	BeforeEach(func() {
		// Create a new RequestContext for each test
		rc = app.RequestContext{}

		validTerraformConfig = []byte(`{"provider":{"google":{}}}`)
		invalidTerraformConfig = []byte(`notjson`)

		cluster1UUID = "a19e2758-0ec5-11e8-ba89-0ed5f89f718b"
		cluster1 = &models.Cluster{
			Id:              cluster1UUID,
			Name:            "cluster",
			Status:          "status",
			TerraformConfig: []byte(`{"provider":{"google":{}}}`),
		}

		cluster2UUID = "a19e2bfe-0ec5-11e8-ba89-0ed5f89f718b"
		cluster2 = &models.Cluster{
			Id:              cluster2UUID,
			Name:            "cluster",
			Status:          "status",
			TerraformConfig: []byte(`{"provider":{"google":{}}}`),
		}
	})

	// ======================================================================
	//                      _
	//   ___ _ __ ___  __ _| |_ ___
	//  / __| '__/ _ \/ _` | __/ _ \
	// | (__| | |  __/ (_| | ||  __/
	//  \___|_|  \___|\__,_|\__\___|
	//
	// ======================================================================

	Describe("Creating a Valid Cluster", func() {
		Context("A cluster is returned from the dao", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				rc.SetTerraformConfig(validTerraformConfig)
				cluster, err = cs.CreateCluster(rc)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
			It("Should have a cluster returned with status provisioning", func() {
				Expect(cluster.Status).To(Equal("provisioning"))
			})
			It("Should set the cluster status in the daos", func() {
				Expect(cluster.Status).To(Equal("provisioning"))
			})
			It("Should eventually be provisioned", func() {
				Eventually(func() string {
					c, err := cs.GetCluster(rc, cluster.Id)
					Expect(err).NotTo(HaveOccurred())
					return c.Status
				}, 2, 0.5).Should(Equal("provision_success"))
			})
		})

		Context("A cluster is not returned from the dao", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				cluster, err = cs.CreateCluster(rc)
			})
			It("Should return an empty Cluster", func() {
				Expect(cluster).To(Equal(&models.Cluster{}))
			})
			It("should error", func() {
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	Describe("Creating an Invalid Cluster", func() {
		Context("Invalid terraform config is used", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				rc.SetTerraformConfig(invalidTerraformConfig)
				cluster, err = cs.CreateCluster(rc)
			})
			It("Should not return an error when requested", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should eventually change status", func() {
				Eventually(func() string {
					c, err := cs.GetCluster(rc, cluster.Id)
					Expect(err).NotTo(HaveOccurred())
					return c.Status
				}, 2, 0.5).Should(Equal("provision_failed"))
			})
		})
	})

	// ======================================================================
	//             _
	//   __ _  ___| |_
	//  / _` |/ _ \ __|
	// | (_| |  __/ |_
	//  \__, |\___|\__|
	//  |___/
	//
	// ======================================================================

	Describe("Retrieving a Cluster for a specific id", func() {
		Context("A cluster is returned from the dao", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap["a19e2758-0ec5-11e8-ba89-0ed5f89f718b"] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
			})
			It("Should return a cluster of the same id", func() {
				Expect(cs.GetCluster(rc, "a19e2758-0ec5-11e8-ba89-0ed5f89f718b")).To(Equal(cluster1))
			})
		})

		Context("A cluster is not returned from the dao", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				cluster1, err = cs.GetCluster(rc, "a19e2758-0ec5-11e8-ba89-0ed5f89f718b")
			})
			It("Should return an empty Cluster", func() {
				Expect(cluster1).To(Equal(&models.Cluster{}))
			})
			It("should error", func() {
				Expect(err).Should(HaveOccurred())
			})
		})
	})

	// ======================================================================
	//             _
	//   __ _  ___| |_ ___
	//  / _` |/ _ \ __/ __|
	// | (_| |  __/ |_\__ \
	//  \__, |\___|\__|___/
	//  |___/
	//
	// ======================================================================

	Describe("Retrieving all clusters", func() {
		Context("When Clusters are returned from the dao", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
			})
			It("Should return a slice of all clusters", func() {
				Expect(cs.GetClusters(rc)).To(HaveLen(2))
			})
		})

		Context("When no Clusters are returned from the dao", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				clusters, err = cs.GetClusters(rc)
			})
			It("Should return an empty list of Clusters", func() {
				Expect(clusters).To(HaveLen(0))
			})
			It("should not error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})

		})
	})

	// ======================================================================
	//      _      _      _
	//   __| | ___| | ___| |_ ___
	//  / _` |/ _ \ |/ _ \ __/ _ \
	// | (_| |  __/ |  __/ ||  __/
	//  \__,_|\___|_|\___|\__\___|
	//
	// ======================================================================

	Describe("Deleting a Cluster", func() {
		Context("That exists", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap[cluster1UUID] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				cluster, err = cs.DeleteCluster(rc, cluster1UUID)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return the same cluster", func() {
				Expect(cluster.Id).To(Equal(cluster1.Id))
			})
			It("The returned cluster should have a deleting status", func() {
				Expect(cluster.Status).To(Equal("destroying"))
			})
			It("Should eventually be deleted", func() {
				Eventually(func() string {
					c, err := cs.GetCluster(rc, cluster.Id)
					Expect(err).NotTo(HaveOccurred())
					return c.Status
				}, 2, 0.5).Should(Equal("destroyed"))
			})
		})

		Context("That doesn't exist", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				cluster, err = cs.DeleteCluster(rc, cluster1UUID)
			})
			It("should error", func() {
				Expect(err).Should(HaveOccurred())
			})
			It("Should return a nil cluster", func() {
				Expect(cluster).To(BeNil())
			})
		})

		Context("That has already been deleted", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cluster1.Status = "destroyed"
				clustersMap[cluster1UUID] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				cluster, err = cs.DeleteCluster(rc, cluster1UUID)
			})
			It("Should error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should not change status of the original cluster", func() {
				Expect(cluster.Status).To(Equal(cluster1.Status))
			})
			It("Should return the original cluster", func() {
				Expect(cluster.Id).To(Equal(cluster1.Id))
			})
		})
	})

})

func NewMockDB() *MockDB {
	return &MockDB{}
}

type MockDB struct {
	db *sqlx.DB
}

type ValidClusterDao struct {
	clustersMap map[string]*models.Cluster
}

func NewValidClusterDao(cm map[string]*models.Cluster) *ValidClusterDao {
	return &ValidClusterDao{
		clustersMap: cm,
	}
}

func (dao *ValidClusterDao) CreateCluster(db *sqlx.DB, config []byte) (*models.Cluster, error) {
	uuid := uuid.Must(uuid.NewV4()).String()
	dao.clustersMap[uuid] = &models.Cluster{
		Id:              uuid,
		Name:            "cluster",
		Status:          "status",
		TerraformConfig: config,
	}
	return dao.clustersMap[uuid], nil
}

func (dao *ValidClusterDao) UpdateCluster(db *sqlx.DB, cluster *models.Cluster) (*models.Cluster, error) {
	dao.clustersMap[cluster.Id] = cluster
	return dao.clustersMap[cluster.Id], nil
}

func (dao *ValidClusterDao) GetCluster(db *sqlx.DB, id string) (*models.Cluster, error) {
	return dao.clustersMap[id], nil
}

func (dao *ValidClusterDao) GetClusters(db *sqlx.DB) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	cluster := models.Cluster{}
	clusters = append(clusters, cluster)
	clusters = append(clusters, cluster)
	return clusters, nil
}

func (dao *ValidClusterDao) DeleteCluster(db *sqlx.DB, id string) (*models.Cluster, error) {
	if _, ok := dao.clustersMap[id]; !ok {
		return nil, errors.New("foo")
	} else {
		dao.clustersMap[id].Status = "destroying"
		return dao.clustersMap[id], nil
	}
}

type EmptyClusterDao struct {
	clustersMap map[string]*models.Cluster
}

func NewEmptyClusterDao() *EmptyClusterDao {
	return &EmptyClusterDao{}
}

func (dao *EmptyClusterDao) CreateCluster(db *sqlx.DB, config []byte) (*models.Cluster, error) {
	return &models.Cluster{}, errors.New("foo")
}

func (dao *EmptyClusterDao) UpdateCluster(db *sqlx.DB, cluster *models.Cluster) (*models.Cluster, error) {
	return cluster, nil
}

func (dao *EmptyClusterDao) GetCluster(db *sqlx.DB, id string) (*models.Cluster, error) {
	return &models.Cluster{}, errors.New("foo")
}

func (dao *EmptyClusterDao) GetClusters(db *sqlx.DB) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	return clusters, nil
}

func (dao *EmptyClusterDao) DeleteCluster(db *sqlx.DB, id string) (*models.Cluster, error) {
	return nil, errors.New("foo")
}
