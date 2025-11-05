package byoc_project

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateProject creates a new standard project
func CreateProject(c *gin.Context) {
	var request CreateProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	// Validate required fields
	if request.ProjectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "projectName is required",
		})
		return
	}

	if request.Plan == "" {
		request.Plan = "Enterprise" // Default plan
	}

	// Generate new project ID
	projectId := fmt.Sprintf("proj-%s", uuid.New().String()[:22])

	// Create project
	project := &Project{
		ProjectName:     request.ProjectName,
		ProjectId:       projectId,
		InstanceCount:   0,
		CreateTimeMilli: time.Now().UnixMilli(),
		Plan:            request.Plan,
	}

	// Store project
	projectStore.Set(projectId, project)

	// Return project ID
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": projectId,
	})
}

// ListProjects returns all projects
func ListProjects(c *gin.Context) {
	projects := projectStore.GetAll()

	projectList := make([]Project, 0, len(projects))
	for _, p := range projects {
		projectList = append(projectList, *p)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": projectList,
	})
}

// GetProjectById returns a specific project by ID
func GetProjectById(c *gin.Context) {
	projectId := c.Param("projectId")

	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "projectId is required",
		})
		return
	}

	project := projectStore.Get(projectId)
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": fmt.Sprintf("Project with ID %s not found", projectId),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": *project,
	})
}

// UpgradeProjectPlan upgrades the plan for a project
func UpgradeProjectPlan(c *gin.Context) {
	projectId := c.Param("projectId")

	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "projectId is required",
		})
		return
	}

	var request UpgradeProjectPlanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	if request.Plan == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "plan is required",
		})
		return
	}

	project := projectStore.Get(projectId)
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": fmt.Sprintf("Project with ID %s not found", projectId),
		})
		return
	}

	// Update the plan
	project.Plan = request.Plan
	projectStore.Set(projectId, project)

	// Return success message
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": fmt.Sprintf("Project %s plan upgraded to %s", projectId, request.Plan),
	})
}

// DeleteProject deletes a project (mock - just returns success)
func DeleteProject(c *gin.Context) {
	projectId := c.Param("projectId")

	if projectId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "projectId is required",
		})
		return
	}

	project := projectStore.Get(projectId)
	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": fmt.Sprintf("Project with ID %s not found", projectId),
		})
		return
	}

	// Delete from store
	projectStore.Delete(projectId)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": fmt.Sprintf("Project %s deleted successfully", projectId),
	})
}
