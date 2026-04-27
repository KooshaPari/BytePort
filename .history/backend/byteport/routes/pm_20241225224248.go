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


func addNewProject(project models.Project)(error){
		fmt.Println("Adding project to db: ", project)
		result := models.DB.Create(&project)
		return result.Error
	 
}
func removeProject(project models.Project)(error){
		fmt.Println("Removing Project From DB: ", project)
		result := models.DB.Create(&project)
		return result.Error
	 
}