package routes

import (
	"byteport/lib"
	"byteport/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func DeployProject(c *gin.Context){
	// add project to db after compiling to obj;
	// get project from request
	// compile project
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	if err := newProject.BeforeSave(models.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save project"})
		return}
	fmt.Println("Deploying project: ", newProject)
	url := "http://localhost:3000/deploy"
	
	jsonProject, err := json.Marshal(newProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert project to json"})}
		
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonProject))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
	}
	accessToken,err := lib.GenerateNVMSToken(newProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// include credentails in cookie
	req.AddCookie(&http.Cookie{
		Name:  "Authorization",
		Value: "Bearer " + accessToken,
	})
	// add project to request body as json decode
	req.Header.Set("Content-Type", "application/json")
	// add new project to request body
	// convert project to json
	// add json to request body
	
	
	


	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deploy project"})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deploy project"})
	}

	if resp.StatusCode != http.StatusOK {
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deploy project"})
	}

	if err := json.Unmarshal(body, &newProject); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse deployed project response"})
		return
	}
	fmt.Println("Deployed project: ", string(body))
	finalProject := models.Project{
		ID:  newProject.UUID, 
		Owner: newProject.User.UUID,
		Name: newProject.Name,
		RepositoryID: newProject.RepositoryID,
		UUID: newProject.UUID,
		Repository: newProject.Repository,
		Readme: newProject.Readme,
		Description: newProject.Description,
		AccessURL: newProject.AccessURL,
		DeploymentsJSON: newProject.DeploymentsJSON,
	}
	if finalProject.ID == "" {
		finalProject.ID = uuid.New().String()  // Generate new UUID only if not present
	}
	finalProject.SetDeploy(newProject.GetDeploy())
	if err := finalProject.BeforeSave(models.DB); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse deployed project response"})
		return
	}
	fmt.Println("Adding project to db: ", finalProject)
	err = addNewProject(finalProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add project to database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
	})

}
func addNewProject(project models.Project)(error){
		fmt.Println("Adding project to db: ", project)
		result := models.DB.Create(&project)
		return result.Error
	 
}