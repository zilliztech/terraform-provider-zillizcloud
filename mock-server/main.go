package main

import (
	"net/http"

	"example.com/m/pkg/byoc_project"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a default gin router
	r := gin.Default()

	// Define the API endpoint
	v2 := r.Group("/v2")
	{
		v2.GET("/projects", byoc_project.GetProjects)
		clusters := v2.Group("/clusters")
		{
			clusters.POST("/createDedicated", byoc_project.CreateDedicatedCluster)
			clusters.POST("/createServerless", byoc_project.CreateServerlessCluster)
			clusters.POST("/createFree", byoc_project.CreateFreeCluster)
			clusters.GET("/:clusterId", byoc_project.GetCluster)
			clusters.POST("/:clusterId/resume", byoc_project.ResumeCluster)
			clusters.POST("/:clusterId/suspend", byoc_project.SuspendCluster)
			clusters.POST("/:clusterId/modifyReplica", byoc_project.ModifyClusterReplica)
			clusters.POST("/:clusterId/modify", byoc_project.ModifyCluster)
			clusters.GET("/:clusterId/labels", byoc_project.GetLabels)
			clusters.PUT("/:clusterId/labels", byoc_project.UpdateLabels)
			clusters.DELETE("/:clusterId/drop", byoc_project.DropCluster)
		}
		byoc := v2.Group("/byoc")
		{
			dataplane := byoc.Group("/dataplane")
			{
				dataplane.POST("/create", byoc_project.CreateDataplane)
				dataplane.GET("/describe", byoc_project.DescribeDataplane)
				dataplane.DELETE("/delete", byoc_project.DeleteDataplane)
				dataplane.POST("/stop", byoc_project.SuspendDataplane)
				dataplane.POST("/resume", byoc_project.ResumeDataplane)
			}
			// external id
			byoc.GET("/describe", describeByoc)
			op := byoc.Group("/op")
			{

				dataplane := op.Group("/dataplane")

				{
					dataplane.POST("/setting", byoc_project.CreateSettings)
					dataplane.GET("/setting", byoc_project.GetSettings)
					dataplane.POST("/create", byoc_project.CreateOpDataplane)
					// dataplane.GET("/describe", byoc_project.DescribeDataplane)
					// dataplane.DELETE("/delete", byoc_op_project.DeleteDataplane)
				}
			}
		}
	}

	// Start the server
	r.Run(":8080")
}
func describeByoc(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"orgId":          "org-dhwkctdqnwdknwradunwje",
			"externalId":     "cid-c88368a7164f15ad9e1fa9068",
			"serviceAccount": "Not need",
			"clouds":         []string{"aws"},
		},
	})
}
