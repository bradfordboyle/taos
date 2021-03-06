package services_test

import (
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"errors"

	"github.com/satori/go.uuid"

	"github.com/kmacoskey/taos/app"
	"github.com/kmacoskey/taos/models"
	. "github.com/kmacoskey/taos/services"
	"github.com/kmacoskey/taos/terraform"
)

var (
	validTerraformOutputs = "{\"bar\":{\"sensitive\":false,\"type\":\"string\",\"value\":\"foo\" }"
	validTerraformState   = []byte(`{"version":3,"terraform_version":"0.11.3","serial":2,"lineage":"26655d4c-852a-41e4-b6f1-7b31ff2b2981","modules":[{"path":["root"],"outputs":{"foo":{"sensitive":false,"type":"string","value":"bar"}},"resources":{},"depends_on":[]}]}`)
)

var _ = Describe("Cluster", func() {

	var (
		cluster                       *models.Cluster
		cluster1UUID                  string
		cluster1                      *models.Cluster
		cluster2UUID                  string
		cluster2                      *models.Cluster
		clusters                      []models.Cluster
		validRequestId                string
		cs                            *ClusterService
		rc                            app.RequestContext
		err                           error
		validTerraformConfig          []byte
		validTimeout                  string
		invalidTerraformConfig        []byte
		validNoOutputsTerraformConfig []byte
		validProject                  string
		validRegion                   string
		terraformClient               TerraformClient
	)

	BeforeEach(func() {
		log.SetLevel(log.FatalLevel)

		// Create a new RequestContext for each test
		rc = app.RequestContext{}

		validProject = "valid-project-name"
		validRegion = "valid-region"
		validTimeout = "10m"
		validTerraformConfig = []byte(`{"provider":{"google":{"project":"data-gp-toolsmiths","region":"us-central1"}},"output":{"foo":{"value":"bar"}}}`)
		validNoOutputsTerraformConfig = []byte(`{"provider":{"google":{"project":"data-gp-toolsmiths","region":"us-central1"}}}`)
		invalidTerraformConfig = []byte(`notjson`)

		cluster1UUID = "a19e2758-0ec5-11e8-ba89-0ed5f89f718b"
		cluster1 = &models.Cluster{
			Id:              cluster1UUID,
			Name:            "cluster",
			Status:          "status",
			TerraformConfig: []byte(`{"provider":{"google":{}}}`),
			Project:         validProject,
			Region:          validRegion,
		}

		cluster2UUID = "a19e2bfe-0ec5-11e8-ba89-0ed5f89f718b"
		cluster2 = &models.Cluster{
			Id:              cluster2UUID,
			Name:            "cluster",
			Status:          "status",
			TerraformConfig: []byte(`{"provider":{"google":{}}}`),
			Project:         validProject,
			Region:          validRegion,
		}

		validRequestId = "ff459ef4-514b-11e8-9c2d-fa7ae01bbebc"
	})

	// ======================================================================
	//                      _
	//   ___ _ __ ___  __ _| |_ ___
	//  / __| '__/ _ \/ _` | __/ _ \
	// | (__| | |  __/ (_| | ||  __/
	//  \___|_|  \___|\__,_|\__\___|
	//
	// ======================================================================

	Describe("Creating a cluster", func() {
		Context("When everything goes ok", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				terraformClient = new(PassingClient)
				cluster, err = cs.CreateCluster(validTerraformConfig, validTimeout, validProject, validRegion, validRequestId, terraformClient)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
			It("Should set the project of the terraform client", func() {
				Expect(terraformClient.Project()).To(Equal(validProject))
			})
			It("Should set the region of the terraform client", func() {
				Expect(terraformClient.Region()).To(Equal(validRegion))
			})
		})

		Context("When a cluster is not returned from the dao", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				client := new(FailingClient)
				cluster, err = cs.CreateCluster(validTerraformConfig, validTimeout, validRequestId, validProject, validRegion, client)
			})
			It("Should error", func() {
				Expect(err).Should(HaveOccurred())
			})
			It("Should not return a cluster", func() {
				Expect(cluster).To(BeNil())
			})
		})

		Context("When invalid terraform config is used", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				client := new(PassingClient)
				cluster, err = cs.CreateCluster(invalidTerraformConfig, validTimeout, validRequestId, validProject, validRegion, client)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
		})

		Context("When there are no outputs defined in the config", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				client := new(FailingClient)
				cluster, err = cs.CreateCluster(validNoOutputsTerraformConfig, validTimeout, validRequestId, validProject, validRegion, client)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
		})

	})

	Describe("Terraform Provisioning a cluster", func() {
		Context("When everything goes ok", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap[cluster1.Id] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				rc.SetTerraformConfig(validTerraformConfig)
				client := new(PassingClient)
				cluster = cs.TerraformProvisionCluster(client, cluster1, validTerraformConfig, cluster1UUID)
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
			It("Should set the cluster status as expected", func() {
				Expect(cluster.Status).To(Equal(models.ClusterStatusProvisionSuccess))
			})
			It("Should set the cluster message as expected", func() {
				Expect(cluster.Message).To(Equal(terraform.ApplySuccess))
			})
			It("Should set the cluster state as expected", func() {
				Expect(cluster.TerraformState).To(Equal(validTerraformState))
			})
			It("Should set the cluster outputs as expected", func() {
				Expect(cluster.Outputs).To(Equal([]byte(validTerraformOutputs)))
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

	Describe("Getting a cluster", func() {
		Context("When everything goes ok", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap[cluster1.Id] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				cluster, err = cs.GetCluster(validRequestId, cluster1.Id)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return a cluster", func() {
				Expect(cluster).NotTo(BeNil())
			})
			It("Should return the expected cluster", func() {
				Expect(cluster.Id).To(Equal(cluster1.Id))
			})
		})

		Context("When the cluster does not exist", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				cluster, err = cs.GetCluster(validRequestId, cluster1.Id)
			})
			It("Should error", func() {
				Expect(err).Should(HaveOccurred())
			})
			It("Should not return a cluster", func() {
				Expect(cluster).To(BeNil())
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

	Describe("Getting all clusters", func() {
		Context("When everything goes ok", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap[cluster1.Id] = cluster1
				clustersMap[cluster2.Id] = cluster2
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				clusters, err = cs.GetClusters(validRequestId)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should return clusters", func() {
				Expect(clusters).To(HaveLen(2))
			})
			It("Should return the expected clusters", func() {
				Expect(clusters).To(ContainElement(*cluster1))
				Expect(clusters).To(ContainElement(*cluster2))
			})
		})

		Context("When there are no clusters", func() {
			BeforeEach(func() {
				cs = NewClusterService(NewEmptyClusterDao(), NewMockDB().db)
				clusters, err = cs.GetClusters(validRequestId)
			})
			It("should not error", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("Should return an empty list of Clusters", func() {
				Expect(clusters).To(HaveLen(0))
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

	Describe("Deleting a cluster", func() {

		Context("When everything goes ok", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				clustersMap[cluster1UUID] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				terraformClient = new(PassingClient)
				cluster, err = cs.DeleteCluster(validRequestId, terraformClient, cluster1UUID)
			})
			It("Should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			It("Should set the project of the terraform client", func() {
				Expect(terraformClient.Project()).To(Equal(validProject))
			})
			It("Should set the region of the terraform client", func() {
				Expect(terraformClient.Region()).To(Equal(validRegion))
			})
			It("Should return the expected cluster", func() {
				Expect(cluster.Id).To(Equal(cluster1.Id))
			})
			It("The should be destroying", func() {
				Expect(cluster.Status).To(Equal("destroying"))
			})
		})

		Context("When it does not exist", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				client := new(PassingClient)
				cluster, err = cs.DeleteCluster(validRequestId, client, cluster1.Id)
			})
			It("should error", func() {
				Expect(err).Should(HaveOccurred())
			})
			It("Should not return a cluster", func() {
				Expect(cluster).To(BeNil())
			})
		})

		Context("That has already been deleted", func() {
			BeforeEach(func() {
				clustersMap := make(map[string]*models.Cluster)
				cluster1.Status = "destroyed"
				clustersMap[cluster1.Id] = cluster1
				cs = NewClusterService(NewValidClusterDao(clustersMap), NewMockDB().db)
				client := new(PassingClient)
				cluster, err = cs.DeleteCluster(validRequestId, client, cluster1.Id)
			})
			It("Should error", func() {
				Expect(err).To(HaveOccurred())
			})
			It("Should not return a cluster", func() {
				Expect(cluster).To(BeNil())
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

type PassingClient struct {
	Terraform   terraform.TerraformInfra
	project     string
	region      string
	credentials string
}

func (client *PassingClient) ClientInit() error                 { return nil }
func (client *PassingClient) ClientDestroy() error              { return nil }
func (client *PassingClient) Config() []byte                    { return []byte(`json`) }
func (client *PassingClient) SetConfig(config []byte)           { return }
func (client *PassingClient) State() []byte                     { return []byte(`json`) }
func (client *PassingClient) SetState(state []byte)             { return }
func (client *PassingClient) Project() string                   { return client.project }
func (client *PassingClient) SetProject(project string)         { client.project = project }
func (client *PassingClient) Region() string                    { return client.region }
func (client *PassingClient) SetRegion(region string)           { client.region = region }
func (client *PassingClient) Credentials() string               { return client.credentials }
func (client *PassingClient) SetCredentials(credentials string) { client.credentials = credentials }
func (client *PassingClient) Init() (string, error)             { return "foo", nil }
func (client *PassingClient) Plan(destroy bool) (string, error) { return "foo", nil }
func (client *PassingClient) Outputs() (string, error)          { return validTerraformOutputs, nil }
func (client *PassingClient) Apply() ([]byte, string, error) {
	return validTerraformState, terraform.ApplySuccess, nil
}
func (client *PassingClient) Destroy() ([]byte, string, error) { return []byte(`json`), "foo", nil }

type FailingClient struct{}

func (client *FailingClient) ClientInit() error                 { return errors.New("foo") }
func (client *FailingClient) ClientDestroy() error              { return errors.New("foo") }
func (client *FailingClient) Config() []byte                    { return []byte(`json`) }
func (client *FailingClient) SetConfig(config []byte)           { return }
func (client *FailingClient) State() []byte                     { return []byte(`json`) }
func (client *FailingClient) SetState(state []byte)             { return }
func (client *FailingClient) Init() (string, error)             { return "foo", errors.New("foo") }
func (client *FailingClient) Plan(destroy bool) (string, error) { return "foo", errors.New("foo") }
func (client *FailingClient) Outputs() (string, error)          { return "", errors.New("foo") }
func (client *FailingClient) Apply() ([]byte, string, error)    { return nil, "", errors.New("") }
func (client *FailingClient) Destroy() ([]byte, string, error)  { return nil, "", errors.New("foo") }
func (client *FailingClient) Project() string                   { return "" }
func (client *FailingClient) SetProject(project string)         { return }
func (client *FailingClient) Region() string                    { return "" }
func (client *FailingClient) SetRegion(region string)           { return }
func (client *FailingClient) Credentials() string               { return "" }
func (client *FailingClient) SetCredentials(credentials string) { return }

type ValidClusterDao struct {
	clustersMap map[string]*models.Cluster
}

func NewValidClusterDao(cm map[string]*models.Cluster) *ValidClusterDao {
	return &ValidClusterDao{
		clustersMap: cm,
	}
}

func (dao *ValidClusterDao) CreateCluster(db *sqlx.DB, config []byte, timeout string, requestId string, project string, region string) (*models.Cluster, error) {
	uuid := uuid.Must(uuid.NewV4()).String()
	dao.clustersMap[uuid] = &models.Cluster{
		Id:              uuid,
		Name:            "cluster",
		Status:          "status",
		TerraformConfig: config,
	}
	return dao.clustersMap[uuid], nil
}

func (dao *ValidClusterDao) UpdateClusterField(db *sqlx.DB, id string, field string, value interface{}, requestId string) error {
	cluster := &models.Cluster{}
	cluster = dao.clustersMap[id]
	switch field {
	case "status":
		cluster.Status = value.(string)
	case "message":
		cluster.Message = value.(string)
	case "outputs":
		cluster.Outputs = value.([]byte)
	case "terraform_config":
		cluster.TerraformConfig = value.([]byte)
	case "terraform_state":
		cluster.TerraformState = value.([]byte)
	}
	dao.clustersMap[id] = cluster
	return nil
}

func (dao *ValidClusterDao) GetCluster(db *sqlx.DB, id string, requestId string) (*models.Cluster, error) {
	return dao.clustersMap[id], nil
}

func (dao *ValidClusterDao) GetClusters(db *sqlx.DB, requestId string) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	for _, cluster := range dao.clustersMap {
		clusters = append(clusters, *cluster)
	}
	return clusters, nil
}

func (dao *ValidClusterDao) GetExpiredClusters(db *sqlx.DB, requestId string) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	for _, cluster := range dao.clustersMap {
		clusters = append(clusters, *cluster)
	}
	return clusters, nil
}

func (dao *ValidClusterDao) DeleteCluster(db *sqlx.DB, id string, requestId string) (*models.Cluster, error) {
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

func (dao *EmptyClusterDao) CreateCluster(db *sqlx.DB, config []byte, timeout string, requestId string, project string, region string) (*models.Cluster, error) {
	return nil, errors.New("foo")
}

func (dao *EmptyClusterDao) UpdateClusterField(db *sqlx.DB, id string, field string, value interface{}, requestId string) error {
	return nil
}

func (dao *EmptyClusterDao) GetCluster(db *sqlx.DB, id string, requestId string) (*models.Cluster, error) {
	return nil, errors.New("foo")
}

func (dao *EmptyClusterDao) GetClusters(db *sqlx.DB, requestId string) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	return clusters, nil
}

func (dao *EmptyClusterDao) GetExpiredClusters(db *sqlx.DB, requestId string) ([]models.Cluster, error) {
	clusters := []models.Cluster{}
	return clusters, nil
}

func (dao *EmptyClusterDao) DeleteCluster(db *sqlx.DB, id string, requestId string) (*models.Cluster, error) {
	return nil, errors.New("foo")
}
