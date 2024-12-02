package routes

import (
	"byteport/lib"
	"byteport/models"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeployProject(c *gin.Context){
	// add project to db after compiling to obj;
	// get project from request
	// compile project
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	}
	models.DB.Create(&newProject)
	fmt.Println("Deploying project: ", newProject)
	url := fmt.Sprintf("http://localhost:3000/")
	req, err := http.NewRequest("GET", url, nil)
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

	req.Body


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
	fmt.Println("Deployed project: ", string(body))
	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"project": newProject,
	})

}